// Package iabparser provides validation and parsing for IAB blacklist (robots/spiders)
// and whitelist (valid browsers) files from the IAB Transparency Center.
package iabparser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

// ValidateBlacklist validates all entries in a blacklist file line-by-line
// without allocating entry structs. Returns an error if any entry fails validation.
func ValidateBlacklist(r io.Reader) error {
	return validateLines(r, parseBlacklistLine)
}

// ValidateWhitelist validates all entries in a whitelist file line-by-line
// without allocating entry structs. Returns an error if any entry fails validation.
func ValidateWhitelist(r io.Reader) error {
	return validateLines(r, parseWhitelistLine)
}

// ParseBlacklist validates and parses all entries from a blacklist file into structured types.
// Returns an error if any entry fails validation.
func ParseBlacklist(r io.Reader) ([]BlacklistEntry, error) {
	return parseLines(r, parseBlacklistLine)
}

// ParseWhitelist validates and parses all entries from a whitelist file into structured types.
// Returns an error if any entry fails validation.
func ParseWhitelist(r io.Reader) ([]WhitelistEntry, error) {
	return parseLines(r, parseWhitelistLine)
}

func validateLines[T any](r io.Reader, parseLine func(string) (T, error)) error {
	_, err := scanLines(r, parseLine, false)
	return err
}

func parseLines[T any](r io.Reader, parseLine func(string) (T, error)) ([]T, error) {
	return scanLines(r, parseLine, true)
}

func scanLines[T any](r io.Reader, parseLine func(string) (T, error), collect bool) ([]T, error) {
	var (
		entries []T
		lineNum int
	)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if shouldSkipLine(line) {
			continue
		}
		entry, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
		if collect {
			entries = append(entries, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}
	return entries, nil
}

func shouldSkipLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "" || strings.HasPrefix(trimmed, "#")
}

func parseBlacklistLine(line string) (BlacklistEntry, error) {
	fields := strings.Split(line, "|")
	nf := len(fields)

	if nf != 6 && nf != 7 {
		return BlacklistEntry{}, fmt.Errorf("expected 6 or 7 pipe-delimited fields, got %d", nf)
	}

	if fields[0] == "" {
		return BlacklistEntry{}, fmt.Errorf("pattern (field 1) must be non-empty")
	}

	isActive, err := parseBinaryFlag(fields[1], "active flag (field 2)")
	if err != nil {
		return BlacklistEntry{}, err
	}

	var exceptions []string
	if fields[2] != "" {
		exceptions = strings.Split(fields[2], ", ")
	}

	dualPass, err := parseBinaryFlagInt(fields[3], "dual-pass flag (field 4)")
	if err != nil {
		return BlacklistEntry{}, err
	}

	impactType, err := parseImpactType(fields[4])
	if err != nil {
		return BlacklistEntry{}, err
	}

	startOfString, err := parseBinaryFlag(fields[5], "start-of-string flag (field 6)")
	if err != nil {
		return BlacklistEntry{}, err
	}

	var inactiveDate time.Time
	if nf == 7 {
		inactiveDate, err = parseInactiveDate(fields[6])
		if err != nil {
			return BlacklistEntry{}, err
		}
	}

	return BlacklistEntry{
		Pattern:       fields[0],
		Active:        isActive,
		Exceptions:    exceptions,
		DualPassFlag:  dualPass,
		ImpactType:    impactType,
		StartOfString: startOfString,
		InactiveDate:  inactiveDate,
	}, nil
}

func parseWhitelistLine(line string) (WhitelistEntry, error) {
	fields := strings.Split(line, "|")
	nf := len(fields)

	if nf != 3 && nf != 4 {
		return WhitelistEntry{}, fmt.Errorf("expected 3 or 4 pipe-delimited fields, got %d", nf)
	}

	if fields[0] == "" {
		return WhitelistEntry{}, fmt.Errorf("pattern (field 1) must be non-empty")
	}

	isActive, err := parseBinaryFlag(fields[1], "active flag (field 2)")
	if err != nil {
		return WhitelistEntry{}, err
	}

	startOfString, err := parseBinaryFlag(fields[2], "start-of-string flag (field 3)")
	if err != nil {
		return WhitelistEntry{}, err
	}

	var inactiveDate time.Time
	if nf == 4 {
		inactiveDate, err = parseInactiveDate(fields[3])
		if err != nil {
			return WhitelistEntry{}, err
		}
	}

	return WhitelistEntry{
		Pattern:       fields[0],
		Active:        isActive,
		StartOfString: startOfString,
		InactiveDate:  inactiveDate,
	}, nil
}

func parseBinaryFlag(s, name string) (bool, error) {
	switch s {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, fmt.Errorf("%s must be \"0\" or \"1\", got %q", name, s)
	}
}

func parseBinaryFlagInt(s, name string) (int, error) {
	switch s {
	case "0":
		return 0, nil
	case "1":
		return 1, nil
	default:
		return 0, fmt.Errorf("%s must be \"0\" or \"1\", got %q", name, s)
	}
}

func parseImpactType(s string) (int, error) {
	switch s {
	case "0":
		return 0, nil
	case "1":
		return 1, nil
	case "2":
		return 2, nil
	default:
		return 0, fmt.Errorf("impact type (field 5) must be \"0\", \"1\", or \"2\", got %q", s)
	}
}

func parseInactiveDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse("01/02/2006", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("inactive date must be mm/dd/yyyy format, got %q", s)
	}
	return t, nil
}
