package engine

import (
	"go/importer"
	"go/token"
	"go/types"
	pathpkg "path"
	"strings"
)

// fastTypesImporter wraps the default importer with a small cache and a stubbed
// config package to keep type-checking lightweight and deterministic.
type fastTypesImporter struct {
	base     types.Importer
	baseFrom types.ImporterFrom
	cache    map[string]*types.Package
}

// newFastTypesImporter builds the importer used by types parse mode.
func newFastTypesImporter() *fastTypesImporter {
	base := importer.Default()
	typedImporter := &fastTypesImporter{
		base:  base,
		cache: make(map[string]*types.Package),
	}
	if from, ok := base.(types.ImporterFrom); ok {
		typedImporter.baseFrom = from
	}
	return typedImporter
}

// Import satisfies types.Importer by delegating to ImportFrom.
func (i *fastTypesImporter) Import(path string) (*types.Package, error) {
	return i.ImportFrom(path, "", 0)
}

// ImportFrom serves cached packages, injects a local stub for config.Config,
// and falls back to the default importer behavior.
func (i *fastTypesImporter) ImportFrom(path, dir string, mode types.ImportMode) (*types.Package, error) {
	if pkg, ok := i.cache[path]; ok {
		return pkg, nil
	}

	if path == configPackagePath {
		pkg := newStubConfigPackage(path)
		i.cache[path] = pkg
		return pkg, nil
	}

	if i.baseFrom != nil {
		if pkg, err := i.baseFrom.ImportFrom(path, dir, mode); err == nil {
			i.cache[path] = pkg
			return pkg, nil
		}
	}
	if i.base != nil {
		if pkg, err := i.base.Import(path); err == nil {
			i.cache[path] = pkg
			return pkg, nil
		}
	}

	pkg := types.NewPackage(path, pathpkg.Base(path))
	pkg.MarkComplete()
	i.cache[path] = pkg
	return pkg, nil
}

// newStubConfigPackage declares only the config.Config methods cdoc needs to
// recognize receiver types during selection resolution.
func newStubConfigPackage(path string) *types.Package {
	pkg := types.NewPackage(path, "config")
	configObj := types.NewTypeName(token.NoPos, pkg, "Config", nil)
	configType := types.NewNamed(configObj, types.NewStruct(nil, nil), nil)
	pkg.Scope().Insert(configObj)
	optType := newStubConfigOptType(pkg, configType)

	getterNames := []string{
		"GetStringVar",
		"GetBoolVar",
		"GetStringSliceVar",
		"GetFloat64Var",
		"GetDurationVar",
		"GetIntVar",
		"GetInt64Var",
		"GetReloadableStringVar",
		"GetReloadableBoolVar",
		"GetReloadableStringSliceVar",
		"GetReloadableFloat64Var",
		"GetReloadableDurationVar",
		"GetReloadableIntVar",
		"GetReloadableInt64Var",
	}
	for _, methodName := range getterNames {
		configType.AddMethod(newStubConfigMethod(pkg, configType, methodName))
	}
	for _, functionName := range getterNames {
		_ = pkg.Scope().Insert(newStubConfigFunction(pkg, functionName))
	}
	_ = pkg.Scope().Insert(newStubConfigNewFunction(pkg, configType, optType))
	_ = pkg.Scope().Insert(newStubWithEnvPrefixFunction(pkg, optType))

	pkg.MarkComplete()
	return pkg
}

// newStubConfigOptType builds the package-level Opt type used by config.New.
func newStubConfigOptType(pkg *types.Package, configType *types.Named) *types.Named {
	optObj := types.NewTypeName(token.NoPos, pkg, "Opt", nil)
	param := types.NewVar(token.NoPos, pkg, "cfg", types.NewPointer(configType))
	sig := types.NewSignatureType(nil, nil, nil, types.NewTuple(param), nil, false)
	optType := types.NewNamed(optObj, sig, nil)
	pkg.Scope().Insert(optObj)
	return optType
}

// newStubConfigMethod creates a permissive variadic signature for stubbed
// methods; argument typing is irrelevant for receiver resolution here.
func newStubConfigMethod(pkg *types.Package, recvType *types.Named, methodName string) *types.Func {
	recv := types.NewVar(token.NoPos, pkg, "cfg", types.NewPointer(recvType))
	param := types.NewVar(token.NoPos, pkg, "args", types.NewSlice(types.NewInterfaceType(nil, nil).Complete()))
	params := types.NewTuple(param)
	result := types.NewVar(token.NoPos, pkg, "value", configMethodResultType(methodName))
	results := types.NewTuple(result)
	signature := types.NewSignatureType(recv, nil, nil, params, results, true)
	return types.NewFunc(token.NoPos, pkg, methodName, signature)
}

// newStubConfigFunction builds package-level config.Get*Var stub signatures.
func newStubConfigFunction(pkg *types.Package, functionName string) *types.Func {
	param := types.NewVar(token.NoPos, pkg, "args", types.NewSlice(types.NewInterfaceType(nil, nil).Complete()))
	params := types.NewTuple(param)
	result := types.NewVar(token.NoPos, pkg, "value", configMethodResultType(functionName))
	results := types.NewTuple(result)
	signature := types.NewSignatureType(nil, nil, nil, params, results, true)
	return types.NewFunc(token.NoPos, pkg, functionName, signature)
}

// newStubConfigNewFunction builds the signature for config.New(opts ...Opt) *Config.
func newStubConfigNewFunction(pkg *types.Package, configType *types.Named, optType types.Type) *types.Func {
	param := types.NewVar(token.NoPos, pkg, "opts", types.NewSlice(optType))
	params := types.NewTuple(param)
	result := types.NewVar(token.NoPos, pkg, "cfg", types.NewPointer(configType))
	results := types.NewTuple(result)
	signature := types.NewSignatureType(nil, nil, nil, params, results, true)
	return types.NewFunc(token.NoPos, pkg, "New", signature)
}

// newStubWithEnvPrefixFunction builds the signature for config.WithEnvPrefix.
func newStubWithEnvPrefixFunction(pkg *types.Package, optType types.Type) *types.Func {
	param := types.NewVar(token.NoPos, pkg, "prefix", types.Typ[types.String])
	params := types.NewTuple(param)
	result := types.NewVar(token.NoPos, pkg, "opt", optType)
	results := types.NewTuple(result)
	signature := types.NewSignatureType(nil, nil, nil, params, results, false)
	return types.NewFunc(token.NoPos, pkg, "WithEnvPrefix", signature)
}

// configMethodResultType keeps return types close to real signatures so method
// sets and downstream inference stay sensible in partial type-checks.
func configMethodResultType(methodName string) types.Type {
	switch {
	case strings.Contains(methodName, "StringSlice"):
		return types.NewSlice(types.Typ[types.String])
	case strings.Contains(methodName, "String"):
		return types.Typ[types.String]
	case strings.Contains(methodName, "Bool"):
		return types.Typ[types.Bool]
	case strings.Contains(methodName, "Float64"):
		return types.Typ[types.Float64]
	case strings.Contains(methodName, "Int64"), strings.Contains(methodName, "Duration"):
		return types.Typ[types.Int64]
	case strings.Contains(methodName, "Int"):
		return types.Typ[types.Int]
	default:
		return types.NewInterfaceType(nil, nil).Complete()
	}
}
