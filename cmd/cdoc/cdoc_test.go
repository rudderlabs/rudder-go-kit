package main

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	cdoc "github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/cdoc"
	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/cdoc/model"
	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/cdoc/render"
	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/cdoc/scanner"
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
		wantConfigKeys [][]string
		wantEnvKeys    [][]string // nil row means: no env keys expected for that entry
		wantWarnings   []string   // expected warning substrings
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
			wantConfigKeys: [][]string{
				{"etcd.hosts"},
			},
			wantEnvKeys: [][]string{
				{"ETCD_HOSTS"},
			},
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
			name: "GetDurationVar/non-literal quantity with known unit",
			src: `package test
import (
	"time"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	defaultTimeout := baseTimeout
	//cdoc:desc d
	conf.GetDurationVar(defaultTimeout, time.Second, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"${defaultTimeout}"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetDurationVar/zero quantity with known unit",
			src: `package test
import (
	"time"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetDurationVar(0, time.Second, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"0"},
			wantDesc: []string{"d"},
		},
		{
			name: "GetDurationVar/unknown unit expression",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	customUnit := scale
	//cdoc:desc d
	conf.GetDurationVar(5, customUnit, "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"5 customUnit"},
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
			name: "GetReloadableBoolVar/true",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetReloadableBoolVar(true, "key")
}`,
			wantKeys:       []string{"key"},
			wantDefs:       []string{"true"},
			wantDesc:       []string{"d"},
			wantReloadable: []bool{true},
		},
		{
			name: "GetReloadableBoolVar/false",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetReloadableBoolVar(false, "key")
}`,
			wantKeys:       []string{"key"},
			wantDefs:       []string{"false"},
			wantDesc:       []string{"d"},
			wantReloadable: []bool{true},
		},
		{
			name: "GetReloadableStringSliceVar/empty slice",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetReloadableStringSliceVar([]string{}, "key")
}`,
			wantKeys:       []string{"key"},
			wantDefs:       []string{"[]"},
			wantDesc:       []string{"d"},
			wantReloadable: []bool{true},
		},
		{
			name: "GetReloadableStringSliceVar/non-empty slice",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetReloadableStringSliceVar([]string{"a", "b"}, "key")
}`,
			wantKeys:       []string{"key"},
			wantDefs:       []string{"[a, b]"},
			wantDesc:       []string{"d"},
			wantReloadable: []bool{true},
		},
		{
			name: "GetReloadableInt64Var/unit=1",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetReloadableInt64Var(500, 1, "key")
}`,
			wantKeys:       []string{"key"},
			wantDefs:       []string{"500"},
			wantDesc:       []string{"d"},
			wantReloadable: []bool{true},
		},
		{
			name: "GetReloadableInt64Var/bytesize.MB",
			src: `package test
