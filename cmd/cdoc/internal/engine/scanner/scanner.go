package scanner

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/model"
)

const maxLookback = 10

var envVarRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// IsEnvVarStyle reports whether key matches UPPER_SNAKE_CASE env style.
func IsEnvVarStyle(key string) bool {
	return envVarRe.MatchString(key)
}

// FileScanResult captures the scanner output for one file.
type FileScanResult struct {
	Entries                []model.Entry
	GroupOrderDeclarations []model.GroupOrderDeclaration
	Warnings               []model.Warning
}

type ScanOptions struct {
	CallFilter func(*ast.CallExpr) bool
}

// ScanFile scans one parsed Go file and returns extracted entries, project-level
// group order declarations, and warnings.
// ScanOptions.CallFilter applies additional filtering on matched getter calls.
// A nil filter accepts all matched calls.
func ScanFile(fset *token.FileSet, file *ast.File, filePath string, opts ScanOptions) FileScanResult {
	lineDirectives := collectDirectives(fset, file)
	directives := flattenDirectives(lineDirectives)

	result := FileScanResult{
		GroupOrderDeclarations: extractGroupOrderDeclarations(filePath, directives),
	}

	// Extract entries and warnings in one pass over the event stream.
	result.Entries, result.Warnings = extractEntriesWithWarnings(fset, file, filePath, directives, opts.CallFilter)
	return result
}

type directiveKind int

const (
	directiveGroup directiveKind = iota
	directiveDesc
	directiveIgnore
	directiveKey
	directiveDefault
)

type directive struct {
	kind  directiveKind
	value string
	line  int
	order int
}

// collectDirectives parses cdoc directives from comments keyed by source line.
func collectDirectives(fset *token.FileSet, file *ast.File) map[int][]directive {
	lineDirectives := make(map[int][]directive)
	order := 0
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			line := fset.Position(c.Pos()).Line
			text := strings.TrimPrefix(c.Text, "//")
			text = strings.TrimSpace(text)

			if value, ok := strings.CutPrefix(text, "cdoc:group "); ok {
				lineDirectives[line] = append(lineDirectives[line], directive{kind: directiveGroup, value: value, line: line, order: order})
			} else if value, ok := strings.CutPrefix(text, "cdoc:desc "); ok {
				lineDirectives[line] = append(lineDirectives[line], directive{kind: directiveDesc, value: value, line: line, order: order})
			} else if text == "cdoc:ignore" {
				lineDirectives[line] = append(lineDirectives[line], directive{kind: directiveIgnore, line: line, order: order})
			} else if value, ok := strings.CutPrefix(text, "cdoc:key "); ok {
				lineDirectives[line] = append(lineDirectives[line], directive{kind: directiveKey, value: value, line: line, order: order})
			} else if value, ok := strings.CutPrefix(text, "cdoc:default "); ok {
				lineDirectives[line] = append(lineDirectives[line], directive{kind: directiveDefault, value: value, line: line, order: order})
			}
			order++
		}
	}
	return lineDirectives
}

// flattenDirectives returns directives sorted by line while preserving line-local order.
func flattenDirectives(lineDirectives map[int][]directive) []directive {
	var lines []int
	for line := range lineDirectives {
		lines = append(lines, line)
	}
	sort.Ints(lines)

	var directives []directive
	for _, line := range lines {
		directives = append(directives, lineDirectives[line]...)
	}
	return directives
}

type methodFamily int

const (
	familySimple         methodFamily = iota // (default, keys...)
	familyDuration                           // (quantity, unit, keys...)
	familyWithMultiplier                     // (default, unit, keys...)
)

type methodSpec struct {
	family     methodFamily
	reloadable bool
}

// classifyMethod maps a getter method name to extraction semantics.
func classifyMethod(name string) (methodSpec, bool) {
	if !strings.HasPrefix(name, "Get") {
		return methodSpec{}, false
	}
	base := strings.TrimPrefix(name, "Get")
	reloadable := strings.HasPrefix(base, "Reloadable")
	if reloadable {
		base = strings.TrimPrefix(base, "Reloadable")
	}
	if !strings.HasSuffix(base, "Var") {
		return methodSpec{}, false
	}
	base = strings.TrimSuffix(base, "Var")

	spec := methodSpec{reloadable: reloadable}
	switch base {
	case "String", "Bool", "StringSlice", "Float64":
		spec.family = familySimple
	case "Duration":
		spec.family = familyDuration
	case "Int", "Int64":
		spec.family = familyWithMultiplier
	default:
		return methodSpec{}, false
	}
	return spec, true
}

