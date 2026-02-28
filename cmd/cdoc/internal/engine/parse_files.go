package engine

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/model"
)

// parseProjectFiles walks the repo and parses non-test Go files once so both
// AST and types modes can share the same syntax trees.
func parseProjectFiles(rootDir string, fset *token.FileSet) ([]parsedFile, []model.Warning, error) {
	var files []parsedFile
	var warnings []model.Warning

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := d.Name()
			if base == "vendor" || base == ".git" || base == "node_modules" {
				return filepath.SkipDir
			}
			rel, _ := filepath.Rel(rootDir, path)
			if rel == filepath.Join("cmd", "cdoc") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if parseErr != nil {
			warnings = append(warnings, model.Warning{
				Code:    model.WarningCodeParseFailed,
				Message: fmt.Sprintf("failed to parse %s: %v", path, parseErr),
			})
			return nil
		}

		files = append(files, parsedFile{
			path:    path,
			dir:     filepath.Dir(path),
			pkgName: file.Name.Name,
			file:    file,
		})
		return nil
	})
	if err != nil {
		return nil, warnings, fmt.Errorf("walking directory: %w", err)
	}
	return files, warnings, nil
}

type parsedFile struct {
	path    string
	dir     string
	pkgName string
	file    *ast.File
}
