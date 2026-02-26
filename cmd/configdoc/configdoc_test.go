package main

import (
	"flag"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/config"
)

var update = flag.Bool("update", false, "update golden files")

func TestExtractFromFile(t *testing.T) {
	tests := []struct {
		name           string
		src            string
		wantKeys       []string // expected primary keys
		wantDefs       []string // expected defaults (parallel to wantKeys)
		wantDesc       []string // expected descriptions
		wantGrps       []string // expected groups
		wantReloadable []bool   // expected reloadable flags (nil = all false)
	}{
		{
			name: "simple string var",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//configdoc:group General
	//configdoc:description Comma-separated etcd endpoints
	conf.GetStringVar("localhost:2379", "etcd.hosts", "ETCD_HOSTS")
}`,
			wantKeys: []string{"etcd.hosts"},
			wantDefs: []string{"localhost:2379"},
			wantDesc: []string{"Comma-separated etcd endpoints"},
			wantGrps: []string{"General"},
		},
		{
			name: "int var with min",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//configdoc:group HTTP
	//configdoc:description HTTP server port
	conf.GetIntVar(8080, 1, "http.port")
}`,
			wantKeys: []string{"http.port"},
			wantDefs: []string{"8080"},
			wantDesc: []string{"HTTP server port"},
			wantGrps: []string{"HTTP"},
		},
		{
			name: "duration var",
			src: `package test
import (
	"time"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//configdoc:group HTTP
	//configdoc:description Read header timeout
	conf.GetDurationVar(10, time.Second, "http.timeout")
}`,
			wantKeys: []string{"http.timeout"},
			wantDefs: []string{"10s"},
			wantDesc: []string{"Read header timeout"},
			wantGrps: []string{"HTTP"},
		},
		{
			name: "bool var",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//configdoc:group Migration
	//configdoc:description Whether gateway runs as separate service
	conf.GetBoolVar(true, "gatewaySeparateService")
}`,
			wantKeys: []string{"gatewaySeparateService"},
			wantDefs: []string{"true"},
			wantDesc: []string{"Whether gateway runs as separate service"},
			wantGrps: []string{"Migration"},
		},
		{
			name: "ignore directive",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//configdoc:ignore
	conf.GetStringVar("", "k8s.client.key")
	//configdoc:group General
	//configdoc:description Some config
	conf.GetStringVar("val", "some.key")
}`,
			wantKeys: []string{"some.key"},
			wantDefs: []string{"val"},
			wantDesc: []string{"Some config"},
			wantGrps: []string{"General"},
		},
		{
			name: "group inheritance",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//configdoc:group GroupA
	//configdoc:description First
	conf.GetStringVar("a", "key.a")
	//configdoc:description Second
	conf.GetStringVar("b", "key.b")
	//configdoc:group GroupB
	//configdoc:description Third
	conf.GetStringVar("c", "key.c")
}`,
			wantKeys: []string{"key.a", "key.b", "key.c"},
			wantDefs: []string{"a", "b", "c"},
			wantDesc: []string{"First", "Second", "Third"},
			wantGrps: []string{"GroupA", "GroupA", "GroupB"},
		},
		{
			name: "float64 var",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//configdoc:group Retry
	//configdoc:description Randomization factor
	conf.GetFloat64Var(0.5, "k8s.client.retry.randomizationFactor")
}`,
			wantKeys: []string{"k8s.client.retry.randomizationFactor"},
			wantDefs: []string{"0.5"},
			wantDesc: []string{"Randomization factor"},
			wantGrps: []string{"Retry"},
		},
		{
			name: "package-level config call",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
//configdoc:group General
//configdoc:description Number of partitions
var x = config.GetIntVar(64, 1, "partitionCount", "PARTITION_COUNT")
`,
			wantKeys: []string{"partitionCount"},
			wantDefs: []string{"64"},
			wantDesc: []string{"Number of partitions"},
			wantGrps: []string{"General"},
		},
		{
			name: "duration with milliseconds",
			src: `package test
import (
	"time"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//configdoc:group K8s
	//configdoc:description Initial retry interval
	conf.GetDurationVar(200, time.Millisecond, "k8s.client.retry.initialInterval")
}`,
			wantKeys: []string{"k8s.client.retry.initialInterval"},
			wantDefs: []string{"200ms"},
			wantDesc: []string{"Initial retry interval"},
			wantGrps: []string{"K8s"},
		},
		{
			name: "varkey for dynamic key",
			src: `package test
import (
	"fmt"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config, wsID string) {
	//configdoc:group Workspace
	//configdoc:description Per-workspace timeout
	//configdoc:varkey workspace.<id>.timeout
	conf.GetStringVar("30s", fmt.Sprintf("workspace.%s.timeout", wsID))
}`,
			wantKeys: []string{"workspace.<id>.timeout"},
			wantDefs: []string{"30s"},
			wantDesc: []string{"Per-workspace timeout"},
			wantGrps: []string{"Workspace"},
		},
		{
			name: "varkey mixed with static keys",
			src: `package test
import (
	"fmt"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config, wsID string) {
	//configdoc:group Workspace
	//configdoc:description Workspace timeout
	//configdoc:varkey WORKSPACE_<id>_TIMEOUT
	conf.GetStringVar("30s", "workspace.timeout", fmt.Sprintf("WORKSPACE_%s_TIMEOUT", wsID))
}`,
			wantKeys: []string{"workspace.timeout"},
			wantDefs: []string{"30s"},
			wantDesc: []string{"Workspace timeout"},
			wantGrps: []string{"Workspace"},
		},
		{
			name: "multiple varkeys for multiple dynamic args",
			src: `package test
import (
	"fmt"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config, wsID string) {
	//configdoc:group Workspace
	//configdoc:description Workspace timeout
	//configdoc:varkey workspace.<id>.timeout
	//configdoc:varkey WORKSPACE_<id>_TIMEOUT
	conf.GetStringVar("30s", fmt.Sprintf("workspace.%s.timeout", wsID), fmt.Sprintf("WORKSPACE_%s_TIMEOUT", wsID))
}`,
			wantKeys: []string{"workspace.<id>.timeout"},
			wantDefs: []string{"30s"},
			wantDesc: []string{"Workspace timeout"},
			wantGrps: []string{"Workspace"},
		},
		{
			name: "vardefault for dynamic default",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config, defaultVal string) {
	//configdoc:group General
	//configdoc:description Retry interval
	//configdoc:vardefault 5s
	conf.GetStringVar(defaultVal, "retry.interval")
}`,
			wantKeys: []string{"retry.interval"},
			wantDefs: []string{"5s"},
			wantDesc: []string{"Retry interval"},
			wantGrps: []string{"General"},
		},
		{
			name: "vardefault not applied when default is literal",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//configdoc:group General
	//configdoc:description Retry interval
	//configdoc:vardefault 10s
	conf.GetStringVar("5s", "retry.interval")
}`,
			wantKeys: []string{"retry.interval"},
			wantDefs: []string{"10s"},
			wantDesc: []string{"Retry interval"},
			wantGrps: []string{"General"},
		},
		{
			name: "deprecated non-Var methods are ignored",
			src: `package test
