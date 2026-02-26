package main

import (
	"flag"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
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
		wantKeys       []string // expected primary keys (comma-joined)
		wantDefs       []string // expected defaults (parallel to wantKeys)
		wantDesc       []string // expected descriptions
		wantGrps       []string // expected groups
		wantReloadable []bool   // expected reloadable flags (nil = all false)
	}{
		{
			name: "GetStringVar/literal default",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetStringVar("literal", "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"literal"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetStringVar/non-literal default",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	defaultString := "default"
	//cdoc:desc d
	conf.GetStringVar(defaultString, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"${defaultString}"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetStringVar/multiple keys and env key",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetStringVar("localhost:2379", "etcd.hosts", "ETCD_HOSTS")
}`,
			wantKeys: []string{"etcd.hosts,ETCD_HOSTS"},
			wantDefs: []string{"localhost:2379"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetBoolVar/true",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetBoolVar(true, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"true"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetBoolVar/false",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetBoolVar(false, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"false"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetFloat64Var/literal",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetFloat64Var(1.1, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"1.1"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetStringSliceVar/empty slice",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetStringSliceVar([]string{}, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"[]"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetIntVar/negative default",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetIntVar(-1, 1, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"-1"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetIntVar/type-converted default",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetIntVar(int(100), 1, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"100"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetStringVar/selector default",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
var defaults struct { Name string }
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetStringVar(defaults.Name, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"${defaults.Name}"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetIntVar/literal default unit=1",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetIntVar(8080, 1, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"8080"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetIntVar/non-literal default unit=1",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	defaultInt := 8080
	//cdoc:desc d
	conf.GetIntVar(defaultInt, 1, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"${defaultInt}"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetIntVar/non-literal default with cdoc:default override",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	defaultInt := 8080
	//cdoc:desc d
	//cdoc:default 12
	conf.GetIntVar(defaultInt, 1, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"12"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetIntVar/two numeric literals",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetIntVar(10, 10, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"10 * 10"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetIntVar/int(time.Second)",
			src: `package test
import (
	"time"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetIntVar(1, int(time.Second), "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"1sec"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetIntVar/int(bytesize.KB)",
			src: `package test
import (
	"github.com/rudderlabs/rudder-go-kit/bytesize"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetIntVar(1, int(bytesize.KB), "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"1KB"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetIntVar/non-literal unit",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	intScale := 10
	//cdoc:desc d
	conf.GetIntVar(1, intScale, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"1 ${intScale}"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetInt64Var/literal default unit=1",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetInt64Var(8080, 1, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"8080"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetInt64Var/non-literal default unit=1",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	defaultInt64 := int64(8080)
	//cdoc:desc d
	conf.GetInt64Var(defaultInt64, 1, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"${defaultInt64}"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetInt64Var/non-literal default with cdoc:default override",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	defaultInt64 := int64(8080)
	//cdoc:desc d
	//cdoc:default 12
	conf.GetInt64Var(defaultInt64, 1, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"12"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetInt64Var/two numeric literals",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetInt64Var(10, 10, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"10 * 10"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetInt64Var/int64(time.Second)",
			src: `package test
import (
	"time"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetInt64Var(1, int64(time.Second), "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"1sec"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetInt64Var/bytesize.KB",
			src: `package test
import (
	"github.com/rudderlabs/rudder-go-kit/bytesize"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetInt64Var(1, bytesize.KB, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"1KB"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetInt64Var/non-literal unit",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	int64Scale := int64(10)
	//cdoc:desc d
	conf.GetInt64Var(1, int64Scale, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"1 ${int64Scale}"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetDurationVar/time.Second",
			src: `package test
import (
	"time"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetDurationVar(10, time.Second, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"10s"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetDurationVar/time.Millisecond",
			src: `package test
import (
	"time"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetDurationVar(200, time.Millisecond, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"200ms"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetReloadableIntVar/unit=1",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetReloadableIntVar(100, 1, "key")
}`,
			wantKeys:       []string{"key"},
			wantDefs:       []string{"100"},
			wantDesc:       []string{"d"},
			wantReloadable: []bool{true},
		},
		{
			name: "GetReloadableDurationVar/time.Second",
			src: `package test
import (
	"time"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetReloadableDurationVar(30, time.Second, "key")
}`,
			wantKeys:       []string{"key"},
			wantDefs:       []string{"30s"},
			wantDesc:       []string{"d"},
			wantReloadable: []bool{true},
		},
		{
			name: "GetReloadableStringVar/literal",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetReloadableStringVar("val", "key")
}`,
			wantKeys:       []string{"key"},
			wantDefs:       []string{"val"},
			wantDesc:       []string{"d"},
			wantReloadable: []bool{true},
		},
		{
			name: "GetReloadableFloat64Var/literal",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetReloadableFloat64Var(0.5, "key")
}`,
			wantKeys:       []string{"key"},
			wantDefs:       []string{"0.5"},
			wantDesc:       []string{"d"},
			wantReloadable: []bool{true},
		},
		{
			name: "directive/ignore",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:ignore
	conf.GetStringVar("", "k8s.client.key")
	//cdoc:group General
	//cdoc:desc Some config
	conf.GetStringVar("val", "some.key")
}`,
			wantKeys: []string{"some.key"},
			wantDefs: []string{"val"},
			wantDesc: []string{"Some config"},
			wantGrps: []string{"General"},
		},
		{
			name: "directive/group inheritance",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:group GroupA
	//cdoc:desc First
	conf.GetStringVar("a", "key.a")
	//cdoc:desc Second
	conf.GetStringVar("b", "key.b")
	//cdoc:group GroupB
	//cdoc:desc Third
	conf.GetStringVar("c", "key.c")
}`,
			wantKeys: []string{"key.a", "key.b", "key.c"},
			wantDefs: []string{"a", "b", "c"},
			wantDesc: []string{"First", "Second", "Third"},
			wantGrps: []string{"GroupA", "GroupA", "GroupB"},
		},
		{
			name: "directive/package-level config call",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
//cdoc:group General
//cdoc:desc Number of partitions
var x = config.GetIntVar(64, 1, "partitionCount", "PARTITION_COUNT")
`,
			wantKeys: []string{"partitionCount,PARTITION_COUNT"},
			wantDefs: []string{"64"},
			wantDesc: []string{"Number of partitions"},
			wantGrps: []string{"General"},
		},
		{
			name: "directive/varkey for dynamic key",
			src: `package test
import (
	"fmt"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config, wsID string) {
	//cdoc:desc d
	//cdoc:key workspace.<id>.timeout
	conf.GetStringVar("30s", fmt.Sprintf("workspace.%s.timeout", wsID))
}`,
			wantKeys: []string{"workspace.<id>.timeout"},
			wantDefs: []string{"30s"},
			wantDesc: []string{"d"},
		},
		{
			name: "directive/varkey mixed with static keys",
			src: `package test
import (
	"fmt"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config, wsID string) {
	//cdoc:desc d
	//cdoc:key WORKSPACE_<id>_TIMEOUT
	conf.GetStringVar("30s", "workspace.timeout", fmt.Sprintf("WORKSPACE_%s_TIMEOUT", wsID))
}`,
			wantKeys: []string{"workspace.timeout,WORKSPACE_<id>_TIMEOUT"},
			wantDefs: []string{"30s"},
			wantDesc: []string{"d"},
		},
		{
			name: "directive/multiple varkeys for multiple dynamic args",
			src: `package test
import (
	"fmt"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config, wsID string) {
	//cdoc:desc d
	//cdoc:key workspace.<id>.timeout
	//cdoc:key WORKSPACE_<id>_TIMEOUT
	conf.GetStringVar("30s", fmt.Sprintf("workspace.%s.timeout", wsID), fmt.Sprintf("WORKSPACE_%s_TIMEOUT", wsID))
}`,
			wantKeys: []string{"workspace.<id>.timeout,WORKSPACE_<id>_TIMEOUT"},
			wantDefs: []string{"30s"},
			wantDesc: []string{"d"},
		},
		{
			name: "directive/vardefault for dynamic default",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config, defaultVal string) {
	//cdoc:desc d
	//cdoc:default 5s
	conf.GetStringVar(defaultVal, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"5s"},
			wantDesc: []string{"d"},
		},
		{
			name: "directive/vardefault overrides literal default",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	//cdoc:default 10s
	conf.GetStringVar("5s", "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"10s"},
			wantDesc: []string{"d"},
		},
		{
			name: "directive/deprecated non-Var methods are ignored",
			src: `package test
import (
	"time"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//cdoc:group General
	//cdoc:desc Ignored bool
	conf.GetBool("feature.enabled", false)
	//cdoc:desc Ignored duration
	conf.GetDuration("shutdown.timeout", 10, time.Second)
	//cdoc:desc Kept entry
	conf.GetStringVar("val", "kept.key")
}`,
			wantKeys: []string{"kept.key"},
			wantDefs: []string{"val"},
			wantDesc: []string{"Kept entry"},
			wantGrps: []string{"General"},
		},
		{
			name: "directive/non-reloadable and reloadable mixed",
			src: `package test
import (
	"time"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//cdoc:group HTTP
	//cdoc:desc HTTP port
	conf.GetIntVar(8080, 1, "http.port")
	//cdoc:desc Read timeout
	conf.GetReloadableDurationVar(30, time.Second, "http.readTimeout")
}`,
			wantKeys:       []string{"http.port", "http.readTimeout"},
			wantDefs:       []string{"8080", "30s"},
			wantDesc:       []string{"HTTP port", "Read timeout"},
			wantGrps:       []string{"HTTP", "HTTP"},
			wantReloadable: []bool{false, true},
		},
		{
			name: "directive/no description warns but still extracts",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:group General
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
				if tt.wantGrps != nil {
					require.Equal(t, tt.wantGrps[i], e.Group, "entry[%d] group", i)
				}
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
	//cdoc:group 2 HTTP
	//cdoc:desc Port
	conf.GetStringVar("8080", "http.port")
	//cdoc:group 1 General
	//cdoc:desc Name
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

func TestParseProject(t *testing.T) {
	entries, warnings, err := parseProject("testdata")
	require.NoError(t, err)

	// Verify expected warnings from extraction.
	requireWarningContains(t, warnings, "non-literal config key argument without //cdoc:key directive")

	// Verify that -warn would produce missing-description warnings.
	missingWarnings := generateWarnings(entries)
	requireWarningContains(t, missingWarnings, `"missingDescription" has no //cdoc:desc`)

	md := formatMarkdown(entries, "PREFIX")

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

func TestRun(t *testing.T) {
	output := filepath.Join(t.TempDir(), "output.md")
	err := run("testdata", output, "PREFIX", true)
	require.NoError(t, err)

	got, err := os.ReadFile(output)
	require.NoError(t, err)

	expected, err := os.ReadFile("testdata/expected_output.md")
	require.NoError(t, err, "reading golden file (run with -update to generate)")
	require.Equal(t, string(expected), string(got), "run output does not match golden file")
}

func requireWarningContains(t *testing.T, warnings []string, substr string) {
	t.Helper()
	for _, w := range warnings {
		if strings.Contains(w, substr) {
			return
		}
	}
	t.Errorf("expected warning containing %q, got: %v", substr, warnings)
}
