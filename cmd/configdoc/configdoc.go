// Package main provides a tool that extracts configuration options from Go source code
// and generates a markdown documentation table.
//
// It parses Go files using go/ast to find calls to the rudder-go-kit/config package,
// extracts config keys, default values, and descriptions from //configdoc: annotations.
//
// Supported annotations:
//
//	//configdoc:group [N] <Group Name> — sets the group (and optional sort order) for subsequent config entries
//	//configdoc:description <text>   — sets the description for the next config entry (same line or up to 3 lines above)
//	//configdoc:varkey <key>         — provides a key for a non-literal (dynamic) config key argument (up to 10 lines above)
//	//configdoc:vardefault <value>   — provides a default value for a non-literal (dynamic) default argument (up to 10 lines above)
//	//configdoc:ignore               — excludes the config entry from output (same line or 1 line above)
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/rudderlabs/rudder-go-kit/config"
)

// methodFamily describes how to parse arguments for a config getter method.
type methodFamily int

const (
	familySimple   methodFamily = iota // (default, keys...)
	familyInt                          // (default, min, keys...)
	familyDuration                     // (quantity, unit, keys...)
	familyInt64                        // (default, multiplier, keys...)
)

type methodSpec struct {
	family methodFamily
}

var methodSpecs = map[string]methodSpec{
	"GetStringVar":             {family: familySimple},
	"GetBoolVar":               {family: familySimple},
	"GetFloat64Var":            {family: familySimple},
	"GetStringSliceVar":        {family: familySimple},
	"GetReloadableStringVar":   {family: familySimple},
	"GetReloadableFloat64Var":  {family: familySimple},
	"GetIntVar":                {family: familyInt},
	"GetReloadableIntVar":      {family: familyInt},
	"GetDurationVar":           {family: familyDuration},
	"GetReloadableDurationVar": {family: familyDuration},
	"GetInt64Var":              {family: familyInt64},
}

// configEntry represents a single extracted configuration option.
type configEntry struct {
	PrimaryKey  string   // first key argument, used for dedup/sorting
	ConfigKeys  []string // dotted/camelCase config-file keys
	EnvKeys     []string // UPPERCASE_STYLE env var keys (explicit)
	Default     string
	Description string
	Reloadable  bool
	Group       string
	GroupOrder  int // sort order for the group (0 = unordered, placed last)
	File        string
	Line        int
}

// envVarStyle returns true if the key looks like an environment variable (all uppercase + underscores).
var envVarRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

func isEnvVarStyle(key string) bool {
	return envVarRe.MatchString(key)
}

// durationUnits maps time package selectors to abbreviations.
var durationUnits = map[string]string{
	"Nanosecond":  "ns",
	"Microsecond": "µs",
	"Millisecond": "ms",
	"Second":      "s",
	"Minute":      "m",
	"Hour":        "h",
}

// bytesizeMultipliers maps bytesize selectors to abbreviations.
var bytesizeMultipliers = map[string]string{
	"B":  "B",
	"KB": "KB",
	"MB": "MB",
	"GB": "GB",
	"TB": "TB",
}

// parseProject walks the project directory and extracts all config entries.
func parseProject(rootDir string) ([]configEntry, []string, error) {
	var entries []configEntry
	var warnings []string
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
			// Skip the configdoc tool itself
			rel, _ := filepath.Rel(rootDir, path)
			if rel == filepath.Join("cmd", "configdoc") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("warning: failed to parse %s: %v", path, err))
			return nil
		}

		fileEntries, fileWarnings := extractFromFile(fset, file, path)
		entries = append(entries, fileEntries...)
		warnings = append(warnings, fileWarnings...)
		return nil
	})
	if err != nil {
		return nil, warnings, fmt.Errorf("walking directory: %w", err)
	}

	merged, mergeWarnings := deduplicateEntries(entries)
	warnings = append(warnings, mergeWarnings...)
	return merged, warnings, nil
}