type callInfo struct {
	call      *ast.CallExpr
	startLine int
	endLine   int
	pos       token.Pos
	spec      methodSpec
}

// collectCalls finds supported getter calls and records their scan metadata.
func collectCalls(fset *token.FileSet, file *ast.File, callFilter func(*ast.CallExpr) bool) []callInfo {
	var calls []callInfo
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		methodName := selectorMethodName(call.Fun)
		spec, ok := classifyMethod(methodName)
		if !ok {
			return true
		}
		if callFilter != nil && !callFilter(call) {
			return true
		}
		calls = append(calls, callInfo{
			call:      call,
			startLine: fset.Position(call.Pos()).Line,
			endLine:   fset.Position(call.End()).Line,
			pos:       call.Pos(),
			spec:      spec,
		})
		return true
	})
	return calls
}

type eventKind int

const (
	eventDirective eventKind = iota
	eventCall
)

type fileEvent struct {
	line int
	kind eventKind
	dir  directive
	call callInfo
}

// buildEvents merges directives and calls into one ordered event stream.
func buildEvents(calls []callInfo, directives []directive) []fileEvent {
	events := make([]fileEvent, 0, len(calls)+len(directives))
	for _, dir := range directives {
		events = append(events, fileEvent{line: dir.line, kind: eventDirective, dir: dir})
	}
	for _, call := range calls {
		// Process a call after all directives inside its span are discovered.
		events = append(events, fileEvent{line: call.endLine, kind: eventCall, call: call})
	}

	sort.SliceStable(events, func(i, j int) bool {
		if events[i].line != events[j].line {
			return events[i].line < events[j].line
		}
		if events[i].kind != events[j].kind {
			// Apply same-line directives to same-line calls.
			return events[i].kind == eventDirective
		}
		if events[i].kind == eventDirective {
			return events[i].dir.order < events[j].dir.order
		}
		if events[i].call.pos != events[j].call.pos {
			return events[i].call.pos < events[j].call.pos
		}
		return events[i].call.startLine < events[j].call.startLine
	})

	return events
}

type pendingDirective struct {
	directive directive
	consumed  bool
}

// extractEntriesWithWarnings drives directive consumption and entry extraction.
func extractEntriesWithWarnings(fset *token.FileSet, file *ast.File, filePath string, directives []directive, callFilter func(*ast.CallExpr) bool) ([]model.Entry, []model.Warning) {
	var entries []model.Entry
	var warnings []model.Warning

	calls := collectCalls(fset, file, callFilter)
	events := buildEvents(calls, directives)

	currentGroup := ""
	currentGroupOrder := 0
	var pending []pendingDirective

	for _, event := range events {
		if event.kind == eventDirective {
			switch event.dir.kind {
			case directiveGroup:
				currentGroup, currentGroupOrder = parseGroupDirective(event.dir.value)
			default:
				pending = append(pending, pendingDirective{directive: event.dir})
			}
			continue
		}

		callStartLine := event.call.startLine
		callEndLine := event.call.endLine
		if consumeIgnoreDirective(pending, callStartLine) {
			continue
		}

		description, hasDescription := consumeNearestDirectiveValue(pending, directiveDesc, callStartLine, 10)
		varKeys, varDefault, keyDirectiveIndexes := collectKeyAndDefaultOverrides(pending, callStartLine, callEndLine)
		entry, entryWarnings, err := extractEntry(fset, event.call, filePath, callStartLine, varKeys, varDefault)
		for _, idx := range keyDirectiveIndexes {
			pending[idx].consumed = true
		}
		if err != nil {
			warnings = append(warnings, model.Warning{
				Code:    model.WarningCodeArgCount,
				File:    filePath,
				Line:    callStartLine,
				Message: err.Error(),
			})
			continue
		}
		warnings = append(warnings, entryWarnings...)

		if hasDescription {
			entry.Description = description
		}

		entry.Group = currentGroup
		entry.GroupOrder = currentGroupOrder
		entries = append(entries, entry)
	}

	for _, pendingDirective := range pending {
		if pendingDirective.consumed {
			continue
		}
		switch pendingDirective.directive.kind {
		case directiveDesc:
			warnings = append(warnings, model.Warning{
				Code:    model.WarningCodeUnusedDescDirective,
				File:    filePath,
				Line:    pendingDirective.directive.line,
				Message: "unused //cdoc:desc directive",
			})
		case directiveDefault:
			warnings = append(warnings, model.Warning{
				Code:    model.WarningCodeUnusedDefaultDirective,
				File:    filePath,
				Line:    pendingDirective.directive.line,
				Message: "unused //cdoc:default directive",
			})
		}
	}

	return entries, warnings
}

