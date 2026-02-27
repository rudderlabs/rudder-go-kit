package engine

import (
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/model"
	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/render"
	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/scanner"
)

// ParseProject walks a project and extracts config entries plus warnings.
func ParseProject(rootDir string) ([]model.Entry, []model.Warning, error) {
	var entries []model.Entry
	var groupOrderDeclarations []model.GroupOrderDeclaration
	var warnings []model.Warning
	fset := token.NewFileSet()

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := d.Name()
			if base == "vendor" || base == ".git" || base == "node_modules" {
				return filepath.SkipDir
			}
			// Skip the cdoc tool itself.
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

		// Intentionally AST-only scanning.
		// Do not reintroduce go/types filtering here: it caused hangs and major
		// slowdowns on large repos in real usage.
		scan := scanner.ScanFile(fset, file, path)
		entries = append(entries, scan.Entries...)
		groupOrderDeclarations = append(groupOrderDeclarations, scan.GroupOrderDeclarations...)
		warnings = append(warnings, scan.Warnings...)
		return nil
	})
	if err != nil {
		return nil, warnings, fmt.Errorf("walking directory: %w", err)
	}

	merged, mergeWarnings := model.DeduplicateEntries(entries)
	warnings = append(warnings, mergeWarnings...)
	groupOrderWarnings := model.ApplyProjectGroupOrders(merged, groupOrderDeclarations)
	warnings = append(warnings, groupOrderWarnings...)
	return merged, warnings, nil
}

// GenerateWarnings emits warnings for entries missing descriptions or groups.
func GenerateWarnings(entries []model.Entry) []model.Warning {
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

// FormatMarkdown formats extracted entries as documentation markdown.
func FormatMarkdown(entries []model.Entry, envPrefix string) string {
	return render.FormatMarkdown(entries, envPrefix)
}

type RunOptions struct {
	RootDir       string
	Output        string
	EnvPrefix     string
	ExtraWarnings bool
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

	entries, warnings, err := ParseProject(opts.RootDir)
	if err != nil {
		return err
	}
	if opts.ExtraWarnings {
		warnings = append(warnings, GenerateWarnings(entries)...)
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
