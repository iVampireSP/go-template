// Package main implements autodi, a compile-time dependency injection code generator.
//
// autodi scans Go packages for exported New* constructor functions, builds a
// dependency graph via type analysis, performs topological sorting with cycle
// detection, and generates a complete main.go with two-phase DI — replacing
// runtime DI frameworks like uber/fx with zero-reflection, compile-time safe code.
//
// Generation flow:
//
//  1. Read go.mod → module path
//  2. Read generate.go → //autodi:app/embed/group annotations
//  3. Scan internal/ + pkg/ → provider candidates (New* constructors)
//  4. Scan cmd/ → discover commands (entry points)
//  5. Filter candidates to reachable providers (BFS from command params)
//  6. Build dependency graph + resolve bindings + detect Close/Shutdown/Stop
//  7. For each DI command:
//     Analyze New* params → trace transitive deps → generate init function
//  8. Generate main.go with two-phase DI
//
// Usage:
//
//	//go:generate go run github.com/iVampireSP/autodi@latest
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	verbose := flag.Bool("verbose", false, "enable verbose logging")
	dryRun := flag.Bool("dry-run", false, "print generated code without writing")
	flag.Parse()

	// Resolve module root: walk up from cwd to find go.mod
	moduleRoot, err := findModuleRoot()
	if err != nil {
		log.Fatalf("autodi: %v", err)
	}

	// Build config from conventions (go.mod + generate.go)
	cfg, err := BuildConfig(moduleRoot)
	if err != nil {
		log.Fatalf("autodi: %v", err)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "autodi: module=%s root=%s\n", cfg.Module, moduleRoot)
		fmt.Fprintf(os.Stderr, "autodi: app=%s\n", cfg.AppName)
	}

	totalStart := time.Now()

	// Load gitignore patterns
	gitignorePatterns := LoadGitignore(moduleRoot)

	// ── Pass 1: Scan provider candidates ──

	t0 := time.Now()
	scanner := NewScanner(cfg, moduleRoot, gitignorePatterns)
	candidates, err := scanner.Scan()
	if err != nil {
		log.Fatalf("autodi: scan: %v", err)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "autodi: [%s] scan: discovered %d candidates\n", time.Since(t0), len(candidates))
	}

	// ── Pass 2: Discover commands from cmd/ packages ──

	t1 := time.Now()
	detector := NewCommandDetector(cfg, moduleRoot)
	commands, err := detector.Detect()
	if err != nil {
		log.Fatalf("autodi: detect commands: %v", err)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "autodi: [%s] detect: discovered %d commands\n", time.Since(t1), len(commands))
		for _, cmd := range commands {
			var paramTypes []string
			for _, p := range cmd.Params {
				paramTypes = append(paramTypes, toShortTypeName(p.TypeStr))
			}
			kind := "multi"
			if cmd.IsSingle {
				kind = "single"
			}
			if !cmd.HasDeps() {
				kind += "/zero-dep"
			}
			var handlers []string
			for _, h := range cmd.Handlers {
				handlers = append(handlers, h.MethodName)
			}
			fmt.Fprintf(os.Stderr, "  [%s] %s: %s.%s(%s) → [%s]\n",
				kind, cmd.Name, cmd.StructName, cmd.FuncName,
				strings.Join(paramTypes, ", "), strings.Join(handlers, ", "))
		}
	}

	// ── Pass 3: Filter to reachable providers only ──

	t2 := time.Now()
	providers := FilterReachable(candidates, commands, cfg, scanner.IfaceTypes, *verbose)

	if *verbose {
		fmt.Fprintf(os.Stderr, "autodi: [%s] reachable: %d candidates → %d providers\n",
			time.Since(t2), len(candidates), len(providers))
	}

	// ── Pass 4: Build dependency graph ──

	t3 := time.Now()
	graph, errs := BuildGraph(providers, cfg, scanner.PkgIndex, scanner.IfaceTypes)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "autodi: %v\n", e)
		}
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "autodi: [%s] build graph\n", time.Since(t3))
	}

	t4 := time.Now()
	if errs := graph.VerifyAcyclic(); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "autodi: %v\n", e)
		}
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "autodi: [%s] verify acyclic\n", time.Since(t4))
	}

	// Resolve interface bindings for command parameters
	t5 := time.Now()
	graph.BindCommandInterfaces(commands)

	if *verbose {
		fmt.Fprintf(os.Stderr, "autodi: [%s] bind command interfaces\n", time.Since(t5))
	}

	// Validate per-command dependencies
	t6 := time.Now()
	hasValidationErr := false
	for _, cmd := range commands {
		if !cmd.HasDeps() {
			continue
		}
		var neededTypes []string
		for _, param := range cmd.Params {
			neededTypes = append(neededTypes, param.TypeStr)
		}
		pp, err := graph.ProvidersForTypes(neededTypes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "autodi: command %s: %v\n", cmd.Name, err)
			hasValidationErr = true
			continue
		}
		if errs := graph.ValidateEntry(cmd.Name, pp); len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "autodi: %v\n", e)
			}
			hasValidationErr = true
		}
		if *verbose {
			fmt.Fprintf(os.Stderr, "autodi: command %s: %d providers\n", cmd.Name, len(pp))
		}
	}
	if hasValidationErr {
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "autodi: [%s] validate commands\n", time.Since(t6))
	}

	// ── Generate code ──

	t7 := time.Now()
	gen := NewCodeGen(cfg, graph, commands, moduleRoot)
	files, err := gen.Generate()
	if err != nil {
		log.Fatalf("autodi: generate: %v", err)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "autodi: [%s] generate code\n", time.Since(t7))
	}

	// Write or print generated files
	t8 := time.Now()
	for _, f := range files {
		if *dryRun {
			fmt.Fprintf(os.Stdout, "// === %s ===\n%s\n", f.Name, f.Content)
			continue
		}
		path := filepath.Join(moduleRoot, f.Name)
		if *verbose {
			fmt.Fprintf(os.Stderr, "autodi: writing %s\n", path)
		}
		if err := os.WriteFile(path, f.Content, 0644); err != nil {
			log.Fatalf("autodi: write %s: %v", path, err)
		}
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "autodi: [%s] write files\n", time.Since(t8))
	}

	if !*dryRun {
		fmt.Fprintf(os.Stderr, "autodi: generated %d files in %s\n", len(files), time.Since(totalStart))
	}
}

// findModuleRoot walks up from cwd to find the directory containing go.mod.
func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found in any parent directory")
}

func joinStrings(ss []string, sep string) string {
	return strings.Join(ss, sep)
}