// extractFromFile extracts config entries from a single parsed Go file.
func extractFromFile(fset *token.FileSet, file *ast.File, filePath string) ([]configEntry, []string) {
	var entries []configEntry
	var warnings []string

	// Build a map of line number → comment text for quick lookup.
	// Also track configdoc directives.
	type directive struct {
		kind  string // "group", "description", "ignore", "varkey"
		value string
	}
	lineDirectives := make(map[int]directive) // line → directive
	lineComments := make(map[int]string)      // line → raw comment text

	for _, cg := range file.Comments {
		for _, c := range cg.List {
			line := fset.Position(c.Pos()).Line
			text := strings.TrimPrefix(c.Text, "//")
			text = strings.TrimSpace(text)
			lineComments[line] = text

			if v, ok := strings.CutPrefix(text, "configdoc:group "); ok {
				lineDirectives[line] = directive{kind: "group", value: v}
			} else if v, ok := strings.CutPrefix(text, "configdoc:description "); ok {
				lineDirectives[line] = directive{kind: "description", value: v}
			} else if text == "configdoc:ignore" {
				lineDirectives[line] = directive{kind: "ignore"}
			} else if v, ok := strings.CutPrefix(text, "configdoc:varkey "); ok {
				lineDirectives[line] = directive{kind: "varkey", value: v}
			} else if v, ok := strings.CutPrefix(text, "configdoc:vardefault "); ok {
				lineDirectives[line] = directive{kind: "vardefault", value: v}
			}
		}
	}

	// Track the current group and its order as we scan top-to-bottom.
	currentGroup := ""
	currentGroupOrder := 0

	// Collect all config calls with their line numbers first, so we can process directives in order.
	type callInfo struct {
		call *ast.CallExpr
		line int
	}
	var calls []callInfo

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		methodName := selectorMethodName(call.Fun)
		if methodName == "" {
			return true
		}
		if _, ok := methodSpecs[methodName]; !ok {
			return true
		}
		line := fset.Position(call.Pos()).Line
		calls = append(calls, callInfo{call: call, line: line})
		return true
	})

	// Sort calls by line number to process directives in order.
	sort.Slice(calls, func(i, j int) bool { return calls[i].line < calls[j].line })

	// Collect all directive lines sorted.
	var directiveLines []int
	for l := range lineDirectives {
		directiveLines = append(directiveLines, l)
	}
	sort.Ints(directiveLines)

	// Track consumed per-call directives so they aren't reused by subsequent calls.
	consumed := make(map[int]bool)

	for _, ci := range calls {
		call := ci.call
		line := ci.line

		// Update currentGroup and currentGroupOrder based on directives before this line.
		for _, dl := range directiveLines {
			if dl > line {
				break
			}
			if lineDirectives[dl].kind == "group" {
				currentGroup, currentGroupOrder = parseGroupDirective(lineDirectives[dl].value)
			}
		}

		// Check for ignore directive on same line or line above.
		if d, ok := lineDirectives[line]; ok && d.kind == "ignore" && !consumed[line] {
			consumed[line] = true
			continue
		}
		if d, ok := lineDirectives[line-1]; ok && d.kind == "ignore" && !consumed[line-1] {
			consumed[line-1] = true
			continue
		}

		methodName := selectorMethodName(call.Fun)
		spec := methodSpecs[methodName]

		// Collect varkey and vardefault directives above this call (up to 10 lines above, in order).
		var varKeys []string
		var varDefault string
		var varKeyLines []int
		for offset := 10; offset >= 0; offset-- {
			checkLine := line - offset
			if consumed[checkLine] {
				continue
			}
			if d, ok := lineDirectives[checkLine]; ok && d.kind == "varkey" {
				varKeys = append(varKeys, d.value)
				varKeyLines = append(varKeyLines, checkLine)
			}
			if d, ok := lineDirectives[checkLine]; ok && d.kind == "vardefault" {
				varDefault = d.value
				consumed[checkLine] = true
			}
		}

		entry, entryWarnings, err := extractEntry(fset, spec, call, filePath, line, varKeys, varDefault)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("warning: %s:%d: %v", filePath, line, err))
			continue
		}
		warnings = append(warnings, entryWarnings...)

		entry.Reloadable = strings.HasPrefix(methodName, "GetReloadable")

		// Mark consumed varkey lines.
		for _, vkl := range varKeyLines {
			consumed[vkl] = true
		}

		// Look for description directive on same line or lines above (up to 3 lines above).
		for offset := 0; offset <= 3; offset++ {
			checkLine := line - offset
			if consumed[checkLine] {
				continue
			}
			if d, ok := lineDirectives[checkLine]; ok && d.kind == "description" {
				entry.Description = d.value
				consumed[checkLine] = true
				break
			}
		}

		entry.Group = currentGroup
		entry.GroupOrder = currentGroupOrder
		entries = append(entries, entry)
	}

	return entries, warnings
}

