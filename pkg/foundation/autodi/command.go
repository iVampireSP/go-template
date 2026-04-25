package main

import (
	"fmt"
	"go/types"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
)

// DiscoveredCommand represents a command package found in cmd/.
type DiscoveredCommand struct {
	Name       string        // directory name: "admin", "admin_api", "kafka"
	PkgPath    string        // full import path
	PkgName    string        // Go package name
	StructName string        // return type name: "Admin", "Worker", "Kafka"
	FuncName   string        // constructor: "NewAdmin", "NewWorker", "NewKafka"
	Params     []TypeRef     // constructor parameters (empty for zero-dep)
	Handlers   []HandlerInfo // exported handler methods on the struct
	IsSingle   bool          // has Handle method (leaf command, no subcommands)
}

// HasDeps returns true if the command constructor has parameters.
func (dc *DiscoveredCommand) HasDeps() bool {
	return len(dc.Params) > 0
}

// HandlerInfo describes an exported handler method on a command struct.
type HandlerInfo struct {
	MethodName string // Go method name: "Create", "List", "Handle"
}

// CommandDetector scans cmd/ packages for command definitions.
type CommandDetector struct {
	cfg        *Config
	moduleRoot string
}

// NewCommandDetector creates a command detector.
func NewCommandDetector(cfg *Config, moduleRoot string) *CommandDetector {
	return &CommandDetector{cfg: cfg, moduleRoot: moduleRoot}
}

// Detect loads cmd/ packages and discovers commands.
//
// Detection rules:
//   - Find exported New* functions returning *T where T has Command() *cobra.Command
//   - T must also have handler methods: exported, func(*cobra.Command) error
//   - If T has a Handle method → single command (leaf)
//   - If T has other handler methods (Create, List, etc.) → multi-subcommand
//   - Constructor params determine DI vs zero-dep
func (d *CommandDetector) Detect() ([]*DiscoveredCommand, error) {
	pattern := d.cfg.Module + "/cmd/..."

	pkgCfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedSyntax | packages.NeedFiles | packages.NeedImports,
		Dir: d.moduleRoot,
	}

	pkgs, err := packages.Load(pkgCfg, pattern)
	if err != nil {
		return nil, fmt.Errorf("load cmd packages: %w", err)
	}

	var commands []*DiscoveredCommand
	for _, pkg := range pkgs {
		rel := strings.TrimPrefix(pkg.PkgPath, d.cfg.Module+"/")
		if rel == "cmd" {
			continue
		}

		cmd := d.analyzePackage(pkg, rel)
		if cmd != nil {
			commands = append(commands, cmd)
		}
	}

	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	return commands, nil
}

// analyzePackage scans a cmd/ package for a command constructor.
// Finds the first exported New* function that returns *T where T has
// both Command() *cobra.Command and at least one handler method.
func (d *CommandDetector) analyzePackage(pkg *packages.Package, relPath string) *DiscoveredCommand {
	scope := pkg.Types.Scope()

	names := scope.Names()
	sort.Strings(names)

	for _, name := range names {
		if !strings.HasPrefix(name, "New") || !isExported(name) {
			continue
		}

		obj := scope.Lookup(name)
		funcObj, ok := obj.(*types.Func)
		if !ok {
			continue
		}

		sig := funcObj.Type().(*types.Signature)
		results := sig.Results()

		// Must return exactly 1 result: *T (pointer to named type)
		if results.Len() != 1 {
			continue
		}

		ptrType, ok := results.At(0).Type().(*types.Pointer)
		if !ok {
			continue
		}
		namedType, ok := ptrType.Elem().(*types.Named)
		if !ok {
			continue
		}

		// T must have Command() *cobra.Command method
		if !hasCommandMethod(namedType) {
			continue
		}

		// Find handler methods on *T
		handlers, isSingle := findHandlerMethods(namedType)
		if len(handlers) == 0 {
			continue
		}

		// Extract constructor parameters
		params := sig.Params()
		var paramTypes []TypeRef
		for i := 0; i < params.Len(); i++ {
			t := params.At(i).Type()
			paramTypes = append(paramTypes, TypeRef{
				Type:    t,
				TypeStr: types.TypeString(t, nil),
				PkgPath: typePkgPath(t),
				IsIface: isInterface(t),
			})
		}

		dirName := strings.TrimPrefix(relPath, "cmd/")
		dirName = strings.ReplaceAll(dirName, "/", "_")

		return &DiscoveredCommand{
			Name:       dirName,
			PkgPath:    pkg.PkgPath,
			PkgName:    pkg.Name,
			StructName: namedType.Obj().Name(),
			FuncName:   name,
			Params:     paramTypes,
			Handlers:   handlers,
			IsSingle:   isSingle,
		}
	}

	return nil
}

// hasCommandMethod checks if *T has a Command() *cobra.Command method.
func hasCommandMethod(named *types.Named) bool {
	mset := types.NewMethodSet(types.NewPointer(named))
	for i := 0; i < mset.Len(); i++ {
		method := mset.At(i)
		if method.Obj().Name() != "Command" {
			continue
		}
		sig, ok := method.Type().(*types.Signature)
		if !ok {
			continue
		}
		if sig.Params().Len() != 0 {
			continue
		}
		if sig.Results().Len() != 1 {
			continue
		}
		if isCobraCommandPtr(sig.Results().At(0).Type()) {
			return true
		}
	}
	return false
}

// findHandlerMethods finds exported methods matching func(*cobra.Command) error on *T.
// Returns the handlers and whether the struct has a Handle method (single command).
func findHandlerMethods(named *types.Named) ([]HandlerInfo, bool) {
	mset := types.NewMethodSet(types.NewPointer(named))
	var handlers []HandlerInfo
	isSingle := false

	for i := 0; i < mset.Len(); i++ {
		method := mset.At(i)
		name := method.Obj().Name()

		// Skip unexported, Command, and non-handler methods
		if !method.Obj().Exported() || name == "Command" {
			continue
		}

		sig, ok := method.Type().(*types.Signature)
		if !ok {
			continue
		}

		// Handler signature: exactly 1 param (*cobra.Command), returns error
		if sig.Params().Len() != 1 || sig.Results().Len() != 1 {
			continue
		}
		if !isCobraCommandPtr(sig.Params().At(0).Type()) {
			continue
		}
		if !isErrorType(sig.Results().At(0).Type()) {
			continue
		}

		handlers = append(handlers, HandlerInfo{MethodName: name})
		if name == "Handle" {
			isSingle = true
		}
	}

	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].MethodName < handlers[j].MethodName
	})

	return handlers, isSingle
}

// isCobraCommandPtr checks if a type is *cobra.Command.
func isCobraCommandPtr(t types.Type) bool {
	ptr, ok := t.(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := ptr.Elem().(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() != nil && obj.Pkg().Path() == "github.com/spf13/cobra" && obj.Name() == "Command"
}

// isExported checks if a name is exported (starts with uppercase).
func isExported(name string) bool {
	if name == "" {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}

// pascalToKebab converts PascalCase method names to kebab-case command names.
// Create → create, UpdatePassword → update-password, CleanSuspended → clean-suspended
func pascalToKebab(s string) string {
	var buf strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			buf.WriteByte('-')
		}
		buf.WriteRune(unicode.ToLower(r))
	}
	return buf.String()
}