import (
	"github.com/rudderlabs/rudder-go-kit/bytesize"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetReloadableInt64Var(500, bytesize.MB, "key")
}`,
			wantKeys:       []string{"key"},
			wantDefs:       []string{"500MB"},
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
			name: "directive/group does not apply to previous call",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc Before group
	conf.GetStringVar("a", "same.key")
	//cdoc:group General
	//cdoc:desc After group
	conf.GetStringVar("b", "same.key")
}`,
			wantKeys: []string{"same.key", "same.key"},
			wantDefs: []string{"a", "b"},
			wantDesc: []string{"Before group", "After group"},
			wantGrps: []string{"", "General"},
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
			wantConfigKeys: [][]string{
				{"workspace.<id>.timeout"},
			},
			wantEnvKeys: [][]string{
				nil,
			},
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
			wantConfigKeys: [][]string{
				{"workspace.timeout", "WORKSPACE_<id>_TIMEOUT"},
			},
			wantEnvKeys: [][]string{
				nil,
			},
		},
		{
			name: "directive/varkey variadic keys with comma-separated directive",
			src: `package test
import (
	"fmt"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config, wsID string) {
	keys := []string{fmt.Sprintf("workspace.%s.timeout", wsID), "workspace.timeout"}
	//cdoc:desc d
	//cdoc:key workspace.<id>.timeout, workspace.timeout
	conf.GetStringVar("30s", keys...)
}`,
			wantKeys: []string{"workspace.<id>.timeout,workspace.timeout"},
			wantDefs: []string{"30s"},
			wantDesc: []string{"d"},
			wantConfigKeys: [][]string{
				{"workspace.<id>.timeout", "workspace.timeout"},
			},
			wantEnvKeys: [][]string{
				nil,
			},
		},
		{
			name: "directive/varkey variadic keys with two key directives",
			src: `package test
import (
	"fmt"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config, wsID string) {
	keys := []string{fmt.Sprintf("workspace.%s.timeout", wsID), fmt.Sprintf("WORKSPACE_%s_TIMEOUT", wsID)}
	//cdoc:desc d
	//cdoc:key workspace.<id>.timeout
	//cdoc:key WORKSPACE_<id>_TIMEOUT
	conf.GetStringVar("30s", keys...)
}`,
			wantKeys: []string{"workspace.<id>.timeout,WORKSPACE_<id>_TIMEOUT"},
			wantDefs: []string{"30s"},
			wantDesc: []string{"d"},
			wantConfigKeys: [][]string{
				{"workspace.<id>.timeout", "WORKSPACE_<id>_TIMEOUT"},
			},
			wantEnvKeys: [][]string{
				nil,
			},
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
			wantConfigKeys: [][]string{
				{"workspace.<id>.timeout", "WORKSPACE_<id>_TIMEOUT"},
			},
			wantEnvKeys: [][]string{
				nil,
			},
		},
		{
			name: "directive/varkey unused override warns",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	//cdoc:key workspace.<id>.timeout
	conf.GetStringVar("30s", "workspace.timeout")
}`,
			wantKeys: []string{"workspace.timeout"},
			wantDefs: []string{"30s"},
			wantDesc: []string{"d"},
			wantWarnings: []string{
				"unused //cdoc:key override(s): workspace.<id>.timeout",
			},
		},
		{
			name: "directive/varkey extra override warns",
			src: `package test
import (
	"fmt"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config, wsID string) {
	//cdoc:desc d
	//cdoc:key workspace.<id>.timeout, workspace.timeout
	conf.GetStringVar("30s", fmt.Sprintf("workspace.%s.timeout", wsID))
}`,
			wantKeys: []string{"workspace.<id>.timeout"},
			wantDefs: []string{"30s"},
			wantDesc: []string{"d"},
			wantWarnings: []string{
				"unused //cdoc:key override(s): workspace.timeout",
			},
		},
		{
			name: "warning/non-literal key without cdoc:key directive",
			src: `package test