// extractEntry extracts a configEntry from a call expression based on the method spec.
func extractEntry(fset *token.FileSet, spec methodSpec, call *ast.CallExpr, filePath string, line int, varKeys []string, varDefault string) (configEntry, []string, error) {
	args := call.Args
	entry := configEntry{
		File: filePath,
		Line: line,
	}

	var keyArgs []ast.Expr
	var defaultExpr ast.Expr // tracks the expression used for the default value
	switch spec.family {
	case familySimple:
		// (default, keys...)
		if len(args) < 2 {
			return entry, nil, fmt.Errorf("expected at least 2 args, got %d", len(args))
		}
		defaultExpr = args[0]
		entry.Default = renderExpr(fset, args[0])
		keyArgs = args[1:]

	case familyInt:
		// (default, min, keys...)
		if len(args) < 3 {
			return entry, nil, fmt.Errorf("expected at least 3 args, got %d", len(args))
		}
		defaultExpr = args[0]
		entry.Default = renderExpr(fset, args[0])
		keyArgs = args[2:]

	case familyDuration:
		// (quantity, unit, keys...)
		if len(args) < 3 {
			return entry, nil, fmt.Errorf("expected at least 3 args, got %d", len(args))
		}
		defaultExpr = args[0]
		entry.Default = renderDuration(fset, args[0], args[1])
		keyArgs = args[2:]

	case familyInt64:
		// (default, multiplier, keys...)
		if len(args) < 3 {
			return entry, nil, fmt.Errorf("expected at least 3 args, got %d", len(args))
		}
		defaultExpr = args[0]
		entry.Default = renderInt64WithMultiplier(fset, args[0], args[1])
		keyArgs = args[2:]

	}

	if varDefault != "" {
		entry.Default = varDefault
	} else if defaultExpr != nil && !isLiteralDefault(defaultExpr) {
		entry.Default = "${" + entry.Default + "}"
	}

	warnings := classifyKeys(&entry, keyArgs, varKeys, filePath, line)

	// Set primary key.
	if len(entry.ConfigKeys) > 0 {
		entry.PrimaryKey = entry.ConfigKeys[0]
	} else if len(entry.EnvKeys) > 0 {
		entry.PrimaryKey = entry.EnvKeys[0]
	}

	return entry, warnings, nil
}

// classifyKeys separates key arguments into config keys and env var keys.
// Non-literal arguments are filled from varKeys (in order). If there are
// non-literal arguments without matching varkey directives, a warning is emitted.
func classifyKeys(entry *configEntry, args []ast.Expr, varKeys []string, filePath string, line int) []string {
	var warnings []string
	varKeyIdx := 0
	for _, arg := range args {
		key := renderStringLit(arg)
		if key == "" {
			// Non-literal key argument — try to fill from //configdoc:varkey directives.
			if varKeyIdx < len(varKeys) {
				key = varKeys[varKeyIdx]
				varKeyIdx++
			} else {
				warnings = append(warnings, fmt.Sprintf("warning: %s:%d: non-literal config key argument without //configdoc:varkey directive", filePath, line))
				continue
			}
		}
		if isEnvVarStyle(key) {
			entry.EnvKeys = append(entry.EnvKeys, key)
		} else {
			entry.ConfigKeys = append(entry.ConfigKeys, key)
		}
	}
	return warnings
}

