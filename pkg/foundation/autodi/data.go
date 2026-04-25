package main

import (
	"fmt"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// generatePkgDiagram scans every package under the module and produces an interactive
// Sigma.js HTML file showing exported types (structs, interfaces), their fields/methods,
// standalone functions, and inter-type relationships:
//
//   - implements — struct implements an in-module interface
//   - depends    — constructor (New*) parameter references another module type
//   - field      (-->)   — struct field is a module-internal named type
//   - decorator  — a struct both implements an interface AND its constructor depends on
//     the same interface (<<decorator>> annotation)
func generatePkgDiagram(cfg *Config, moduleRoot string) ([]byte, error) {
	gen := &pkgDiagramGen{
		cfg:        cfg,
		moduleRoot: moduleRoot,
		typeIndex:  make(map[string]*pdType),
		pkgOfType:  make(map[string]*pdPkgInfo),
	}
	return gen.generate()
}

// ── Data model ────────────────────────────────────────────────────────────────

type pkgDiagramGen struct {
	cfg        *Config
	moduleRoot string
	pkgInfos   []*pdPkgInfo

	typeIndex map[string]*pdType    // full typeStr → type info
	pkgOfType map[string]*pdPkgInfo // full typeStr → owning package
}

type pdPkgInfo struct {
	PkgPath   string
	RelPath   string // "internal/notify/email"
	ShortName string // "email"
	Structs   []*pdType
	Ifaces    []*pdType
	Funcs     []*pdFunc
}

type pdType struct {
	Name    string
	Named   *types.Named // preserved for Implements checks
	Fields  []pdField
	Methods []pdMethod
}

type pdField struct {
	Name     string
	TypeStr  string // brief, e.g. "*Config"
	FullType string // full, e.g. "*example.com/testapp/internal/config.Config"
}

type pdMethod struct {
	Name            string
	Params          []string // brief type names
	Returns         []string
	ParamFullTypes  []string
	ReturnFullTypes []string
}

type pdFunc struct {
	Name            string
	Params          []string
	Returns         []string
	ParamFullTypes  []string
	ReturnFullTypes []string
}

type pdRelation struct {
	Kind   string // "implements", "depends", "field"
	FromID string
	ToID   string
}

// ── Main pipeline ─────────────────────────────────────────────────────────────

func (g *pkgDiagramGen) generate() ([]byte, error) {
	if err := g.loadAndParse(); err != nil {
		return nil, err
	}
	rels := g.inferRelations()

	// Compute usage counts for node sizing / tooltips.
	refByNode := make(map[string]int)
	implByNode := make(map[string]int)
	for _, r := range rels {
		switch r.Kind {
		case "depends", "field":
			refByNode[r.ToID]++
		case "implements":
			implByNode[r.ToID]++
		}
	}
	return renderPkgHTML(g.pkgInfos, rels, refByNode, implByNode), nil
}

// ── Phase 1: Load & parse ─────────────────────────────────────────────────────

func (g *pkgDiagramGen) loadAndParse() error {
	pattern := g.cfg.Module + "/..."
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes |
			packages.NeedTypesInfo | packages.NeedImports,
		Dir: g.moduleRoot,
	}
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return fmt.Errorf("pkgdiagram load: %w", err)
	}

	for _, pkg := range pkgs {
		if pkg.Types == nil {
			continue
		}
		// Skip root package (generated main)
		if pkg.PkgPath == g.cfg.Module {
			continue
		}
		if !strings.HasPrefix(pkg.PkgPath, g.cfg.Module+"/") {
			continue
		}

		rel := strings.TrimPrefix(pkg.PkgPath, g.cfg.Module+"/")
		info := &pdPkgInfo{
			PkgPath:   pkg.PkgPath,
			RelPath:   rel,
			ShortName: pkg.Name,
		}

		scope := pkg.Types.Scope()
		names := scope.Names()
		sort.Strings(names)

		for _, name := range names {
			obj := scope.Lookup(name)
			if !obj.Exported() {
				continue
			}

			switch o := obj.(type) {
			case *types.TypeName:
				named, ok := o.Type().(*types.Named)
				if !ok {
					continue
				}
				fullStr := types.TypeString(named, nil)
				switch named.Underlying().(type) {
				case *types.Struct:
					t := g.parseStructType(named)
					info.Structs = append(info.Structs, t)
					g.typeIndex[fullStr] = t
					g.pkgOfType[fullStr] = info
				case *types.Interface:
					t := g.parseIfaceType(named)
					info.Ifaces = append(info.Ifaces, t)
					g.typeIndex[fullStr] = t
					g.pkgOfType[fullStr] = info
				}
			case *types.Func:
				sig := o.Type().(*types.Signature)
				if sig.Recv() == nil { // standalone function
					info.Funcs = append(info.Funcs, g.parseFuncObj(o))
				}
			}
		}

		if len(info.Structs) > 0 || len(info.Ifaces) > 0 || len(info.Funcs) > 0 {
			g.pkgInfos = append(g.pkgInfos, info)
		}
	}

	sort.Slice(g.pkgInfos, func(i, j int) bool {
		return g.pkgInfos[i].RelPath < g.pkgInfos[j].RelPath
	})
	return nil
}

