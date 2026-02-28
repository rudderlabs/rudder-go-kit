package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeduplicateEntries_MergeAndConflictBehavior(t *testing.T) {
	entries := []Entry{
		{
			PrimaryKey: "same.key",
			ConfigKeys: []string{"same.key"},
			Default:    "a",
			File:       "a.go",
			Line:       10,
		},
		{
			PrimaryKey:  "same.key",
			ConfigKeys:  []string{"same.key"},
			Default:     "a",
			Description: "desc",
			Group:       "LocalGroup",
			GroupOrder:  5,
			Reloadable:  true,
			EnvKeys:     []string{"SAME_KEY"},
			File:        "a.go",
			Line:        20,
		},
		{
			PrimaryKey: "same.key",
			ConfigKeys: []string{"same.key"},
			Default:    "b",
			Group:      "OtherGroup",
			File:       "b.go",
			Line:       1,
		},
	}

	merged, warnings := DeduplicateEntries(entries)
	require.Len(t, merged, 1)

	entry := merged[0]
	require.Equal(t, "desc", entry.Description)
	require.Equal(t, "OtherGroup", entry.Group, "group should backfill from other files")
	require.Equal(t, 0, entry.GroupOrder, "group order from same-file-later duplicate should not backfill")
	require.True(t, entry.Reloadable)
	require.Equal(t, []string{"SAME_KEY"}, entry.EnvKeys)

	require.Len(t, warnings, 1)
	require.Equal(t, WarningCodeConflictingDefault, warnings[0].Code)
	require.Equal(t, "b.go", warnings[0].File)
	require.Equal(t, 1, warnings[0].Line)
	require.Contains(t, warnings[0].Message, "a.go:10")
	require.Contains(t, warnings[0].Message, "b.go:1")
}

func TestApplyProjectGroupOrders_FirstDeclarationWins(t *testing.T) {
	entries := []Entry{
		{PrimaryKey: "a", Group: "HTTP"},
		{PrimaryKey: "b", Group: "General"},
		{PrimaryKey: "c", Group: "NoOrder"},
	}
	declarations := []GroupOrderDeclaration{
		{Group: "HTTP", Order: 2, File: "one.go", Line: 1},
		{Group: "HTTP", Order: 3, File: "two.go", Line: 1},
		{Group: "General", Order: 1, File: "two.go", Line: 2},
	}

	warnings := ApplyProjectGroupOrders(entries, declarations)

	require.Equal(t, 2, entries[0].GroupOrder)
	require.Equal(t, 1, entries[1].GroupOrder)
	require.Equal(t, 0, entries[2].GroupOrder)

	require.Len(t, warnings, 1)
	require.Equal(t, WarningCodeConflictingGroupOrder, warnings[0].Code)
	require.Contains(t, warnings[0].Message, "conflicting group orders")
}

func TestDeduplicateEntries_BasicMerge(t *testing.T) {
	entries := []Entry{
		{PrimaryKey: "http.port", ConfigKeys: []string{"http.port"}, Default: "8080", Description: "HTTP port", Group: "HTTP"},
		{PrimaryKey: "http.port", ConfigKeys: []string{"http.port"}, Default: "8080", Description: "", Group: ""},
		{PrimaryKey: "other.key", ConfigKeys: []string{"other.key"}, Default: "val", Description: "Other", Group: "General"},
	}

	result, warnings := DeduplicateEntries(entries)
	require.Len(t, result, 2)
	require.Empty(t, warnings)
	require.Equal(t, "HTTP port", result[0].Description)
	require.Equal(t, "HTTP", result[0].Group)
}

func TestDeduplicateEntries_ConflictWarnings(t *testing.T) {
	entries := []Entry{
		{PrimaryKey: "key", ConfigKeys: []string{"key"}, Default: "a", Group: "G1"},
		{PrimaryKey: "key", ConfigKeys: []string{"key"}, Default: "b", Group: "G2"},
	}

	_, warnings := DeduplicateEntries(entries)
	require.Len(t, warnings, 2, "expected default + group conflict warnings")
	codes := []WarningCode{warnings[0].Code, warnings[1].Code}
	require.ElementsMatch(t, []WarningCode{WarningCodeConflictingDefault, WarningCodeConflictingGroup}, codes)
}
