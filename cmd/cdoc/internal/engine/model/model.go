package model

import (
	"fmt"
	"slices"
)

// Entry represents a single extracted configuration option.
type Entry struct {
	PrimaryKey string   // all key arguments joined with ",", used for dedup/sorting
	ConfigKeys []string // dotted/camelCase config-file keys
	EnvKeys    []string // UPPERCASE_STYLE env var keys (explicit)

	Default     string // default value as a string
	Description string // from //cdoc:desc
	Reloadable  bool   // true if from a GetReloadableX call, false if from GetX
	Group       string // from //cdoc:group, empty if ungrouped
	GroupOrder  int    // sort order for the group (0 = unordered, placed last)

	File string // for warnings
	Line int    // for warnings
}

type GroupOrderDeclaration struct {
	Group string
	Order int
	File  string
	Line  int
}

// DeduplicateEntries merges entries with the same primary key.
func DeduplicateEntries(entries []Entry) ([]Entry, []Warning) {
	var warnings []Warning
	seen := make(map[string]int) // primaryKey → index in result
	var result []Entry

	for _, e := range entries {
		if e.PrimaryKey == "" {
			continue
		}
		if idx, ok := seen[e.PrimaryKey]; ok {
			existing := &result[idx]
			// Do not backfill group metadata from a later duplicate in the same file.
			// This keeps group directives forward-only within a file.
			sameFileLater := existing.File != "" && e.File != "" && existing.File == e.File && e.Line > existing.Line
			skipGroupBackfill := sameFileLater && existing.Group == "" && e.Group != ""
			// Merge description.
			if existing.Description == "" && e.Description != "" {
				existing.Description = e.Description
			}
			// Merge group.
			if existing.Group == "" && e.Group != "" && !skipGroupBackfill {
				existing.Group = e.Group
			} else if existing.Group != "" && e.Group != "" && existing.Group != e.Group {
				warnings = append(warnings, Warning{
					Code: WarningCodeConflictingGroup,
					File: e.File,
					Line: e.Line,
					Message: fmt.Sprintf(
						"conflicting groups for %q: %q (%s) vs %q (%s)",
						e.PrimaryKey,
						existing.Group,
						entryLocation(*existing),
						e.Group,
						entryLocation(e),
					),
				})
			}
			// Warn on default mismatch.
			if existing.Default != e.Default {
				warnings = append(warnings, Warning{
					Code: WarningCodeConflictingDefault,
					File: e.File,
					Line: e.Line,
					Message: fmt.Sprintf(
						"conflicting defaults for %q: %q (%s) vs %q (%s)",
						e.PrimaryKey,
						existing.Default,
						entryLocation(*existing),
						e.Default,
						entryLocation(e),
					),
				})
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
			if existing.GroupOrder == 0 && e.GroupOrder != 0 && !skipGroupBackfill {
				existing.GroupOrder = e.GroupOrder
			}
		} else {
			seen[e.PrimaryKey] = len(result)
			result = append(result, e)
		}
	}
	return result, warnings
}

// entryLocation formats an entry source location for conflict warnings.
func entryLocation(entry Entry) string {
	if entry.File == "" && entry.Line <= 0 {
		return "unknown"
	}
	if entry.Line > 0 {
		return fmt.Sprintf("%s:%d", entry.File, entry.Line)
	}
	return entry.File
}

// ApplyProjectGroupOrders applies globally declared group orders to all entries.
// First declaration wins for a given group; conflicting declarations emit warnings.
func ApplyProjectGroupOrders(entries []Entry, declarations []GroupOrderDeclaration) []Warning {
	var warnings []Warning

	groupOrder := make(map[string]int)
	groupOrderOrigin := make(map[string]string)
	for _, declaration := range declarations {
		if existing, ok := groupOrder[declaration.Group]; ok {
			if existing != declaration.Order {
				warnings = append(warnings, Warning{
					Code: WarningCodeConflictingGroupOrder,
					Message: fmt.Sprintf(
						"conflicting group orders for %q: %d (%s) vs %d (%s:%d)",
						declaration.Group,
						existing,
						groupOrderOrigin[declaration.Group],
						declaration.Order,
						declaration.File,
						declaration.Line,
					),
				})
			}
			continue
		}
		groupOrder[declaration.Group] = declaration.Order
		groupOrderOrigin[declaration.Group] = fmt.Sprintf("%s:%d", declaration.File, declaration.Line)
	}

	for i := range entries {
		if entries[i].Group == "" {
			continue
		}
		if order, ok := groupOrder[entries[i].Group]; ok {
			entries[i].GroupOrder = order
		}
	}
	return warnings
}
