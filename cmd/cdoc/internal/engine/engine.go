package engine

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"os"
	"path/filepath"

	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/model"
	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/render"
	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/scanner"
)

type ParseMode string

const (
	ParseModeAST     ParseMode = "ast"
	ParseModeTypes   ParseMode = "types"
	DefaultParseMode           = ParseModeTypes
)

type RunOptions struct {
	RootDir       string
	Output        string
	EnvPrefix     string
	ExtraWarnings bool
	ParseMode     ParseMode
	Policy        *model.WarningPolicy
	Stdout        io.Writer
	Stderr        io.Writer
}

// Run executes extraction, warning handling, and markdown output.
func Run(opts RunOptions) error {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	policy := model.DefaultWarningPolicy()
	if opts.Policy != nil {
		policy = *opts.Policy
	}

	entries, warnings, err := parseProjectWithMode(opts.RootDir, opts.ParseMode)
	if err != nil {
		return err
	}
	if opts.ExtraWarnings {
		warnings = append(warnings, generateExtraWarnings(entries)...)
	}

	classified := model.ClassifyWarnings(warnings, policy)
	errorCount := 0
	for _, classifiedWarning := range classified {
		switch classifiedWarning.Severity {
		case model.SeverityIgnore:
			continue
		case model.SeverityError:
			errorCount++
		}
		fmt.Fprintln(stderr, classifiedWarning.Warning.String())
	}

	markdown := render.FormatMarkdown(entries, opts.EnvPrefix)
	if opts.Output == "" {
		fmt.Fprint(stdout, markdown)
	} else {
		if err := os.MkdirAll(filepath.Dir(opts.Output), 0o755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
		if err := os.WriteFile(opts.Output, []byte(markdown), 0o644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("found %d warning(s)", errorCount)
	}
	return nil
}

// parseProjectWithMode walks a project and extracts config entries plus warnings.
func parseProjectWithMode(rootDir string, mode ParseMode) ([]model.Entry, []model.Warning, error) {
	switch mode {
	case ParseModeAST, ParseModeTypes:
	default:
		return nil, nil, fmt.Errorf("invalid parse mode %q (expected %q or %q)", mode, ParseModeAST, ParseModeTypes)
	}

	var entries []model.Entry
	var groupOrderDeclarations []model.GroupOrderDeclaration
	fset := token.NewFileSet()
	files, warnings, err := parseProjectFiles(rootDir, fset)
	if err != nil {
		return nil, warnings, err
	}

	var allowCallsByFile map[string]map[token.Pos]struct{}
	if mode == ParseModeTypes {
		allowCallsByFile = buildTypeCheckedAllowlist(fset, files)
	}

	for _, parsed := range files {
		scan := scanParsedFileWithMode(fset, parsed, mode, allowCallsByFile)
		entries = append(entries, scan.Entries...)
		groupOrderDeclarations = append(groupOrderDeclarations, scan.GroupOrderDeclarations...)
		warnings = append(warnings, scan.Warnings...)
	}

	merged, mergeWarnings := model.DeduplicateEntries(entries)
	warnings = append(warnings, mergeWarnings...)
	groupOrderWarnings := model.ApplyProjectGroupOrders(merged, groupOrderDeclarations)
	warnings = append(warnings, groupOrderWarnings...)
	return merged, warnings, nil
}

func scanParsedFileWithMode(fset *token.FileSet, parsed parsedFile, mode ParseMode, allowCallsByFile map[string]map[token.Pos]struct{}) scanner.FileScanResult {
	scanOpts := scanner.ScanOptions{}

	if mode == ParseModeTypes {
		scanOpts.CallFilter = func(call *ast.CallExpr) bool {
			_, ok := allowCallsByFile[parsed.path][call.Pos()]
			return ok
		}
	}
	return scanner.ScanFile(fset, parsed.file, parsed.path, scanOpts)
}

func scanSingleFileWithMode(fset *token.FileSet, file *ast.File, filePath string, mode ParseMode) (scanner.FileScanResult, error) {
	switch mode {
	case ParseModeAST, ParseModeTypes:
	default:
		return scanner.FileScanResult{}, fmt.Errorf("invalid parse mode %q (expected %q or %q)", mode, ParseModeAST, ParseModeTypes)
	}
	parsed := parsedFile{
		path:    filePath,
		dir:     filepath.Dir(filePath),
		pkgName: file.Name.Name,
		file:    file,
	}
	var allowCallsByFile map[string]map[token.Pos]struct{}
	if mode == ParseModeTypes {
		allowCallsByFile = buildTypeCheckedAllowlist(fset, []parsedFile{parsed})
	}
	return scanParsedFileWithMode(fset, parsed, mode, allowCallsByFile), nil
}

// generateExtraWarnings emits warnings for entries missing descriptions or groups.
func generateExtraWarnings(entries []model.Entry) []model.Warning {
	var warnings []model.Warning
	for _, entry := range entries {
		if entry.Description == "" {
			warnings = append(warnings, model.Warning{
				Code:    model.WarningCodeMissingDescription,
				File:    entry.File,
				Line:    entry.Line,
				Message: fmt.Sprintf("config key %q has no //cdoc:desc", entry.PrimaryKey),
			})
		}
		if entry.Group == "" {
			warnings = append(warnings, model.Warning{
				Code:    model.WarningCodeMissingGroup,
				File:    entry.File,
				Line:    entry.Line,
				Message: fmt.Sprintf("config key %q has no //cdoc:group", entry.PrimaryKey),
			})
		}
	}
	return warnings
}