// isLiteralDefault returns true if the expression is a compile-time literal
// (basic literal, true/false/nil, composite literal, unary of literal, or type conversion of literal).
func isLiteralDefault(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return true
	case *ast.Ident:
		return e.Name == "true" || e.Name == "false" || e.Name == "nil"
	case *ast.CompositeLit:
		return true
	case *ast.UnaryExpr:
		return isLiteralDefault(e.X)
	case *ast.CallExpr:
		// Type conversions like int64(100)
		if len(e.Args) == 1 {
			return isLiteralDefault(e.Args[0])
		}
	}
	return false
}

// renderExpr renders an AST expression as a human-readable default value string.
func renderExpr(fset *token.FileSet, expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		// Strip quotes from string literals.
		if e.Kind == token.STRING {
			if s, err := strconv.Unquote(e.Value); err == nil {
				return s
			}
			return e.Value
		}
		return e.Value
	case *ast.Ident:
		// true, false, nil
		return e.Name
	case *ast.CompositeLit:
		// []string{} → []
		return "[]"
	case *ast.UnaryExpr:
		// -1 etc.
		return e.Op.String() + renderExpr(fset, e.X)
	case *ast.CallExpr:
		// Type conversions like int64(backoff.DefaultInitialInterval)
		if len(e.Args) == 1 {
			return renderExpr(fset, e.Args[0])
		}
	case *ast.SelectorExpr:
		// e.g. backoff.DefaultInitialInterval, time.Second
		return exprToString(fset, expr)
	}
	// Fallback: render the expression as Go source.
	return exprToString(fset, expr)
}

// renderStringLit extracts the string value from a string literal expression.
func renderStringLit(expr ast.Expr) string {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	if s, err := strconv.Unquote(lit.Value); err == nil {
		return s
	}
	return lit.Value
}

// renderDuration renders a duration default from quantity and unit expressions.
func renderDuration(fset *token.FileSet, quantity, unit ast.Expr) string {
	qty := renderExpr(fset, quantity)

	unitStr := durationUnitAbbrev(fset, unit)
	if unitStr != "" {
		// If quantity is a non-numeric expression (e.g. backoff.DefaultInitialInterval), render as-is.
		if !isNumeric(qty) {
			return qty
		}
		if qty == "0" {
			return "0"
		}
		return qty + unitStr
	}
	return qty + " " + exprToString(fset, unit)
}

// isNumeric returns true if s looks like a plain number.
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// durationUnitAbbrev extracts the duration unit abbreviation from a time.X selector.
func durationUnitAbbrev(_ *token.FileSet, expr ast.Expr) string {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	if abbrev, ok := durationUnits[sel.Sel.Name]; ok {
		return abbrev
	}
	return ""
}

// renderInt64WithMultiplier renders an int64 default with a bytesize multiplier.
func renderInt64WithMultiplier(fset *token.FileSet, defaultExpr, multiplierExpr ast.Expr) string {
	def := renderExpr(fset, defaultExpr)
	sel, ok := multiplierExpr.(*ast.SelectorExpr)
	if ok {
		if abbrev, ok := bytesizeMultipliers[sel.Sel.Name]; ok {
			return def + abbrev
		}
	}
	return def + " " + exprToString(fset, multiplierExpr)
}

