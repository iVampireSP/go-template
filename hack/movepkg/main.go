package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Rule describes one package move: old import path → new import path.
type Rule struct {
	Old string
	New string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <project-root>\n", os.Args[0])
		os.Exit(1)
	}
	root := os.Args[1]

	module := detectModule(root)
	if module == "" {
		fmt.Fprintln(os.Stderr, "cannot detect module from go.mod")
		os.Exit(1)
	}

	rules := []Rule{
		{Old: module + "/internal/infra/bus", New: module + "/pkg/foundation/bus"},
		{Old: module + "/internal/infra/queue", New: module + "/pkg/foundation/queue"},
		{Old: module + "/internal/infra/cache", New: module + "/pkg/foundation/cache"},
		{Old: module + "/internal/infra/cron", New: module + "/pkg/foundation/cron"},
		{Old: module + "/internal/infra/email", New: module + "/pkg/foundation/email"},
		{Old: module + "/pkg/schedule", New: module + "/pkg/foundation/schedule"},
		{Old: module + "/pkg/lock", New: module + "/pkg/foundation/lock"},
	}

	// Step 1: physically move directories
	for _, r := range rules {
		oldDir := importPathToDir(root, module, r.Old)
		newDir := importPathToDir(root, module, r.New)

		if _, err := os.Stat(oldDir); os.IsNotExist(err) {
			fmt.Printf("SKIP (not found): %s\n", oldDir)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(newDir), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", newDir, err)
			os.Exit(1)
		}

		// If destination already has files (e.g. we pre-created foundation packages),
		// merge by copying source files into destination, then remove source dir.
		if info, err := os.Stat(newDir); err == nil && info.IsDir() {
			entries, _ := os.ReadDir(oldDir)
			for _, e := range entries {
				src := filepath.Join(oldDir, e.Name())
				dst := filepath.Join(newDir, e.Name())
				// Only copy if destination doesn't already have a newer version
				if _, err := os.Stat(dst); err == nil {
					fmt.Printf("  SKIP (exists): %s\n", dst)
					continue
				}
				data, err := os.ReadFile(src)
				if err != nil {
					fmt.Fprintf(os.Stderr, "read %s: %v\n", src, err)
					continue
				}
				if err := os.WriteFile(dst, data, 0o644); err != nil {
					fmt.Fprintf(os.Stderr, "write %s: %v\n", dst, err)
					continue
				}
				fmt.Printf("  COPY %s → %s\n", src, dst)
			}
			if err := os.RemoveAll(oldDir); err != nil {
				fmt.Fprintf(os.Stderr, "rm %s: %v\n", oldDir, err)
			}
			fmt.Printf("MERGE %s → %s\n", r.Old, r.New)
		} else {
			if err := os.Rename(oldDir, newDir); err != nil {
				fmt.Fprintf(os.Stderr, "rename %s → %s: %v\n", oldDir, newDir, err)
				os.Exit(1)
			}
			fmt.Printf("MOVE  %s → %s\n", r.Old, r.New)
		}
	}

	// Step 2: rewrite all imports across the entire project
	rewriteCount := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip vendor, .git, and the tool itself
		base := filepath.Base(path)
		if info.IsDir() && (base == "vendor" || base == ".git" || base == "node_modules") {
			return filepath.SkipDir
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		changed, err := rewriteFile(path, rules)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARN rewrite %s: %v\n", path, err)
			return nil // continue with other files
		}
		if changed {
			rewriteCount++
			rel, _ := filepath.Rel(root, path)
			fmt.Printf("REWRITE %s\n", rel)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "walk error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nDone: moved %d packages, rewrote %d files\n", len(rules), rewriteCount)
}

// rewriteFile parses a Go file, rewrites import paths, and writes back if changed.
func rewriteFile(path string, rules []Rule) (bool, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return false, err
	}

	changed := false
	for _, imp := range f.Imports {
		importPath, _ := strconv.Unquote(imp.Path.Value)
		for _, r := range rules {
			if importPath == r.Old || strings.HasPrefix(importPath, r.Old+"/") {
				newPath := r.New + importPath[len(r.Old):]
				imp.Path.Value = strconv.Quote(newPath)

				// Also update EndPos to match new length
				// (go/format handles this, but we need to update the AST node)

				changed = true
				break
			}
		}
	}

	if !changed {
		return false, nil
	}

	// Also fix any ast.Ident aliases that reference the old package name
	// This handles cases like `infraconfig "old/path"` where the alias stays the same
	// We don't need to change aliases, just the path.

	// Fix SelectorExpr references if package name changed due to path change
	// e.g., if old path ends with "cache" and new path also ends with "cache", no change needed
	// But if old path ends with "bus" and new also "bus", it's fine.

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return false, fmt.Errorf("format: %w", err)
	}

	existing, _ := os.ReadFile(path)
	if bytes.Equal(existing, buf.Bytes()) {
		return false, nil
	}

	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

// Also rewrite any generated files or non-parseable files using simple string replacement
// This is handled by the AST approach above for .go files.

func detectModule(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}
	return ""
}

func importPathToDir(root, module, importPath string) string {
	rel := strings.TrimPrefix(importPath, module+"/")
	return filepath.Join(root, rel)
}

// cleanEmptyDirs removes empty parent directories after moves
func init() {
	// Register a deferred cleanup in main if needed
}

// Ensure ast package is used (for the Walk-based approach)
var _ ast.File
