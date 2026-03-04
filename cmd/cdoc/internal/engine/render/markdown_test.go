package render

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/model"
)

func TestFormatMarkdown_GroupOrderingAndReloadable(t *testing.T) {
	entries := []model.Entry{
		{PrimaryKey: "u.key", ConfigKeys: []string{"u.key"}, Group: "UngroupedName", Default: "u", Description: "u"},
		{PrimaryKey: "h.key", ConfigKeys: []string{"h.key"}, Group: "HTTP", GroupOrder: 2, Default: "1", Description: "http"},
		{PrimaryKey: "g.key", ConfigKeys: []string{"g.key"}, Group: "General", GroupOrder: 1, Default: "2", Description: "general"},
		{PrimaryKey: "r.key", ConfigKeys: []string{"r.key"}, Group: "HTTP", GroupOrder: 2, Default: "3", Description: "reload", Reloadable: true},
	}

	md := FormatMarkdown(entries, "PREFIX")

	generalIdx := strings.Index(md, "## General")
	httpIdx := strings.Index(md, "## HTTP")
	unorderedIdx := strings.Index(md, "## UngroupedName")
	require.Greater(t, generalIdx, -1)
	require.Greater(t, httpIdx, -1)
	require.Greater(t, unorderedIdx, -1)
	require.Less(t, generalIdx, httpIdx)
	require.Less(t, httpIdx, unorderedIdx)
	require.Contains(t, md, "🔄 reload")
}

func TestFormatMarkdown_NoUngroupedHeaderWhenSingleGroup(t *testing.T) {
	entries := []model.Entry{
		{PrimaryKey: "only.key", ConfigKeys: []string{"only.key"}, Default: "value", Description: "desc"},
	}

	md := FormatMarkdown(entries, "PREFIX")
	require.NotContains(t, md, "## Ungrouped")
	require.Contains(t, md, "| Config variable | Env variable | Default | Description |")
}

func TestFormatMarkdown_EnvColumnDeduplicatesDerivedAndExplicit(t *testing.T) {
	entries := []model.Entry{
		{
			PrimaryKey:  "http.port",
			ConfigKeys:  []string{"http.port"},
			EnvKeys:     []string{"PREFIX_HTTP_PORT", "EXPLICIT_ONLY"},
			Default:     "8080",
			Description: "desc",
		},
	}

	md := FormatMarkdown(entries, "PREFIX")
	require.Equal(t, 1, strings.Count(md, "`PREFIX_HTTP_PORT`"))
	require.Equal(t, 1, strings.Count(md, "`EXPLICIT_ONLY`"))
}

func TestFormatMarkdown_EscapesTableCells(t *testing.T) {
	entries := []model.Entry{
		{
			PrimaryKey:  "app|name",
			ConfigKeys:  []string{"app|name"},
			Default:     "a|b\nc",
			Description: "line1|line2\nline3",
		},
	}

	md := FormatMarkdown(entries, "PREFIX")
	require.Contains(t, md, "`app|name`")
	require.Contains(t, md, "`a|b`<br>`c`")
	require.Contains(t, md, "line1\\|line2<br>line3")
}

func TestFormatMarkdown_BackticksInCodeCells(t *testing.T) {
	entries := []model.Entry{
		{
			PrimaryKey:  "tick`key",
			ConfigKeys:  []string{"tick`key"},
			Default:     "a``b",
			Description: "desc",
		},
	}

	md := FormatMarkdown(entries, "PREFIX")
	require.Contains(t, md, "``tick`key``")
	require.Contains(t, md, "```a``b```")
}

func TestFormatMarkdown_GroupOrderSortsOrderedThenAlphabeticalUnordered(t *testing.T) {
	entries := []model.Entry{
		{PrimaryKey: "z.key", ConfigKeys: []string{"z.key"}, Default: "z", Group: "Zebra", GroupOrder: 3},
		{PrimaryKey: "a.key", ConfigKeys: []string{"a.key"}, Default: "a", Group: "Alpha", GroupOrder: 1},
		{PrimaryKey: "m.key", ConfigKeys: []string{"m.key"}, Default: "m", Group: "Middle", GroupOrder: 2},
		{PrimaryKey: "u.key", ConfigKeys: []string{"u.key"}, Default: "u", Group: "Unordered"},
		{PrimaryKey: "v.key", ConfigKeys: []string{"v.key"}, Default: "v", Group: "Another Unordered"},
	}

	md := FormatMarkdown(entries, "PREFIX")

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