// exprToString renders an AST expression back to Go source code.
func exprToString(fset *token.FileSet, expr ast.Expr) string {
	var buf strings.Builder
	if err := printer.Fprint(&buf, fset, expr); err != nil {
		return "<expr>"
	}
	return buf.String()
}

// parseGroupDirective parses a group directive value with an optional numeric order prefix.
// "1 Server" → ("Server", 1), "Server" → ("Server", 0), "2 My Server" → ("My Server", 2).
func parseGroupDirective(value string) (string, int) {
	first, rest, hasSpace := strings.Cut(value, " ")
	if hasSpace {
		if n, err := strconv.Atoi(first); err == nil {
			return rest, n
		}
	}
	return value, 0
}

// selectorMethodName extracts the method name from a call expression's function selector.
// Returns "" if not a selector expression.
func selectorMethodName(fun ast.Expr) string {
	if f, ok := fun.(*ast.SelectorExpr); ok {
		return f.Sel.Name
	}
	return ""
}

// deduplicateEntries merges entries with the same primary key.
func deduplicateEntries(entries []configEntry) ([]configEntry, []string) {
	var warnings []string
	seen := make(map[string]int) // primaryKey → index in result
	var result []configEntry

	for _, e := range entries {
		if e.PrimaryKey == "" {
			continue
		}
		if idx, ok := seen[e.PrimaryKey]; ok {
			existing := &result[idx]
			// Merge description.
			if existing.Description == "" && e.Description != "" {
				existing.Description = e.Description
			}
			// Merge group.
			if existing.Group == "" && e.Group != "" {
				existing.Group = e.Group
			} else if existing.Group != "" && e.Group != "" && existing.Group != e.Group {
				warnings = append(warnings, fmt.Sprintf("warning: conflicting groups for %q: %q vs %q", e.PrimaryKey, existing.Group, e.Group))
			}
			// Warn on default mismatch.
			if existing.Default != e.Default {
				warnings = append(warnings, fmt.Sprintf("warning: conflicting defaults for %q: %q vs %q", e.PrimaryKey, existing.Default, e.Default))
			}
			// Merge env keys.
			for _, ek := range e.EnvKeys {
				if !slices.Contains(existing.EnvKeys, ek) {
					existing.EnvKeys = append(existing.EnvKeys, ek)
				}
			}
			// Merge config keys.
			for _, ck := range e.ConfigKeys {
				if !slices.Contains(existing.ConfigKeys, ck) {
					existing.ConfigKeys = append(existing.ConfigKeys, ck)
				}
			}
			// Merge reloadable.
			if e.Reloadable {
				existing.Reloadable = true
			}
			// Merge group order.
			if existing.GroupOrder == 0 && e.GroupOrder != 0 {
				existing.GroupOrder = e.GroupOrder
			}
		} else {
			seen[e.PrimaryKey] = len(result)
			result = append(result, e)
		}
	}
	return result, warnings
}