// ── Type parsing helpers ──────────────────────────────────────────────────────

func (g *pkgDiagramGen) parseStructType(named *types.Named) *pdType {
	t := &pdType{Name: named.Obj().Name(), Named: named}

	if st, ok := named.Underlying().(*types.Struct); ok {
		for i := 0; i < st.NumFields(); i++ {
			f := st.Field(i)
			if !f.Exported() {
				continue
			}
			full := types.TypeString(f.Type(), nil)
			t.Fields = append(t.Fields, pdField{
				Name:     f.Name(),
				TypeStr:  pdBriefType(full),
				FullType: full,
			})
		}
	}

	mset := types.NewMethodSet(types.NewPointer(named))
	for i := 0; i < mset.Len(); i++ {
		m := mset.At(i)
		if !m.Obj().Exported() {
			continue
		}
		t.Methods = append(t.Methods, g.selToMethod(m))
	}
	return t
}

func (g *pkgDiagramGen) parseIfaceType(named *types.Named) *pdType {
	t := &pdType{Name: named.Obj().Name(), Named: named}
	iface, ok := named.Underlying().(*types.Interface)
	if !ok {
		return t
	}
	for i := 0; i < iface.NumMethods(); i++ {
		m := iface.Method(i)
		if !m.Exported() {
			continue
		}
		t.Methods = append(t.Methods, g.sigToMethod(m.Name(), m.Type().(*types.Signature)))
	}
	return t
}

func (g *pkgDiagramGen) selToMethod(sel *types.Selection) pdMethod {
	return g.sigToMethod(sel.Obj().Name(), sel.Type().(*types.Signature))
}

func (g *pkgDiagramGen) sigToMethod(name string, sig *types.Signature) pdMethod {
	m := pdMethod{Name: name}
	params := sig.Params()
	for i := 0; i < params.Len(); i++ {
		full := types.TypeString(params.At(i).Type(), nil)
		m.Params = append(m.Params, pdBriefType(full))
		m.ParamFullTypes = append(m.ParamFullTypes, full)
	}
	results := sig.Results()
	for i := 0; i < results.Len(); i++ {
		full := types.TypeString(results.At(i).Type(), nil)
		m.Returns = append(m.Returns, pdBriefType(full))
		m.ReturnFullTypes = append(m.ReturnFullTypes, full)
	}
	return m
}

func (g *pkgDiagramGen) parseFuncObj(fn *types.Func) *pdFunc {
	sig := fn.Type().(*types.Signature)
	f := &pdFunc{Name: fn.Name()}
	params := sig.Params()
	for i := 0; i < params.Len(); i++ {
		full := types.TypeString(params.At(i).Type(), nil)
		f.Params = append(f.Params, pdBriefType(full))
		f.ParamFullTypes = append(f.ParamFullTypes, full)
	}
	results := sig.Results()
	for i := 0; i < results.Len(); i++ {
		full := types.TypeString(results.At(i).Type(), nil)
		f.Returns = append(f.Returns, pdBriefType(full))
		f.ReturnFullTypes = append(f.ReturnFullTypes, full)
	}
	return f
}

// ── Phase 2: Infer relationships ──────────────────────────────────────────────

func (g *pkgDiagramGen) inferRelations() []pdRelation {
	var rels []pdRelation
	seen := make(map[string]bool)
	add := func(r pdRelation) {
		key := r.Kind + "|" + r.FromID + "|" + r.ToID
		if !seen[key] {
			seen[key] = true
			rels = append(rels, r)
		}
	}

	// 1. implements: struct → interface
	for _, pkg := range g.pkgInfos {
		for _, st := range pkg.Structs {
			stID := g.typeNodeID(pkg, st.Name)
			for _, pkg2 := range g.pkgInfos {
				for _, iface := range pkg2.Ifaces {
					ifaceU, ok := iface.Named.Underlying().(*types.Interface)
					if !ok {
						continue
					}
					ptrT := types.NewPointer(st.Named)
					if types.Implements(st.Named, ifaceU) || types.Implements(ptrT, ifaceU) {
						add(pdRelation{
							Kind:   "implements",
							FromID: stID,
							ToID:   g.typeNodeID(pkg2, iface.Name),
						})
					}
				}
			}
		}
	}

	// 2. depends: constructor (New*) param references another module type
	for _, pkg := range g.pkgInfos {
		for _, fn := range pkg.Funcs {
			if !strings.HasPrefix(fn.Name, "New") {
				continue
			}
			// Associate deps with the returned struct
			returnedType := g.findReturnedStruct(pkg, fn)
			if returnedType == nil {
				continue
			}
			fromID := g.typeNodeID(pkg, returnedType.Name)
			for _, paramFull := range fn.ParamFullTypes {
				g.addTypeRef(add, "depends", fromID, paramFull)
			}
		}
	}

	// 3. field: exported struct field referencing another module type
	for _, pkg := range g.pkgInfos {
		for _, st := range pkg.Structs {
			stID := g.typeNodeID(pkg, st.Name)
			for _, f := range st.Fields {
				g.addTypeRef(add, "field", stID, f.FullType)
			}
		}
	}

	return rels
}

