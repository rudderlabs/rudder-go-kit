package engine

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/model"
	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/render"
	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/scanner"
)

func TestScanSingleFileWithMode_ExtractionMatrix(t *testing.T) {
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
			name: "GetStringVar/non-literal single-arg call default",
			src: `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func helper(v string) string { return v }
func f(conf *config.Config) {
	//cdoc:desc d
	conf.GetStringVar(helper("default"), "key")
}`,
			wantKeys: []string{"key"},
			wantDefs: []string{"${helper(\"default\")}"},
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

	for _, mode := range []ParseMode{ParseModeAST, ParseModeTypes} {
		t.Run(string(mode), func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					fset := token.NewFileSet()
					file, err := parser.ParseFile(fset, "test.go", tt.src, parser.ParseComments)
					require.NoError(t, err, "parse error")

					result, err := scanSingleFileWithMode(fset, file, "test.go", mode)
					require.NoError(t, err)
					entries := result.Entries
					warnings := result.Warnings
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
		})
	}
}

func TestParseProject_ParseWarningCode(t *testing.T) {
	for _, mode := range []ParseMode{ParseModeAST, ParseModeTypes} {
		t.Run(string(mode), func(t *testing.T) {
			root := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(root, "bad.go"), []byte("package bad\nfunc broken("), 0o644))

			entries, warnings, err := parseProjectWithMode(root, mode)
			require.NoError(t, err)
			require.Empty(t, entries)
			require.Len(t, warnings, 1)
			require.Equal(t, model.WarningCodeParseFailed, warnings[0].Code)
			require.Contains(t, warnings[0].Message, "failed to parse")
		})
	}
}

func TestParseProjectWithMode_TypesFiltersByReceiverType(t *testing.T) {
	root := t.TempDir()
	nonConfigSrc := `package test
type fake struct{}
func (fake) GetStringVar(value string, key string) string { return value }
func f() {
	//cdoc:desc d
	var cfg fake
	cfg.GetStringVar("value", "fake.key")
}`
	configSrc := `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func g(conf *config.Config) {
	//cdoc:desc method
	conf.GetStringVar("value", "real.key")
}
func h() {
	//cdoc:desc function
	config.GetStringVar("value", "func.key")
}`
	require.NoError(t, os.WriteFile(filepath.Join(root, "non_config.go"), []byte(nonConfigSrc), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "config.go"), []byte(configSrc), 0o644))

	astEntries, _, err := parseProjectWithMode(root, ParseModeAST)
	require.NoError(t, err)
	require.Len(t, astEntries, 3)

	typedEntries, _, err := parseProjectWithMode(root, ParseModeTypes)
	require.NoError(t, err)
	require.Len(t, typedEntries, 2)
	keys := []string{typedEntries[0].PrimaryKey, typedEntries[1].PrimaryKey}
	require.ElementsMatch(t, []string{"real.key", "func.key"}, keys)
}

func TestParseProjectWithMode_TypesDetectsConfigNewReceiver(t *testing.T) {
	root := t.TempDir()
	src := `package test
import "github.com/rudderlabs/rudder-go-kit/config"

func f() {
	conf := config.New(config.WithEnvPrefix("TEST"))
	//cdoc:desc json library
	_ = use(conf.GetStringVar("jsoniter", "jsonLib"))
}

func use(value string) string { return value }
`
	require.NoError(t, os.WriteFile(filepath.Join(root, "example.go"), []byte(src), 0o644))

	typedEntries, _, err := parseProjectWithMode(root, ParseModeTypes)
	require.NoError(t, err)
	require.Len(t, typedEntries, 1)
	require.Equal(t, "jsonLib", typedEntries[0].PrimaryKey)
	require.Equal(t, "json library", typedEntries[0].Description)
}

func TestGenerateWarnings_Codes(t *testing.T) {
	entries := []model.Entry{{PrimaryKey: "missing.all", File: "x.go", Line: 9}}

	warnings := generateExtraWarnings(entries)
	require.Len(t, warnings, 2)
	require.Equal(t, model.WarningCodeMissingDescription, warnings[0].Code)
	require.Equal(t, model.WarningCodeMissingGroup, warnings[1].Code)
}

