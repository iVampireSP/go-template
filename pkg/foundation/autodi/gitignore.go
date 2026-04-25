package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// GitignorePattern represents a single gitignore pattern.
type GitignorePattern struct {
	Pattern  string
	Negation bool
	DirOnly  bool
}

// LoadGitignore parses .gitignore from the module root.
func LoadGitignore(root string) []GitignorePattern {
	path := filepath.Join(root, ".gitignore")
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var patterns []GitignorePattern
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		p := GitignorePattern{}
		if strings.HasPrefix(line, "!") {
			p.Negation = true
			line = line[1:]
		}
		if strings.HasSuffix(line, "/") {
			p.DirOnly = true
			line = strings.TrimSuffix(line, "/")
		}
		p.Pattern = line
		patterns = append(patterns, p)
	}
	return patterns
}

// IsGitignored checks if a relative path matches any gitignore pattern.
func IsGitignored(relPath string, patterns []GitignorePattern) bool {
	// Normalize to forward slashes
	relPath = filepath.ToSlash(relPath)

	ignored := false
	for _, p := range patterns {
		if matchGitignore(relPath, p.Pattern) {
			if p.Negation {
				ignored = false
			} else {
				ignored = true
			}
		}
	}
	return ignored
}

// matchGitignore performs simplified gitignore matching.
func matchGitignore(path, pattern string) bool {
	// Leading / means anchored to root
	if strings.HasPrefix(pattern, "/") {
		pattern = pattern[1:]
		matched, _ := filepath.Match(pattern, path)
		return matched
	}

	// Pattern with / anywhere means match from root
	if strings.Contains(pattern, "/") {
		matched, _ := filepath.Match(pattern, path)
		if matched {
			return true
		}
		// Also try prefix match for directories
		return strings.HasPrefix(path, pattern+"/") || strings.HasPrefix(path, pattern)
	}

	// No /, match against any path component or the basename
	base := filepath.Base(path)
	if matched, _ := filepath.Match(pattern, base); matched {
		return true
	}

	// Try matching against each path segment
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if matched, _ := filepath.Match(pattern, part); matched {
			return true
		}
	}
	return false
}