// formatMarkdown generates the markdown documentation from the extracted entries.
func formatMarkdown(entries []configEntry, envPrefix string) string {
	groupNames := []string{}
	groupMap := make(map[string][]configEntry)
	groupOrderMap := make(map[string]int) // group name → sort order

	for _, e := range entries {
		g := e.Group
		if g == "" {
			g = "Ungrouped"
		}
		if _, ok := groupMap[g]; !ok {
			groupNames = append(groupNames, g)
		}
		groupMap[g] = append(groupMap[g], e)
		if e.GroupOrder != 0 && groupOrderMap[g] == 0 {
			groupOrderMap[g] = e.GroupOrder
		}
	}

	// Sort groups: ordered groups first (by order), then unordered alphabetically.
	sort.SliceStable(groupNames, func(i, j int) bool {
		oi, oj := groupOrderMap[groupNames[i]], groupOrderMap[groupNames[j]]
		if oi != 0 && oj != 0 {
			return oi < oj
		}
		if oi != 0 {
			return true
		}
		if oj != 0 {
			return false
		}
		return groupNames[i] < groupNames[j]
	})

	// Sort entries within each group by primary key.
	for _, groupEntries := range groupMap {
		sort.Slice(groupEntries, func(i, j int) bool {
			return groupEntries[i].PrimaryKey < groupEntries[j].PrimaryKey
		})
	}

	var sb strings.Builder
	sb.WriteString("# Configuration\n\n")
	fmt.Fprintf(&sb, "All configuration is read via environment variables with the `%s_` prefix, or via a configuration file.\n\n", envPrefix)
	sb.WriteString("<!-- This file is auto-generated by configdoc. Do not edit manually. -->\n\n")

	// If all entries are ungrouped, skip group headers entirely.
	singleUngrouped := len(groupNames) == 1 && groupNames[0] == "Ungrouped"

	for _, gName := range groupNames {
		gEntries := groupMap[gName]
		if !singleUngrouped {
			fmt.Fprintf(&sb, "## %s\n\n", gName)
		}
		sb.WriteString("| Config variable | Env variable | Default | Description |\n")
		sb.WriteString("|---|---|---|---|\n")

		for _, e := range gEntries {
			configCol := formatConfigVarColumn(e)
			envCol := formatEnvVarColumn(e, envPrefix)
			defCol := formatDefault(e.Default)
			desc := e.Description
			if e.Reloadable {
				desc = "🔄 " + desc
			}
			fmt.Fprintf(&sb, "| %s | %s | %s | %s |\n", configCol, envCol, defCol, desc)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatConfigVarColumn formats the config variable column.
func formatConfigVarColumn(e configEntry) string {
	var parts []string
	for _, k := range e.ConfigKeys {
		parts = append(parts, "`"+k+"`")
	}
	if len(parts) == 0 {
		// If all keys are env-var style, show the primary key.
		return "`" + e.PrimaryKey + "`"
	}
	return strings.Join(parts, "<br>") // use <br> for line breaks in markdown table
}

// formatEnvVarColumn formats the env variable column.
func formatEnvVarColumn(e configEntry, envPrefix string) string {
	var parts []string

	// Add derived env vars from config keys.
	seen := make(map[string]bool)
	for _, k := range e.ConfigKeys {
		derived := config.ConfigKeyToEnv(envPrefix, k)
		if !seen[derived] {
			parts = append(parts, "`"+derived+"`")
			seen[derived] = true
		}
	}

	// Add explicit env keys that are different from derived ones.
	for _, k := range e.EnvKeys {
		if !seen[k] {
			parts = append(parts, "`"+k+"`")
			seen[k] = true
		}
	}

	return strings.Join(parts, "<br>") // use <br> for line breaks in markdown table
}

// formatDefault formats the default value for display.
func formatDefault(def string) string {
	if def == "" {
		return "``"
	}
	return "`" + def + "`"
}

// generateWarnings produces warnings for entries missing descriptions or groups.
func generateWarnings(entries []configEntry) []string {
	var warnings []string
	for _, e := range entries {
		if e.Description == "" {
			warnings = append(warnings, fmt.Sprintf("warning: %s:%d: config key %q has no //configdoc:description", e.File, e.Line, e.PrimaryKey))
		}
		if e.Group == "" {
			warnings = append(warnings, fmt.Sprintf("warning: %s:%d: config key %q has no //configdoc:group", e.File, e.Line, e.PrimaryKey))
		}
	}
	return warnings
}

func run(rootDir, output, envPrefix string, warn bool) error {
	entries, warnings, err := parseProject(rootDir)
	if err != nil {
		return err
	}

	if warn {
		missingWarnings := generateWarnings(entries)
		warnings = append(warnings, missingWarnings...)
	}

	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, w)
	}

	md := formatMarkdown(entries, envPrefix)

	if output == "" {
		fmt.Print(md)
	} else {
		if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
		if err := os.WriteFile(output, []byte(md), 0o644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	}
	return nil
}
