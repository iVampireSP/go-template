package main

import (
	"go/ast"
	"strings"
)

// Annotation types
const (
	AnnotBind     = "bind"     // //autodi:bind InterfaceName
	AnnotIgnore   = "ignore"   // //autodi:ignore
	AnnotInvoke   = "invoke"   // //autodi:invoke
	AnnotOptional = "optional" // //autodi:optional ParamType
)

// Annotation represents a parsed //autodi: directive.
type Annotation struct {
	Kind  string // bind, ignore, invoke, optional
	Value string // argument (e.g., interface name for bind)
}

// ParseAnnotations extracts //autodi: directives from a function's doc comments.
func ParseAnnotations(fn *ast.FuncDecl) []Annotation {
	if fn.Doc == nil {
		return nil
	}

	var annotations []Annotation
	for _, comment := range fn.Doc.List {
		text := strings.TrimSpace(comment.Text)
		// Remove leading //
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimSpace(text)

		if !strings.HasPrefix(text, "autodi:") {
			continue
		}
		text = strings.TrimPrefix(text, "autodi:")

		parts := strings.SplitN(text, " ", 2)
		kind := strings.TrimSpace(parts[0])
		value := ""
		if len(parts) > 1 {
			value = strings.TrimSpace(parts[1])
		}

		switch kind {
		case AnnotBind, AnnotIgnore, AnnotInvoke, AnnotOptional:
			annotations = append(annotations, Annotation{Kind: kind, Value: value})
		}
	}
	return annotations
}

// HasAnnotation checks if annotations contain a specific kind.
func HasAnnotation(annotations []Annotation, kind string) bool {
	for _, a := range annotations {
		if a.Kind == kind {
			return true
		}
	}
	return false
}

// GetAnnotationValues returns all values for a specific annotation kind.
func GetAnnotationValues(annotations []Annotation, kind string) []string {
	var values []string
	for _, a := range annotations {
		if a.Kind == kind && a.Value != "" {
			values = append(values, a.Value)
		}
	}
	return values
}
