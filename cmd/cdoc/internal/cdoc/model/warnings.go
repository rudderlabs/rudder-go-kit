package model

import "fmt"

type WarningCode string

const (
	WarningCodeParseFailed            WarningCode = "parse_failed"
	WarningCodeArgCount               WarningCode = "invalid_arg_count"
	WarningCodeDynamicKeyMissing      WarningCode = "dynamic_key_missing"
	WarningCodeUnusedKeyOverride      WarningCode = "unused_key_override"
	WarningCodeConflictingGroupOrder  WarningCode = "conflicting_group_order"
	WarningCodeConflictingGroup       WarningCode = "conflicting_group"
	WarningCodeConflictingDefault     WarningCode = "conflicting_default"
	WarningCodeUnusedDescDirective    WarningCode = "unused_desc_directive"
	WarningCodeUnusedDefaultDirective WarningCode = "unused_default_directive"
	WarningCodeMissingDescription     WarningCode = "missing_description"
	WarningCodeMissingGroup           WarningCode = "missing_group"
)

type Warning struct {
	Code    WarningCode
	File    string
	Line    int
	Message string
}

func (w Warning) String() string {
	switch {
	case w.File != "" && w.Line > 0:
		return fmt.Sprintf("warning: %s:%d: %s", w.File, w.Line, w.Message)
	case w.File != "":
		return fmt.Sprintf("warning: %s: %s", w.File, w.Message)
	default:
		return fmt.Sprintf("warning: %s", w.Message)
	}
}

type WarningSeverity int

const (
	SeverityIgnore WarningSeverity = iota
	SeverityWarn
	SeverityError
)

type WarningPolicy struct {
	DefaultSeverity WarningSeverity
	Overrides       map[WarningCode]WarningSeverity
}

func DefaultWarningPolicy() WarningPolicy {
	return WarningPolicy{DefaultSeverity: SeverityWarn}
}

func StrictWarningPolicy() WarningPolicy {
	return WarningPolicy{DefaultSeverity: SeverityError}
}

func (p WarningPolicy) SeverityFor(code WarningCode) WarningSeverity {
	if p.Overrides != nil {
		if severity, ok := p.Overrides[code]; ok {
			return severity
		}
	}
	return p.DefaultSeverity
}

type ClassifiedWarning struct {
	Warning  Warning
	Severity WarningSeverity
}

func ClassifyWarnings(warnings []Warning, policy WarningPolicy) []ClassifiedWarning {
	classified := make([]ClassifiedWarning, 0, len(warnings))
	for _, warning := range warnings {
		classified = append(classified, ClassifiedWarning{
			Warning:  warning,
			Severity: policy.SeverityFor(warning.Code),
		})
	}
	return classified
}