import (
	"time"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//configdoc:group General
	//configdoc:description Ignored bool
	conf.GetBool("feature.enabled", false)
	//configdoc:description Ignored duration
	conf.GetDuration("shutdown.timeout", 10, time.Second)
	//configdoc:description Kept entry
	conf.GetStringVar("val", "kept.key")
}`,
			wantKeys: []string{"kept.key"},
			wantDefs: []string{"val"},
			wantDesc: []string{"Kept entry"},
			wantGrps: []string{"General"},
		},
		{
			name: "reloadable var",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//configdoc:group HTTP
	//configdoc:description Maximum connections
	conf.GetReloadableIntVar(100, 1, "http.maxConns")
}`,
			wantKeys:       []string{"http.maxConns"},
			wantDefs:       []string{"100"},
			wantDesc:       []string{"Maximum connections"},
			wantGrps:       []string{"HTTP"},
			wantReloadable: []bool{true},
		},
		{
			name: "non-reloadable and reloadable mixed",
			src: `package test
import (
	"time"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//configdoc:group HTTP
	//configdoc:description HTTP port
	conf.GetIntVar(8080, 1, "http.port")
	//configdoc:description Read timeout
	conf.GetReloadableDurationVar(30, time.Second, "http.readTimeout")
}`,
			wantKeys:       []string{"http.port", "http.readTimeout"},
			wantDefs:       []string{"8080", "30s"},
			wantDesc:       []string{"HTTP port", "Read timeout"},
			wantGrps:       []string{"HTTP", "HTTP"},
			wantReloadable: []bool{false, true},
		},
		{
			name: "no description warns but still extracts",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//configdoc:group General
	conf.GetStringVar("default", "some.key")
}`,
			wantKeys: []string{"some.key"},
			wantDefs: []string{"default"},
			wantDesc: []string{""},
			wantGrps: []string{"General"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.src, parser.ParseComments)
			require.NoError(t, err, "parse error")

			entries, warnings := extractFromFile(fset, file, "test.go")
			for _, w := range warnings {
				t.Logf("warning: %s", w)
			}

			require.Len(t, entries, len(tt.wantKeys))

			for i, e := range entries {
				require.Equal(t, tt.wantKeys[i], e.PrimaryKey, "entry[%d] primary key", i)
				require.Equal(t, tt.wantDefs[i], e.Default, "entry[%d] default", i)
				require.Equal(t, tt.wantDesc[i], e.Description, "entry[%d] description", i)
				require.Equal(t, tt.wantGrps[i], e.Group, "entry[%d] group", i)
				if tt.wantReloadable != nil {
					require.Equal(t, tt.wantReloadable[i], e.Reloadable, "entry[%d] reloadable", i)
				}
			}
		})
	}
}

func TestKeyClassification(t *testing.T) {
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
			require.Equal(t, tt.wantEnv, isEnvVarStyle(tt.key))
		})
	}
}

func TestConfigKeyToEnv(t *testing.T) {
	tests := []struct {
		configKey string
		want      string
	}{
		{"http.port", "PREFIX_HTTP_PORT"},
		{"k8s.client.retry.initialInterval", "PREFIX_K8S_CLIENT_RETRY_INITIAL_INTERVAL"},
		{"gatewaySeparateService", "PREFIX_GATEWAY_SEPARATE_SERVICE"},
		{"partitionCount", "PREFIX_PARTITION_COUNT"},
		{"http.healthCheckTimeout", "PREFIX_HTTP_HEALTH_CHECK_TIMEOUT"},
		{"newworkspace.poller.baseUrl", "PREFIX_NEWWORKSPACE_POLLER_BASE_URL"},
	}
	for _, tt := range tests {
		t.Run(tt.configKey, func(t *testing.T) {
			require.Equal(t, tt.want, config.ConfigKeyToEnv("PREFIX", tt.configKey))
		})
	}
}

func TestDeduplication(t *testing.T) {
	entries := []configEntry{
		{PrimaryKey: "http.port", ConfigKeys: []string{"http.port"}, Default: "8080", Description: "HTTP port", Group: "HTTP"},
		{PrimaryKey: "http.port", ConfigKeys: []string{"http.port"}, Default: "8080", Description: "", Group: ""},
		{PrimaryKey: "other.key", ConfigKeys: []string{"other.key"}, Default: "val", Description: "Other", Group: "General"},
	}

	result, warnings := deduplicateEntries(entries)
	require.Len(t, result, 2)
	require.Empty(t, warnings)
	require.Equal(t, "HTTP port", result[0].Description)
	require.Equal(t, "HTTP", result[0].Group)
}

func TestDeduplicationConflictWarnings(t *testing.T) {
	entries := []configEntry{
		{PrimaryKey: "key", ConfigKeys: []string{"key"}, Default: "a", Group: "G1"},
		{PrimaryKey: "key", ConfigKeys: []string{"key"}, Default: "b", Group: "G2"},
	}

	_, warnings := deduplicateEntries(entries)
	require.Len(t, warnings, 2, "expected default + group conflict warnings")
}

func TestGroupOrder(t *testing.T) {
	entries := []configEntry{
		{PrimaryKey: "z.key", ConfigKeys: []string{"z.key"}, Default: "z", Group: "Zebra", GroupOrder: 3},
		{PrimaryKey: "a.key", ConfigKeys: []string{"a.key"}, Default: "a", Group: "Alpha", GroupOrder: 1},
		{PrimaryKey: "m.key", ConfigKeys: []string{"m.key"}, Default: "m", Group: "Middle", GroupOrder: 2},
		{PrimaryKey: "u.key", ConfigKeys: []string{"u.key"}, Default: "u", Group: "Unordered"},
		{PrimaryKey: "v.key", ConfigKeys: []string{"v.key"}, Default: "v", Group: "Another Unordered"},
	}

	md := formatMarkdown(entries, "PREFIX")

	alphaIdx := strings.Index(md, "## Alpha")
	middleIdx := strings.Index(md, "## Middle")
	zebraIdx := strings.Index(md, "## Zebra")
	unorderedIdx := strings.Index(md, "## Unordered")
	anotherIdx := strings.Index(md, "## Another Unordered")

	require.Less(t, alphaIdx, middleIdx, "Alpha should come before Middle")
	require.Less(t, middleIdx, zebraIdx, "Middle should come before Zebra")
	require.Less(t, zebraIdx, unorderedIdx, "Zebra should come before Unordered")
	require.Less(t, zebraIdx, anotherIdx, "Zebra should come before Another Unordered")
	require.Less(t, anotherIdx, unorderedIdx, "Another Unordered should come before Unordered (alphabetical)")
}

func TestGroupOrderExtraction(t *testing.T) {
	src := `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//configdoc:group 2 HTTP
	//configdoc:description Port
	conf.GetStringVar("8080", "http.port")
	//configdoc:group 1 General
	//configdoc:description Name
	conf.GetStringVar("app", "app.name")
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err, "parse error")

	entries, _ := extractFromFile(fset, file, "test.go")
	require.Len(t, entries, 2)
	require.Equal(t, 2, entries[0].GroupOrder)
	require.Equal(t, 1, entries[1].GroupOrder)
}

