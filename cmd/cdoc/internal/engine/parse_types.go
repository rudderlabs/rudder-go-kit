package engine

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

const configPackagePath = "github.com/rudderlabs/rudder-go-kit/config"

type packageKey struct {
	dir     string
	pkgName string
}

// buildTypeCheckedAllowlist records getter call positions that resolve to
// methods on config.Config so the scanner can ignore lookalike APIs.
func buildTypeCheckedAllowlist(fset *token.FileSet, files []parsedFile) map[string]map[token.Pos]struct{} {
	allowByFile := make(map[string]map[token.Pos]struct{})

	byPackage := make(map[packageKey][]parsedFile)
	potentialPackage := make(map[packageKey]bool)
	for _, parsed := range files {
		key := packageKey{dir: parsed.dir, pkgName: parsed.pkgName}
		byPackage[key] = append(byPackage[key], parsed)
		if hasPotentialGetterCalls(parsed.file) {
			potentialPackage[key] = true
		}
	}

	importer := newFastTypesImporter()
	for key, pkgFiles := range byPackage {
		if !potentialPackage[key] {
			continue
		}

		astFiles := make([]*ast.File, 0, len(pkgFiles))
		for _, parsed := range pkgFiles {
			astFiles = append(astFiles, parsed.file)
		}

		info := &types.Info{
			Selections: make(map[*ast.SelectorExpr]*types.Selection),
			Uses:       make(map[*ast.Ident]types.Object),
		}
		cfg := types.Config{
			Importer: importer,
			Error:    func(error) {},
		}
		_, _ = cfg.Check(key.dir+"/"+key.pkgName, fset, astFiles, info)

		for _, parsed := range pkgFiles {
			ast.Inspect(parsed.file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || !looksLikeGetterMethod(selector.Sel.Name) {
					return true
				}
				if !isConfigMethodSelection(selector, info.Selections) &&
					!isConfigPackageFunction(selector, info.Selections, info.Uses) {
					return true
				}

				if allowByFile[parsed.path] == nil {
					allowByFile[parsed.path] = make(map[token.Pos]struct{})
				}
				allowByFile[parsed.path][call.Pos()] = struct{}{}
				return true
			})
		}
	}

	return allowByFile
}

// hasPotentialGetterCalls is a fast pre-filter to skip package type-checking
// when there are no matching Get*Var call shapes.
func hasPotentialGetterCalls(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if ok && looksLikeGetterMethod(selector.Sel.Name) {
			found = true
			return false
		}
		return true
	})
	return found
}

// looksLikeGetterMethod matches cdoc-supported config getter naming.
func looksLikeGetterMethod(name string) bool {
	return strings.HasPrefix(name, "Get") && strings.HasSuffix(name, "Var")
}

// isConfigMethodSelection checks whether a selector call resolves to
// github.com/rudderlabs/rudder-go-kit/config.Config.
func isConfigMethodSelection(selector *ast.SelectorExpr, selections map[*ast.SelectorExpr]*types.Selection) bool {
	selection, ok := selections[selector]
	if !ok || selection == nil {
		return false
	}
	recv := types.Unalias(selection.Recv())
	for {
		ptr, ok := recv.(*types.Pointer)
		if !ok {
			break
		}
		recv = types.Unalias(ptr.Elem())
	}

	named, ok := recv.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	return obj.Name() == "Config" && obj.Pkg().Path() == configPackagePath
}

// isConfigPackageFunction checks whether a selector resolves to a package-level
// getter function in the config package.
func isConfigPackageFunction(
	selector *ast.SelectorExpr,
	selections map[*ast.SelectorExpr]*types.Selection,
	uses map[*ast.Ident]types.Object,
) bool {
	// Method calls are handled by selection-based checks above.
	if _, ok := selections[selector]; ok {
		return false
	}

	obj, ok := uses[selector.Sel]
	if !ok {
		return false
	}
	fn, ok := obj.(*types.Func)
	if !ok || fn.Pkg() == nil {
		return false
	}
	if fn.Pkg().Path() != configPackagePath || fn.Name() != selector.Sel.Name {
		return false
	}

	sig, ok := fn.Type().(*types.Signature)
	return ok && sig.Recv() == nil
}