// consumeIgnoreDirective applies //cdoc:ignore on the call line or previous line.
func consumeIgnoreDirective(pending []pendingDirective, line int) bool {
	if idx := findUnconsumedDirectiveOnLine(pending, directiveIgnore, line); idx >= 0 {
		pending[idx].consumed = true
		return true
	}
	if idx := findUnconsumedDirectiveOnLine(pending, directiveIgnore, line-1); idx >= 0 {
		pending[idx].consumed = true
		return true
	}
	return false
}

// collectKeyAndDefaultOverrides gathers //cdoc:key and //cdoc:default directives
// from the start-line lookback window and from inside the multiline call span.
func collectKeyAndDefaultOverrides(pending []pendingDirective, callStartLine, callEndLine int) ([]string, string, []int) {
	var varKeys []string
	var varDefault string
	var keyDirectiveIndexes []int

	for i := range pending {
		if pending[i].consumed {
			continue
		}
		if !directiveAppliesToKeyOrDefault(pending[i].directive.line, callStartLine, callEndLine, maxLookback) {
			continue
		}

		switch pending[i].directive.kind {
		case directiveKey:
			varKeys = append(varKeys, parseVarKeyDirective(pending[i].directive.value)...)
			keyDirectiveIndexes = append(keyDirectiveIndexes, i)
		case directiveDefault:
			varDefault = pending[i].directive.value
			pending[i].consumed = true
		}
	}
	return varKeys, varDefault, keyDirectiveIndexes
}

func directiveAppliesToKeyOrDefault(directiveLine, callStartLine, callEndLine, maxLookback int) bool {
	if inLineWindow(directiveLine, callStartLine-maxLookback, callStartLine) {
		return true
	}
	return inLineWindow(directiveLine, callStartLine, callEndLine)
}

func inLineWindow(line, start, end int) bool {
	return line >= start && line <= end
}

// consumeNearestDirectiveValue consumes the nearest matching directive within maxDistance.
func consumeNearestDirectiveValue(pending []pendingDirective, kind directiveKind, line, maxDistance int) (string, bool) {
	for target := line; target >= line-maxDistance; target-- {
		for i := len(pending) - 1; i >= 0; i-- {
			if pending[i].consumed {
				continue
			}
			if pending[i].directive.kind != kind || pending[i].directive.line != target {
				continue
			}
			pending[i].consumed = true
			return pending[i].directive.value, true
		}
	}
	return "", false
}

// findUnconsumedDirectiveOnLine returns the pending index for kind on line.
func findUnconsumedDirectiveOnLine(pending []pendingDirective, kind directiveKind, line int) int {
	for i := range pending {
		if pending[i].consumed {
			continue
		}
		if pending[i].directive.kind == kind && pending[i].directive.line == line {
			return i
		}
	}
	return -1
}

// extractGroupOrderDeclarations collects project-level group ordering directives.
func extractGroupOrderDeclarations(filePath string, directives []directive) []model.GroupOrderDeclaration {
	declarations := make([]model.GroupOrderDeclaration, 0)
	for _, dir := range directives {
		if dir.kind != directiveGroup {
			continue
		}
		group, order := parseGroupDirective(dir.value)
		group = strings.TrimSpace(group)
		if group == "" || order == 0 {
			continue
		}
		declarations = append(declarations, model.GroupOrderDeclaration{
			Group: group,
			Order: order,
			File:  filePath,
			Line:  dir.line,
		})
	}
	return declarations
}

// parseGroupDirective parses a group directive value with an optional numeric
// order prefix.
func parseGroupDirective(value string) (string, int) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", 0
	}
	parts := strings.Fields(value)
	if len(parts) >= 2 {
		if n, err := strconv.Atoi(parts[0]); err == nil {
			return strings.Join(parts[1:], " "), n
		}
	}
	return value, 0
}

// selectorMethodName extracts the method name from a call expression's function
// selector and returns "" for non-selector expressions.
func selectorMethodName(fun ast.Expr) string {
	if selector, ok := fun.(*ast.SelectorExpr); ok {
		return selector.Sel.Name
	}
	return ""
}