// findReturnedStruct finds the struct in the same package that a New* function returns.
func (g *pkgDiagramGen) findReturnedStruct(pkg *pdPkgInfo, fn *pdFunc) *pdType {
	for _, retFull := range fn.ReturnFullTypes {
		stripped := pdStripPrefixes(retFull)
		if t, ok := g.typeIndex[stripped]; ok {
			if g.pkgOfType[stripped] == pkg {
				return t
			}
		}
	}
	return nil
}

// addTypeRef adds a relation if the type string references a known module type.
func (g *pkgDiagramGen) addTypeRef(add func(pdRelation), kind, fromID, fullType string) {
	stripped := pdStripPrefixes(fullType)
	t, ok := g.typeIndex[stripped]
	if !ok {
		return
	}
	pkg := g.pkgOfType[stripped]
	toID := g.typeNodeID(pkg, t.Name)
	if toID != fromID {
		add(pdRelation{Kind: kind, FromID: fromID, ToID: toID})
	}
}

// pdStripPrefixes removes all leading *, [] prefixes to get the base named type.
// Handles nested cases like **T, [][]T, []*T, etc.
func pdStripPrefixes(s string) string {
	for {
		switch {
		case strings.HasPrefix(s, "[]"):
			s = s[2:]
		case strings.HasPrefix(s, "*"):
			s = s[1:]
		default:
			return s
		}
	}
}

// ── Decorator detection ───────────────────────────────────────────────────────

// isDecorator detects the decorator/aspect pattern: a struct that BOTH implements
// an interface AND whose constructor (New*) takes that same interface as a parameter.
func (g *pkgDiagramGen) isDecorator(pkg *pdPkgInfo, st *pdType) bool {
	// Find interfaces this struct implements
	implemented := make(map[string]bool) // full interface typeStr → true
	for _, pkg2 := range g.pkgInfos {
		for _, iface := range pkg2.Ifaces {
			ifaceU, ok := iface.Named.Underlying().(*types.Interface)
			if !ok {
				continue
			}
			ptrT := types.NewPointer(st.Named)
			if types.Implements(st.Named, ifaceU) || types.Implements(ptrT, ifaceU) {
				implemented[types.TypeString(iface.Named, nil)] = true
			}
		}
	}
	if len(implemented) == 0 {
		return false
	}

	// Check if any New* function in same package takes an implemented interface as param
	for _, fn := range pkg.Funcs {
		if !strings.HasPrefix(fn.Name, "New") {
			continue
		}
		if g.findReturnedStruct(pkg, fn) != st {
			continue
		}
		for _, paramFull := range fn.ParamFullTypes {
			stripped := pdStripPrefixes(paramFull)
			if implemented[stripped] {
				return true
			}
		}
	}
	return false
}

// ── Node ID ───────────────────────────────────────────────────────────────────

// typeNodeID produces a stable sigma node ID from the full relative package path.
// Uses the full path to avoid collisions between packages like "internal/ring" and "pkg/ring".
func (g *pkgDiagramGen) typeNodeID(pkg *pdPkgInfo, typeName string) string {
	return sanitizeMermaidID(pkg.RelPath + "_" + typeName)
}

// pdBriefType returns a display-safe type name for classDiagram members.
// It strips package paths and pointer/slice prefixes so that "*cobra.Command"
// becomes "Command" — avoiding Mermaid's interpretation of trailing "*" as abstract.
func pdBriefType(typeStr string) string {
	// Strip all leading * and [] for display
	s := typeStr
	var prefix string
	for {
		switch {
		case strings.HasPrefix(s, "[]"):
			prefix += "[]"
			s = s[2:]
			continue
		case strings.HasPrefix(s, "*"):
			// Drop pointer — not needed for diagram readability
			s = s[1:]
			continue
		}
		break
	}
	// Strip package path, keep just the type name
	if dot := strings.LastIndex(s, "."); dot >= 0 {
		s = s[dot+1:]
	}
	return prefix + s
}

// pdFilterParts strips common Go project prefixes for shorter IDs.
func pdFilterParts(parts []string) []string {
	skip := map[string]bool{"internal": true, "cmd": true, "pkg": true}
	var result []string
	for _, p := range parts {
		if !skip[p] {
			result = append(result, p)
		}
	}
	if len(result) == 0 && len(parts) > 0 {
		return parts[len(parts)-1:]
	}
	return result
}