import (
	"fmt"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config, wsID string) {
	//cdoc:desc d
	conf.GetStringVar("30s", "workspace.timeout", fmt.Sprintf("workspace.%s.timeout", wsID))
}`,
			wantKeys: []string{"workspace.timeout"},
			wantDefs: []string{"30s"},
			wantDesc: []string{"d"},
			wantWarnings: []string{
				"non-literal config key argument without //cdoc:key directive",
			},
		},
		{
			name: "warning/simple family with too few args",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetStringVar("default")
}`,
			wantKeys: []string{},
			wantDefs: []string{},
			wantDesc: []string{},
			wantWarnings: []string{
				"expected at least 2 args, got 1",
			},
		},
		{
			name: "warning/duration family with too few args",
			src: `package test
import (
	"time"
	"github.com/rudderlabs/rudder-go-kit/config"
)
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetDurationVar(10, time.Second)
}`,
			wantKeys: []string{},
			wantDefs: []string{},
			wantDesc: []string{},
			wantWarnings: []string{
				"expected at least 3 args, got 2",
			},
		},
		{
			name: "warning/with-multiplier family with too few args",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetIntVar(10, 1)
}`,
			wantKeys: []string{},
			wantDefs: []string{},
			wantDesc: []string{},
			wantWarnings: []string{
				"expected at least 3 args, got 2",
			},
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
			wantWarnings: []string{
				"unused //cdoc:desc directive",
				"unused //cdoc:desc directive",
			},
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

	for i := range tests {
		if tests[i].wantConfigKeys == nil || tests[i].wantEnvKeys == nil {
			derivedConfigKeys, derivedEnvKeys := deriveKeyClassification(tests[i].wantKeys)
			if tests[i].wantConfigKeys == nil {
				tests[i].wantConfigKeys = derivedConfigKeys
			}
			if tests[i].wantEnvKeys == nil {
				tests[i].wantEnvKeys = derivedEnvKeys
			}
		}
		if tests[i].wantWarnings == nil {
			tests[i].wantWarnings = []string{}
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.src, parser.ParseComments)
			require.NoError(t, err, "parse error")

			entries, warnings := extractFromFile(fset, file, "test.go")
			require.Len(t, warnings, len(tt.wantWarnings), "warning count")
			for _, wantWarning := range tt.wantWarnings {
				requireWarningContains(t, warnings, wantWarning)
			}

			require.Len(t, entries, len(tt.wantKeys))
			require.Len(t, tt.wantConfigKeys, len(entries), "wantConfigKeys length")
			require.Len(t, tt.wantEnvKeys, len(entries), "wantEnvKeys length")

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
				require.Equal(t, tt.wantConfigKeys[i], e.ConfigKeys, "entry[%d] config keys", i)
				require.Equal(t, tt.wantEnvKeys[i], e.EnvKeys, "entry[%d] env keys", i)
			}
		})
	}
}

func deriveKeyClassification(primaryKeys []string) ([][]string, [][]string) {
	configKeys := make([][]string, 0, len(primaryKeys))
	envKeys := make([][]string, 0, len(primaryKeys))
	for _, primary := range primaryKeys {
		var cfg []string
		var env []string
		for part := range strings.SplitSeq(primary, ",") {
			key := strings.TrimSpace(part)
			if key == "" {
				continue
			}
			if scanner.IsEnvVarStyle(key) {
				env = append(env, key)
			} else {
				cfg = append(cfg, key)
			}
		}
		configKeys = append(configKeys, cfg)
		envKeys = append(envKeys, env)
	}
	return configKeys, envKeys
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
			require.Equal(t, tt.wantEnv, scanner.IsEnvVarStyle(tt.key))
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
	requireWarningContains(t, warnings, `conflicting defaults for "key"`)
	requireWarningContains(t, warnings, `conflicting groups for "key"`)
}

func TestGroupOrder(t *testing.T) {
	entries := []configEntry{
		{PrimaryKey: "z.key", ConfigKeys: []string{"z.key"}, Default: "z", Group: "Zebra", GroupOrder: 3},
		{PrimaryKey: "a.key", ConfigKeys: []string{"a.key"}, Default: "a", Group: "Alpha", GroupOrder: 1},
		{PrimaryKey: "m.key", ConfigKeys: []string{"m.key"}, Default: "m", Group: "Middle", GroupOrder: 2},
		{PrimaryKey: "u.key", ConfigKeys: []string{"u.key"}, Default: "u", Group: "Unordered"},
		{PrimaryKey: "v.key", ConfigKeys: []string{"v.key"}, Default: "v", Group: "Another Unordered"},
	}

	md := render.FormatMarkdown(entries, "PREFIX")

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

func TestParseProjectGroupOrderDeclaredWithoutGetters(t *testing.T) {
	rootDir := t.TempDir()

	configSrc := `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:group General
	//cdoc:desc Name
	conf.GetStringVar("app", "app.name")
	//cdoc:group HTTP
	//cdoc:desc Port
	conf.GetIntVar(8080, 1, "http.port")
}`
	groupOrderSrc := `package test
