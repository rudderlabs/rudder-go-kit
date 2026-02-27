package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWarningStringFormatting(t *testing.T) {
	require.Equal(t, "warning: a.go:12: problem", Warning{File: "a.go", Line: 12, Message: "problem"}.String())
	require.Equal(t, "warning: a.go: problem", Warning{File: "a.go", Message: "problem"}.String())
	require.Equal(t, "warning: problem", Warning{Message: "problem"}.String())
}

func TestClassifyWarnings_WithOverrides(t *testing.T) {
	warnings := []Warning{
		{Code: WarningCodeMissingDescription, Message: "m1"},
		{Code: WarningCodeUnusedKeyOverride, Message: "m2"},
		{Code: WarningCodeConflictingDefault, Message: "m3"},
	}
	policy := WarningPolicy{
		DefaultSeverity: SeverityWarn,
		Overrides: map[WarningCode]WarningSeverity{
			WarningCodeUnusedKeyOverride:  SeverityIgnore,
			WarningCodeConflictingDefault: SeverityError,
		},
	}

	classified := ClassifyWarnings(warnings, policy)
	require.Len(t, classified, 3)
	require.Equal(t, SeverityWarn, classified[0].Severity)
	require.Equal(t, SeverityIgnore, classified[1].Severity)
	require.Equal(t, SeverityError, classified[2].Severity)
}

func TestDefaultAndStrictPolicies(t *testing.T) {
	require.Equal(t, SeverityWarn, DefaultWarningPolicy().SeverityFor(WarningCodeMissingGroup))
	require.Equal(t, SeverityError, StrictWarningPolicy().SeverityFor(WarningCodeMissingGroup))
}