func TestFormatMarkdown(t *testing.T) {
	entries := []configEntry{
		{PrimaryKey: "http.port", ConfigKeys: []string{"http.port"}, Default: "8080", Description: "HTTP server port", Group: "HTTP server"},
		{PrimaryKey: "etcd.hosts", ConfigKeys: []string{"etcd.hosts"}, EnvKeys: []string{"ETCD_HOSTS"}, Default: "localhost:2379", Description: "Etcd endpoints", Group: "General"},
		{PrimaryKey: "http.maxConns", ConfigKeys: []string{"http.maxConns"}, Default: "100", Description: "Max connections", Reloadable: true, Group: "HTTP server"},
	}

	md := formatMarkdown(entries, "PREFIX")

	require.Contains(t, md, "## HTTP server")
	require.Contains(t, md, "## General")
	require.Contains(t, md, "`http.port`")
	require.Contains(t, md, "`PREFIX_HTTP_PORT`")
	require.Contains(t, md, "`ETCD_HOSTS`")
	require.Contains(t, md, "auto-generated")
	require.Contains(t, md, "`PREFIX_`")
	// Reloadable entry should have emoji prefix.
	require.Contains(t, md, "🔄 Max connections")
	// Non-reloadable entry should not have emoji.
	require.NotContains(t, md, "🔄 HTTP server port")
}