// extractEntry converts a matched getter call into a normalized config entry.
func extractEntry(fset *token.FileSet, ci callInfo, filePath string, line int, varKeys []string, varDefault string) (model.Entry, []model.Warning, error) {
	args := ci.call.Args
	entry := model.Entry{
		File:       filePath,
		Line:       line,
		Reloadable: ci.spec.reloadable,
	}

	var keyArgs []ast.Expr
	var defaultExpr ast.Expr // expression used for default rendering policy

	switch ci.spec.family {
	case familySimple:
		// (default, keys...)
		if len(args) < 2 {
			return entry, nil, fmt.Errorf("expected at least 2 args, got %d", len(args))
		}
		defaultExpr = args[0]
		entry.Default = renderExpr(fset, args[0])
		keyArgs = args[1:]

	case familyDuration:
		// (quantity, unit, keys...)
		if len(args) < 3 {
			return entry, nil, fmt.Errorf("expected at least 3 args, got %d", len(args))
		}
		defaultExpr = args[0]
		entry.Default = renderDuration(fset, args[0], args[1])
		keyArgs = args[2:]

	case familyWithMultiplier:
		// (default, unit, keys...)
		if len(args) < 3 {
			return entry, nil, fmt.Errorf("expected at least 3 args, got %d", len(args))
		}
		entry.Default = renderWithUnit(fset, args[0], args[1])
		// renderWithUnit handles ${} wrapping internally.
		defaultExpr = nil
		keyArgs = args[2:]
	}

	if varDefault != "" {
		entry.Default = varDefault
	} else if defaultExpr != nil && !isLiteralDefault(defaultExpr) {
		entry.Default = "${" + entry.Default + "}"
	}

	keyArgsVariadic := ci.call.Ellipsis != token.NoPos && len(keyArgs) > 0
	allKeys, warnings := classifyKeys(&entry, keyArgs, varKeys, keyArgsVariadic, filePath, line)
	entry.PrimaryKey = strings.Join(allKeys, ",")

	return entry, warnings, nil
}

// classifyKeys separates key arguments into config keys and env var keys and
// returns all resolved keys in original argument order.
func classifyKeys(entry *model.Entry, args []ast.Expr, varKeys []string, variadicLastArg bool, filePath string, line int) ([]string, []model.Warning) {
	var warnings []model.Warning
	var allKeys []string
	varKeyIdx := 0

	addKey := func(key string) {
		key = strings.TrimSpace(key)
		if key == "" {
			return
		}
		allKeys = append(allKeys, key)
		if IsEnvVarStyle(key) {
			entry.EnvKeys = append(entry.EnvKeys, key)
		} else {
			entry.ConfigKeys = append(entry.ConfigKeys, key)
		}
	}

	for i, arg := range args {
		key := renderStringLit(arg)
		if key == "" {
			if varKeyIdx < len(varKeys) {
				if variadicLastArg && i == len(args)-1 {
					for ; varKeyIdx < len(varKeys); varKeyIdx++ {
						addKey(varKeys[varKeyIdx])
					}
					continue
				}
				addKey(varKeys[varKeyIdx])
				varKeyIdx++
				continue
			}
			warnings = append(warnings, model.Warning{
				Code:    model.WarningCodeDynamicKeyMissing,
				File:    filePath,
				Line:    line,
				Message: "non-literal config key argument without //cdoc:key directive",
			})
			continue
		}
		addKey(key)
	}

	if varKeyIdx < len(varKeys) {
		unused := strings.Join(varKeys[varKeyIdx:], ", ")
		warnings = append(warnings, model.Warning{
			Code:    model.WarningCodeUnusedKeyOverride,
			File:    filePath,
			Line:    line,
			Message: fmt.Sprintf("unused //cdoc:key override(s): %s", unused),
		})
	}

	return allKeys, warnings
}

// parseVarKeyDirective splits a //cdoc:key directive value by comma.
func parseVarKeyDirective(value string) []string {
	parts := strings.Split(value, ",")
	keys := make([]string, 0, len(parts))
	for _, part := range parts {
		key := strings.TrimSpace(part)
		if key != "" {
			keys = append(keys, key)
		}
	}
	return keys
}

var builtinTypeNames = map[string]struct{}{
	"bool":       {},
	"string":     {},
	"int":        {},
	"int8":       {},
	"int16":      {},
	"int32":      {},
	"int64":      {},
	"uint":       {},
	"uint8":      {},
	"uint16":     {},
	"uint32":     {},
	"uint64":     {},
	"uintptr":    {},
	"byte":       {},
	"rune":       {},
	"float32":    {},
	"float64":    {},
	"complex64":  {},
	"complex128": {},
}

// isBuiltinTypeConversionCall reports whether call has one argument and calls a
// builtin type conversion function such as int(...), string(...), or float64(...).
func isBuiltinTypeConversionCall(call *ast.CallExpr) bool {
	if len(call.Args) != 1 {
		return false
	}
	funIdent, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}
	_, ok = builtinTypeNames[funIdent.Name]
	return ok
}

