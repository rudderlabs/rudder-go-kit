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

func TestScanFile_Cases(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		src      string
		want     expectedScanResult
	}{
		{
			name: "ignore directive on same line",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	conf.GetStringVar("x", "ignored.key") //cdoc:ignore
	//cdoc:desc kept
	conf.GetStringVar("y", "kept.key")
}`,
			want: expectedScanResult{
				entries: []expectedEntry{
					{primaryKey: "kept.key", description: new("kept")},
				},
			},
		},
		{
			name: "multiline inline var keys apply to current call",
			src: `package test
import (
	"fmt"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config, wsID string) {
	//cdoc:desc d
	conf.GetStringVar("30s",
		fmt.Sprintf("workspace.%s.timeout", wsID), //cdoc:key workspace.<id>.timeout
		fmt.Sprintf("WORKSPACE_%s_TIMEOUT", wsID), //cdoc:key WORKSPACE_<id>_TIMEOUT
	)
}`,
			want: expectedScanResult{
				entries: []expectedEntry{
					{
						primaryKey:  "workspace.<id>.timeout,WORKSPACE_<id>_TIMEOUT",
						description: new("d"),
						configKeys:  stringsPtr("workspace.<id>.timeout", "WORKSPACE_<id>_TIMEOUT"),
						envKeys:     stringsPtr(),
					},
				},
			},
		},
		{
			name: "multiline inline var key does not leak to next call",
			src: `package test
import (
	"fmt"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config, wsID string) {
	conf.GetStringVar("30s",
		fmt.Sprintf("workspace.%s.timeout", wsID), //cdoc:key workspace.<id>.timeout
	)
	conf.GetStringVar("30s", fmt.Sprintf("workspace.%s.retry", wsID))
}`,
			want: expectedScanResult{
				entries: []expectedEntry{
					{primaryKey: "workspace.<id>.timeout", configKeys: stringsPtr("workspace.<id>.timeout")},
					{primaryKey: ""},
				},
				warningCode: []model.WarningCode{
					model.WarningCodeDynamicKeyMissing,
				},
			},
		},
		{
			name: "uses nearest description",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc first
	//cdoc:desc second
	conf.GetStringVar("y", "kept.key")
}`,
			want: expectedScanResult{
				entries: []expectedEntry{
					{primaryKey: "kept.key", description: new("second"), configKeys: stringsPtr("kept.key")},
				},
				warningCode: []model.WarningCode{
					model.WarningCodeUnusedDescDirective,
				},
			},
		},
		{
			name:     "group order declarations without calls",
			fileName: "order.go",
			src: `package test
//cdoc:group 2 HTTP
//cdoc:group 1 General`,
			want: expectedScanResult{
				groupOrders: []expectedGroupOrder{
					{group: "HTTP", order: 2},
					{group: "General", order: 1},
				},
			},
		},
		{
			name: "warning codes",
			src: `package test
import (
	"fmt"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config, wsID string) {
	//cdoc:key unused.key
	conf.GetStringVar("ok", "static.key")
	conf.GetStringVar("x", fmt.Sprintf("dynamic.%s", wsID))
}`,
			want: expectedScanResult{
				entries: []expectedEntry{
					{primaryKey: "static.key"},
					{primaryKey: ""},
				},
				warningCode: []model.WarningCode{
					model.WarningCodeUnusedKeyOverride,
					model.WarningCodeDynamicKeyMissing,
				},
			},
		},
		{
			name: "directives do not leak after invalid call",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc should-not-leak
	//cdoc:key leaked.key
	conf.GetStringVar("only-default")
	conf.GetStringVar("ok", "real.key")
}`,
			want: expectedScanResult{
				entries: []expectedEntry{
					{primaryKey: "real.key", description: new(""), configKeys: stringsPtr("real.key")},
				},
				warningCode: []model.WarningCode{
					model.WarningCodeArgCount,
				},
			},
		},
		{
			name: "ast-only method matching",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
type fake struct{}
func (fake) GetStringVar(_ string, _ ...string) {}
func f(conf *config.Config, fk fake) {
	//cdoc:desc real
	conf.GetStringVar("ok", "real.key")
	//cdoc:desc fake
	fk.GetStringVar("bad", "fake.key")
}`,
			want: expectedScanResult{
				entries: []expectedEntry{
					{primaryKey: "real.key", description: new("real"), configKeys: stringsPtr("real.key")},
					{primaryKey: "fake.key", description: new("fake"), configKeys: stringsPtr("fake.key")},
				},
			},
		},
		{
			name: "unused default directive warning",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:default 10
}`,
			want: expectedScanResult{
				warningCode: []model.WarningCode{
					model.WarningCodeUnusedDefaultDirective,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileName := tt.fileName
			if fileName == "" {
				fileName = "test.go"
			}
			result := scanSource(t, fileName, tt.src)
			assertScanResult(t, result, tt.want)
		})
	}
}

func TestParseGroupDirective(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantGroup string
		wantOrder int
	}{
		{
			name:      "trims spacing",
			value:     "  2    HTTP   Server  ",
			wantGroup: "HTTP Server",
			wantOrder: 2,
		},
		{
			name:      "no numeric order",
			value:     "General Config",
			wantGroup: "General Config",
			wantOrder: 0,
		},
		{
			name:      "empty",
			value:     "  ",
			wantGroup: "",
			wantOrder: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, order := parseGroupDirective(tt.value)
			require.Equal(t, tt.wantOrder, order)
			require.Equal(t, tt.wantGroup, group)
		})
	}
}

func TestIsEnvVarStyle(t *testing.T) {
	tests := []struct {
		key     string
		wantEnv bool
	}{
		{"http.port", false},
		{"ETCD_HOSTS", true},
		{"RELEASE_NAME", true},
		{"K8S_IN_CLUSTER", true},
		{"gatewaySeparateService", false},
		{"newworkspace.poller.baseUrl", false},
		{"KUBECONFIG", true},
		{"KUBE_NAMESPACE", true},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			require.Equal(t, tt.wantEnv, IsEnvVarStyle(tt.key))
		})
	}
}

func assertScanResult(t *testing.T, got FileScanResult, want expectedScanResult) {
	t.Helper()
	require.Len(t, got.Entries, len(want.entries), "entry count")
	for i, expected := range want.entries {
		require.Equal(t, expected.primaryKey, got.Entries[i].PrimaryKey, "entry[%d].primaryKey", i)
		if expected.description != nil {
			require.Equal(t, *expected.description, got.Entries[i].Description, "entry[%d].description", i)
		}
		if expected.configKeys != nil {
			require.Equal(t, *expected.configKeys, got.Entries[i].ConfigKeys, "entry[%d].configKeys", i)
		}
		if expected.envKeys != nil {
			require.Equal(t, *expected.envKeys, got.Entries[i].EnvKeys, "entry[%d].envKeys", i)
		}
	}

	require.Len(t, got.Warnings, len(want.warningCode), "warning count")
	for i, code := range want.warningCode {
		require.Equal(t, code, got.Warnings[i].Code, "warning[%d].code", i)
	}

	require.Len(t, got.GroupOrderDeclarations, len(want.groupOrders), "group declaration count")
	for i, declaration := range want.groupOrders {
		require.Equal(t, declaration.group, got.GroupOrderDeclarations[i].Group, "groupDeclaration[%d].group", i)
		require.Equal(t, declaration.order, got.GroupOrderDeclarations[i].Order, "groupDeclaration[%d].order", i)
	}
}

func scanSource(t *testing.T, fileName, src string) FileScanResult {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, fileName, src, parser.ParseComments)
	require.NoError(t, err)
	return ScanFile(fset, file, fileName, ScanOptions{})
}

type expectedEntry struct {
	primaryKey  string
	description *string
	configKeys  *[]string
	envKeys     *[]string
}

type expectedGroupOrder struct {
	group string
	order int
}

type expectedScanResult struct {
	entries     []expectedEntry
	warningCode []model.WarningCode
	groupOrders []expectedGroupOrder
}

func stringsPtr(values ...string) *[]string {
	v := append([]string(nil), values...)
	return &v
}