func TestGoldenOutput(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "testdata/example.go", nil, parser.ParseComments)
	require.NoError(t, err, "parse error")

	rawEntries, warnings := extractFromFile(fset, file, "testdata/example.go")

	// Verify expected warnings from extraction.
	wantWarningSubstrings := []string{
		"non-literal config key argument without //configdoc:varkey directive",
	}
	for _, want := range wantWarningSubstrings {
		found := false
		for _, w := range warnings {
			if strings.Contains(w, want) {
				found = true
				break
			}
		}
		require.True(t, found, "expected warning containing %q, got: %v", want, warnings)
	}

	// Run deduplication (mirrors the real pipeline, filters entries with no keys).
	entries, _ := deduplicateEntries(rawEntries)

	md := formatMarkdown(entries, "PREFIX")

	// Verify that -warn would produce missing-description warnings.
	missingWarnings := generateWarnings(entries)
	wantMissingSubstrings := []string{
		`"missingDescription" has no //configdoc:description`,
	}
	for _, want := range wantMissingSubstrings {
		found := false
		for _, w := range missingWarnings {
			if strings.Contains(w, want) {
				found = true
				break
			}
		}
		require.True(t, found, "expected missing-description warning containing %q, got: %v", want, missingWarnings)
	}

	goldenPath := "testdata/expected_output.md"
	if *update {
		err := os.WriteFile(goldenPath, []byte(md), 0o644)
		require.NoError(t, err, "updating golden file")
		t.Log("updated golden file")
		return
	}

	expected, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "reading golden file (run with -update to generate)")

	require.Equal(t, string(expected), md, "output does not match golden file")
}