// isLiteralDefault reports whether expr should be rendered without ${...} wrapping.
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
		if isBuiltinTypeConversionCall(e) {
			return isLiteralDefault(e.Args[0])
		}
	}
	return false
}

// renderExpr renders an AST expression into cdoc's default-value representation.
func renderExpr(fset *token.FileSet, expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			if s, err := strconv.Unquote(e.Value); err == nil {
				return s
			}
			return e.Value
		}
		return e.Value
	case *ast.Ident:
		return e.Name
	case *ast.CompositeLit:
		if len(e.Elts) == 0 {
			return "[]"
		}
		elts := make([]string, len(e.Elts))
		for i, elt := range e.Elts {
			elts[i] = renderExpr(fset, elt)
		}
		return "[" + strings.Join(elts, ", ") + "]"
	case *ast.UnaryExpr:
		return e.Op.String() + renderExpr(fset, e.X)
	case *ast.CallExpr:
		if isBuiltinTypeConversionCall(e) {
			return renderExpr(fset, e.Args[0])
		}
	case *ast.SelectorExpr:
		return exprToString(fset, expr)
	}
	return exprToString(fset, expr)
}

// renderStringLit returns the unquoted value for a string literal expression.
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

var durationUnits = map[string]string{
	"Nanosecond":  "ns",
	"Microsecond": "µs",
	"Millisecond": "ms",
	"Second":      "s",
	"Minute":      "m",
	"Hour":        "h",
}

var unitAbbreviations = map[string]string{
	"Nanosecond":  "ns",
	"Microsecond": "µs",
	"Millisecond": "ms",
	"Second":      "sec",
	"Minute":      "min",
	"Hour":        "h",
	"B":           "B",
	"KB":          "KB",
	"MB":          "MB",
	"GB":          "GB",
	"TB":          "TB",
}

// renderDuration renders duration quantity/unit arguments into display form.
func renderDuration(fset *token.FileSet, quantity, unit ast.Expr) string {
	qty := renderExpr(fset, quantity)
	unitStr := durationUnitAbbrev(unit)
	if unitStr != "" {
		if !isNonNegativeInteger(qty) {
			return qty
		}
		if qty == "0" {
			return "0"
		}
		return qty + unitStr
	}
	return qty + " " + exprToString(fset, unit)
}

// isNonNegativeInteger reports whether s is a base-10 non-negative integer.
func isNonNegativeInteger(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// durationUnitAbbrev maps known time unit selectors (for example time.Second) to suffixes.
func durationUnitAbbrev(expr ast.Expr) string {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	if abbrev, ok := durationUnits[sel.Sel.Name]; ok {
		return abbrev
	}
	return ""
}

// renderWithUnit renders (default, unit) pairs for int/int64 getter families.
func renderWithUnit(fset *token.FileSet, defaultExpr, unitExpr ast.Expr) string {
	def := renderExpr(fset, defaultExpr)
	defWrapped := def
	if !isLiteralDefault(defaultExpr) {
		defWrapped = "${" + def + "}"
	}

	innerUnit := unitExpr
	if call, ok := unitExpr.(*ast.CallExpr); ok && isBuiltinTypeConversionCall(call) {
		innerUnit = call.Args[0]
	}

	if lit, ok := innerUnit.(*ast.BasicLit); ok && lit.Kind == token.INT && lit.Value == "1" {
		return defWrapped
	}

	if sel, ok := innerUnit.(*ast.SelectorExpr); ok {
		if abbrev, ok := unitAbbreviations[sel.Sel.Name]; ok {
			return defWrapped + abbrev
		}
	}

	if isNumericLiteral(innerUnit) && isNumericLiteral(defaultExpr) {
		return def + " * " + renderExpr(fset, innerUnit)
	}

	unitStr := renderExpr(fset, innerUnit)
	if !isLiteralDefault(innerUnit) {
		unitStr = "${" + unitStr + "}"
	}
	return defWrapped + " " + unitStr
}

// isNumericLiteral reports whether expr is an int/float literal or signed literal.
func isNumericLiteral(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return e.Kind == token.INT || e.Kind == token.FLOAT
	case *ast.UnaryExpr:
		return isNumericLiteral(e.X)
	}
	return false
}

// exprToString pretty-prints an expression using go/printer.
func exprToString(fset *token.FileSet, expr ast.Expr) string {
	var buf strings.Builder
	if err := printer.Fprint(&buf, fset, expr); err != nil {
		return "<expr>"
	}
	return buf.String()
}