//cdoc:group 1 HTTP
//cdoc:group 2 General
`

	err := os.WriteFile(filepath.Join(rootDir, "config.go"), []byte(configSrc), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(rootDir, "group_order.go"), []byte(groupOrderSrc), 0o644)
	require.NoError(t, err)

	entries, warnings, err := parseProject(rootDir)
	require.NoError(t, err)
	require.Empty(t, warnings)
	require.Len(t, entries, 2)

	ordersByKey := make(map[string]int, len(entries))
	for _, e := range entries {
		ordersByKey[e.PrimaryKey] = e.GroupOrder
	}
	require.Equal(t, 1, ordersByKey["http.port"])
	require.Equal(t, 2, ordersByKey["app.name"])

	md := render.FormatMarkdown(entries, "PREFIX")
	httpIdx := strings.Index(md, "## HTTP")
	generalIdx := strings.Index(md, "## General")
	require.GreaterOrEqual(t, httpIdx, 0)
	require.GreaterOrEqual(t, generalIdx, 0)
	require.Less(t, httpIdx, generalIdx)
}

func TestFormatMarkdown(t *testing.T) {
	entries := []configEntry{
		{PrimaryKey: "http.port", ConfigKeys: []string{"http.port"}, Default: "8080", Description: "HTTP server port", Group: "HTTP server"},
		{PrimaryKey: "etcd.hosts", ConfigKeys: []string{"etcd.hosts"}, EnvKeys: []string{"ETCD_HOSTS"}, Default: "localhost:2379", Description: "Etcd endpoints", Group: "General"},
		{PrimaryKey: "http.maxConns", ConfigKeys: []string{"http.maxConns"}, Default: "100", Description: "Max connections", Reloadable: true, Group: "HTTP server"},
	}

	md := render.FormatMarkdown(entries, "PREFIX")

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

	// Verify that -extrawarn would produce missing-description warnings.
	missingWarnings := generateWarnings(entries)
	requireWarningContains(t, missingWarnings, `"missingDescription" has no //cdoc:desc`)
	requireWarningContains(t, missingWarnings, `"deploymentName,RELEASE_NAME" has no //cdoc:group`)

	md := render.FormatMarkdown(entries, "PREFIX")

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

func TestParseProjectParseWarning(t *testing.T) {
	rootDir := t.TempDir()
	badFile := filepath.Join(rootDir, "bad.go")
	err := os.WriteFile(badFile, []byte("package bad\nfunc broken("), 0o644)
	require.NoError(t, err)

	entries, warnings, err := parseProject(rootDir)
	require.NoError(t, err)
	require.Empty(t, entries)
	requireWarningContains(t, warnings, "failed to parse")
	requireWarningContains(t, warnings, "bad.go")
}

func TestRun(t *testing.T) {
	output := filepath.Join(t.TempDir(), "output.md")
	err := run("testdata", output, "PREFIX", true, false)
	require.NoError(t, err)

	got, err := os.ReadFile(output)
	require.NoError(t, err)

	expected, err := os.ReadFile("testdata/expected_output.md")
	require.NoError(t, err, "reading golden file (run with -update to generate)")
	require.Equal(t, string(expected), string(got), "run output does not match golden file")
}

func TestRunFailOnWarning(t *testing.T) {
	output := filepath.Join(t.TempDir(), "output.md")
	err := run("testdata", output, "PREFIX", true, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "found")
	require.Contains(t, err.Error(), "warning")

	got, readErr := os.ReadFile(output)
	require.NoError(t, readErr)
	expected, readExpectedErr := os.ReadFile("testdata/expected_output.md")
	require.NoError(t, readExpectedErr)
	require.Equal(t, string(expected), string(got), "run output should still be generated")
}

func TestRunFailOnWarningNoWarnings(t *testing.T) {
	rootDir := t.TempDir()
	src := `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:group General
	//cdoc:desc d
	conf.GetStringVar("value", "some.key")
}`
	err := os.WriteFile(filepath.Join(rootDir, "example.go"), []byte(src), 0o644)
	require.NoError(t, err)

	output := filepath.Join(t.TempDir(), "output.md")
	err = run(rootDir, output, "PREFIX", true, true)
	require.NoError(t, err)

	got, readErr := os.ReadFile(output)
	require.NoError(t, readErr)
	require.Contains(t, string(got), "`some.key`")
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

type configEntry = model.Entry

func parseProject(rootDir string) ([]configEntry, []string, error) {
	entries, warnings, err := cdoc.ParseProject(rootDir)
	return entries, warningsToStrings(warnings), err
}

func extractFromFile(fset *token.FileSet, file *ast.File, filePath string) ([]configEntry, []string) {
	entries, warnings := scanner.ExtractFromFile(fset, file, filePath)
	return entries, warningsToStrings(warnings)
}

func deduplicateEntries(entries []configEntry) ([]configEntry, []string) {
	result, warnings := model.DeduplicateEntries(entries)
	return result, warningsToStrings(warnings)
}

func generateWarnings(entries []configEntry) []string {
	return warningsToStrings(cdoc.GenerateWarnings(entries))
}

func warningsToStrings(warnings []model.Warning) []string {
	return lo.Map(warnings, func(w model.Warning, _ int) string { return w.String() })
}
