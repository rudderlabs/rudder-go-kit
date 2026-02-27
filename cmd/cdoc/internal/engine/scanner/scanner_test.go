package scanner

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/model"
)

func TestClassifyMethod(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		wantOK     bool
		wantFamily methodFamily
		wantReload bool
	}{
		{name: "simple", method: "GetStringVar", wantOK: true, wantFamily: familySimple},
		{name: "reloadable duration", method: "GetReloadableDurationVar", wantOK: true, wantFamily: familyDuration, wantReload: true},
		{name: "multiplier", method: "GetInt64Var", wantOK: true, wantFamily: familyWithMultiplier},
		{name: "deprecated non-var", method: "GetBool", wantOK: false},
		{name: "unknown family", method: "GetMapVar", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, ok := classifyMethod(tt.method)
			require.Equal(t, tt.wantOK, ok)
			if !tt.wantOK {
				return
			}
			require.Equal(t, tt.wantFamily, spec.family)
			require.Equal(t, tt.wantReload, spec.reloadable)
		})
	}
}

func TestExtractFromFile_IgnoreDirectiveOnSameLine(t *testing.T) {
	src := `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	conf.GetStringVar("x", "ignored.key") //cdoc:ignore
	//cdoc:desc kept
	conf.GetStringVar("y", "kept.key")
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	entries, warnings := ExtractFromFile(fset, file, "test.go")
	require.Empty(t, warnings)
	require.Len(t, entries, 1)
	require.Equal(t, "kept.key", entries[0].PrimaryKey)
}

func TestExtractFromFile_UsesNearestDescription(t *testing.T) {
	src := `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc first
	//cdoc:desc second
	conf.GetStringVar("y", "kept.key")
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	entries, warnings := ExtractFromFile(fset, file, "test.go")
	require.Len(t, entries, 1)
	require.Equal(t, "second", entries[0].Description)
	require.Len(t, warnings, 1)
	require.Equal(t, model.WarningCodeUnusedDescDirective, warnings[0].Code)
}

func TestScanFile_GroupOrderDeclarationsWithoutCalls(t *testing.T) {
	src := `package test
//cdoc:group 2 HTTP
//cdoc:group 1 General`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "order.go", src, parser.ParseComments)
	require.NoError(t, err)

	result := ScanFile(fset, file, "order.go")
	require.Empty(t, result.Entries)
	require.Empty(t, result.Warnings)
	require.Len(t, result.GroupOrderDeclarations, 2)
	require.Equal(t, "HTTP", result.GroupOrderDeclarations[0].Group)
	require.Equal(t, 2, result.GroupOrderDeclarations[0].Order)
	require.Equal(t, "General", result.GroupOrderDeclarations[1].Group)
	require.Equal(t, 1, result.GroupOrderDeclarations[1].Order)
}

func TestExtractFromFile_WarningCodes(t *testing.T) {
	src := `package test
import (
	"fmt"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config, wsID string) {
	//cdoc:key unused.key
	conf.GetStringVar("ok", "static.key")
	conf.GetStringVar("x", fmt.Sprintf("dynamic.%s", wsID))
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	_, warnings := ExtractFromFile(fset, file, "test.go")
	require.Len(t, warnings, 2)
	require.Equal(t, model.WarningCodeUnusedKeyOverride, warnings[0].Code)
	require.Equal(t, model.WarningCodeDynamicKeyMissing, warnings[1].Code)
}

func TestExtractFromFile_DirectivesDoNotLeakAfterInvalidCall(t *testing.T) {
	src := `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc should-not-leak
	//cdoc:key leaked.key
	conf.GetStringVar("only-default")
	conf.GetStringVar("ok", "real.key")
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	entries, warnings := ExtractFromFile(fset, file, "test.go")
	require.Len(t, entries, 1)
	require.Equal(t, "real.key", entries[0].PrimaryKey)
	require.Equal(t, "", entries[0].Description, "description from failed call must not be reused")

	require.Len(t, warnings, 1)
	require.Equal(t, model.WarningCodeArgCount, warnings[0].Code)
}

func TestScanFile_ASTOnlyMethodMatching(t *testing.T) {
	src := `package test
import "github.com/rudderlabs/rudder-go-kit/config"
type fake struct{}
func (fake) GetStringVar(_ string, _ ...string) {}
func f(conf *config.Config, fk fake) {
	//cdoc:desc real
	conf.GetStringVar("ok", "real.key")
	//cdoc:desc fake
	fk.GetStringVar("bad", "fake.key")
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	result := ScanFile(fset, file, "test.go")
	require.Len(t, result.Entries, 2)
	require.Equal(t, "real.key", result.Entries[0].PrimaryKey)
	require.Equal(t, "fake.key", result.Entries[1].PrimaryKey)
	require.Equal(t, "real", result.Entries[0].Description)
	require.Equal(t, "fake", result.Entries[1].Description)
	require.Empty(t, result.Warnings)
}

func TestExtractFromFile_UnusedDefaultDirectiveWarning(t *testing.T) {
	src := `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:default 10
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	_, warnings := ExtractFromFile(fset, file, "test.go")
	require.Len(t, warnings, 1)
	require.Equal(t, model.WarningCodeUnusedDefaultDirective, warnings[0].Code)
}

func TestParseGroupDirective_TrimsSpacing(t *testing.T) {
	group, order := parseGroupDirective("  2    HTTP   Server  ")
	require.Equal(t, 2, order)
	require.Equal(t, "HTTP Server", group)
}
