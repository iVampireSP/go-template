package main

import "go/types"

// CloseableField records a field with a cleanup method.
type CloseableField struct {
	VarName string
	Method  string // "Close", "Shutdown", "Stop"
	HasCtx  bool   // method takes context.Context as first param
}

// checkCloseable checks if a type has Close, Shutdown, or Stop methods.
func checkCloseable(t types.Type, varName string) *CloseableField {
	mset := types.NewMethodSet(t)

	for _, methodName := range []string{"Close", "Shutdown", "Stop"} {
		for i := 0; i < mset.Len(); i++ {
			method := mset.At(i)
			if method.Obj().Name() != methodName {
				continue
			}
			sig, ok := method.Type().(*types.Signature)
			if !ok {
				continue
			}

			params := sig.Params()
			hasCtx := false

			switch params.Len() {
			case 0:
				// No params — valid cleanup method
			case 1:
				if isContextType(params.At(0).Type()) {
					hasCtx = true
				} else {
					continue
				}
			default:
				continue
			}

			return &CloseableField{
				VarName: varName,
				Method:  methodName,
				HasCtx:  hasCtx,
			}
		}
	}
	return nil
}

// isContextType checks if a type is context.Context.
func isContextType(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() != nil && obj.Pkg().Path() == "context" && obj.Name() == "Context"
}

// isNilable checks if a type can be compared to nil.
func isNilable(t types.Type) bool {
	switch t.Underlying().(type) {
	case *types.Pointer, *types.Interface, *types.Map, *types.Slice, *types.Chan:
		return true
	}
	return false
}