func TestRun_WarningPolicyControlsBehavior(t *testing.T) {
	root := t.TempDir()
	src := `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	conf.GetStringVar("value", "some.key")
}`
	require.NoError(t, os.WriteFile(filepath.Join(root, "example.go"), []byte(src), 0o644))

	for _, mode := range []ParseMode{ParseModeAST, ParseModeTypes} {
		t.Run(string(mode), func(t *testing.T) {
			t.Run("warn policy does not fail", func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				err := Run(RunOptions{
					RootDir:       root,
					EnvPrefix:     "PREFIX",
					ExtraWarnings: true,
					ParseMode:     mode,
					Policy:        new(model.DefaultWarningPolicy()),
					Stdout:        &stdout,
					Stderr:        &stderr,
				})
				require.NoError(t, err)
				require.Contains(t, stdout.String(), "`some.key`")
				require.Contains(t, stderr.String(), "has no //cdoc:desc")
				require.Contains(t, stderr.String(), "has no //cdoc:group")
			})

			t.Run("strict policy fails", func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				err := Run(RunOptions{
					RootDir:       root,
					EnvPrefix:     "PREFIX",
					ExtraWarnings: true,
					ParseMode:     mode,
					Policy:        new(model.StrictWarningPolicy()),
					Stdout:        &stdout,
					Stderr:        &stderr,
				})
				require.Error(t, err)
				require.Contains(t, err.Error(), "found 2 warning(s)")
				require.Contains(t, stdout.String(), "`some.key`")
				require.NotEmpty(t, stderr.String())
			})

			t.Run("ignore policy with explicit overrides map suppresses warnings", func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				err := Run(RunOptions{
					RootDir:       root,
					EnvPrefix:     "PREFIX",
					ExtraWarnings: true,
					ParseMode:     mode,
					Policy: new(model.WarningPolicy{
						DefaultSeverity: model.SeverityIgnore,
						Overrides:       map[model.WarningCode]model.WarningSeverity{},
					}),
					Stdout: &stdout,
					Stderr: &stderr,
				})
				require.NoError(t, err)
				require.Contains(t, stdout.String(), "`some.key`")
				require.Empty(t, stderr.String())
			})

			t.Run("nil policy falls back to default warn behavior", func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				err := Run(RunOptions{
					RootDir:       root,
					EnvPrefix:     "PREFIX",
					ExtraWarnings: true,
					ParseMode:     mode,
					Stdout:        &stdout,
					Stderr:        &stderr,
				})
				require.NoError(t, err)
				require.Contains(t, stdout.String(), "`some.key`")
				require.Contains(t, stderr.String(), "has no //cdoc:desc")
			})

			t.Run("zero-value policy can explicitly ignore warnings", func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				err := Run(RunOptions{
					RootDir:       root,
					EnvPrefix:     "PREFIX",
					ExtraWarnings: true,
					ParseMode:     mode,
					Policy:        new(model.WarningPolicy{}),
					Stdout:        &stdout,
					Stderr:        &stderr,
				})
				require.NoError(t, err)
				require.Contains(t, stdout.String(), "`some.key`")
				require.Empty(t, stderr.String())
			})
		})
	}
}

func TestParseProjectWithMode_GroupOrderDeclaredWithoutGetters(t *testing.T) {
	for _, mode := range []ParseMode{ParseModeAST, ParseModeTypes} {
		t.Run(string(mode), func(t *testing.T) {
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

			entries, warnings, err := parseProjectWithMode(rootDir, mode)
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
		})
	}
}

func TestParseProjectWithMode_GoldenOutput(t *testing.T) {
	for _, mode := range []ParseMode{ParseModeAST, ParseModeTypes} {
		t.Run(string(mode), func(t *testing.T) {
			rootDir := filepath.Join("..", "..", "testdata")
			entries, warnings, err := parseProjectWithMode(rootDir, mode)
			require.NoError(t, err)

			requireWarningContains(t, warnings, "non-literal config key argument without //cdoc:key directive")

			missingWarnings := generateExtraWarnings(entries)
			requireWarningContains(t, missingWarnings, `"missingDescription" has no //cdoc:desc`)
			requireWarningContains(t, missingWarnings, `"deploymentName,RELEASE_NAME" has no //cdoc:group`)

			md := render.FormatMarkdown(entries, "PREFIX")

			expected, err := os.ReadFile(filepath.Join(rootDir, "expected_output.md"))
			require.NoError(t, err, "reading golden file")
			require.Equal(t, string(expected), md, "output does not match golden file")
		})
	}
}

func TestScanSingleFileWithMode_GroupOrderExtraction(t *testing.T) {
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
	for _, mode := range []ParseMode{ParseModeAST, ParseModeTypes} {
		t.Run(string(mode), func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
			require.NoError(t, err, "parse error")

			result, err := scanSingleFileWithMode(fset, file, "test.go", mode)
			require.NoError(t, err)
			require.Len(t, result.Entries, 2)
			require.Equal(t, 2, result.Entries[0].GroupOrder)
			require.Equal(t, 1, result.Entries[1].GroupOrder)
		})
	}
}

func requireWarningContains(t *testing.T, warnings []model.Warning, substr string) {
	t.Helper()
	for _, w := range warnings {
		if strings.Contains(w.String(), substr) {
			return
		}
	}
	t.Errorf("expected warning containing %q, got: %v", substr, warnings)
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
