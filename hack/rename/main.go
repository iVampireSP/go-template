package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const templateModuleName = "github.com/iVampireSP/go-template"

func main() {
	root := findProjectRoot()

	fmt.Printf("Current module: %s\n", templateModuleName)
	fmt.Printf("Enter new module name (e.g. github.com/<user>/<project>): ")

	reader := bufio.NewReader(os.Stdin)
	newModName, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	newModName = strings.TrimSpace(newModName)

	if newModName == "" || newModName == templateModuleName {
		fmt.Println("No changes made.")
		return
	}

	fmt.Printf("\nRename module to: %s\n", newModName)
	fmt.Print("Continue? (y/n): ")
	answer, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(answer)) != "y" {
		fmt.Println("Aborted.")
		return
	}

	// Replace module name in go.mod first
	gomodPath := filepath.Join(root, "go.mod")
	if err := replaceInFile(gomodPath, templateModuleName, newModName); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating go.mod: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[OK] Updated go.mod")

	// Replace import paths in all Go files, YAML configs, etc.
	count := 0
	err = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip binary files and this rename tool itself
		ext := filepath.Ext(info.Name())
		if !isTextFile(ext, info.Name()) {
			return nil
		}

		// Skip the rename tool itself
		rel, _ := filepath.Rel(root, path)
		if strings.HasPrefix(rel, "hack/rename") {
			return nil
		}

		changed, err := replaceInFileIfNeeded(path, templateModuleName, newModName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: %s: %v\n", rel, err)
			return nil
		}
		if changed {
			fmt.Printf("  Updated: %s\n", rel)
			count++
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking project: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[OK] Updated %d file(s)\n", count)

	// Update the const in this file too
	renameSelf := filepath.Join(root, "hack", "rename", "main.go")
	_ = replaceInFile(renameSelf, templateModuleName, newModName)

	// Run go mod tidy
	fmt.Println("Running go mod tidy...")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: go mod tidy failed: %v\n", err)
	}

	// Regenerate autodi
	fmt.Println("Running go generate ./generate.go ...")
	cmd = exec.Command("go", "generate", "./generate.go")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: go generate failed: %v\n", err)
	}

	// Regenerate ent
	fmt.Println("Running go generate ./ent/... ...")
	cmd = exec.Command("go", "generate", "./ent/...")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: ent generate failed: %v\n", err)
	}

	fmt.Println("\n[Done] Project renamed successfully!")
	fmt.Printf("  Module: %s\n", newModName)
	fmt.Println("  Run 'go build ./...' to verify.")
}

// findProjectRoot walks up from cwd or hack/rename to find go.mod.
func findProjectRoot() string {
	// Try going up from current directory
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fallback: relative to this script's typical location (hack/rename/)
	if _, err := os.Stat("../../go.mod"); err == nil {
		abs, _ := filepath.Abs("../..")
		return abs
	}

	fmt.Fprintln(os.Stderr, "Error: could not find project root (go.mod)")
	os.Exit(1)
	return ""
}

func isTextFile(ext, name string) bool {
	textExts := map[string]bool{
		".go": true, ".mod": true, ".sum": true, ".yaml": true, ".yml": true,
		".json": true, ".toml": true, ".md": true, ".txt": true, ".sh": true,
		".sql": true, ".proto": true, ".html": true, ".tmpl": true, ".env": true,
		"": true, // Makefile, Dockerfile, etc.
	}
	// Special filenames without extension
	textNames := map[string]bool{
		"Makefile": true, "Dockerfile": true, "Dockerfile2": true,
		".gitignore": true, ".dockerignore": true,
	}
	return textExts[ext] || textNames[name]
}

func replaceInFile(path, old, new string) error {
	input, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	output := strings.ReplaceAll(string(input), old, new)
	return os.WriteFile(path, []byte(output), 0666)
}

func replaceInFileIfNeeded(path, old, new string) (bool, error) {
	input, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	content := string(input)
	if !strings.Contains(content, old) {
		return false, nil
	}
	output := strings.ReplaceAll(content, old, new)
	return true, os.WriteFile(path, []byte(output), 0666)
}
