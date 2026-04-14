package iabparser

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestValidateBlacklist(t *testing.T) {
	t.Run("should accept valid file", func(t *testing.T) {
		input := "Googlebot/|1||0|1|0\nOldBot|0||1|2|0|03/15/2023\n"
		require.NoError(t, ValidateBlacklist(strings.NewReader(input)))
	})

	t.Run("should accept file with no active entries", func(t *testing.T) {
		input := "OldBot|0||0|0|0|01/01/2020\n"
		require.NoError(t, ValidateBlacklist(strings.NewReader(input)))
	})

	t.Run("should reject invalid line", func(t *testing.T) {
		input := "Bot|1||0|BAD|0\n"
		require.ErrorContains(t, ValidateBlacklist(strings.NewReader(input)), "line 1")
	})
}

func TestValidateWhitelist(t *testing.T) {
	t.Run("should accept valid file", func(t *testing.T) {
		input := "Mozilla/|1|1\nOldBrowser|0|0|06/01/2022\n"
		require.NoError(t, ValidateWhitelist(strings.NewReader(input)))
	})

	t.Run("should accept file with no active entries", func(t *testing.T) {
		input := "OldBrowser|0|0|01/01/2020\n"
		require.NoError(t, ValidateWhitelist(strings.NewReader(input)))
	})

	t.Run("should reject invalid line", func(t *testing.T) {
		input := "Mozilla/|bad|1\n"
		require.ErrorContains(t, ValidateWhitelist(strings.NewReader(input)), "line 1")
	})
}

func TestParseBlacklist(t *testing.T) {
	t.Run("should parse active entry with all valid fields", func(t *testing.T) {
		input := "Googlebot/|1||0|2|0\n"
		entries, err := ParseBlacklist(strings.NewReader(input))
		require.NoError(t, err)
		require.Len(t, entries, 1)

		e := entries[0]
		require.Equal(t, "Googlebot/", e.Pattern)
		require.True(t, e.Active)
		require.Empty(t, e.Exceptions)
		require.Equal(t, 0, e.DualPassFlag)
		require.Equal(t, 2, e.ImpactType)
		require.False(t, e.StartOfString)
		require.True(t, e.InactiveDate.IsZero())
	})

	t.Run("should parse multiple valid active entries", func(t *testing.T) {
		input := "Googlebot/|1||0|1|0\nBingbot|1|exception1, exception2|0|0|1\n"
		entries, err := ParseBlacklist(strings.NewReader(input))
		require.NoError(t, err)
		require.Len(t, entries, 2)
	})

	t.Run("should parse entry with exceptions", func(t *testing.T) {
		input := "Bot|1|exc1, exc2, exc3|1|1|1\n"
		entries, err := ParseBlacklist(strings.NewReader(input))
		require.NoError(t, err)
		require.Len(t, entries, 1)
		require.Equal(t, []string{"exc1", "exc2", "exc3"}, entries[0].Exceptions)
	})

	t.Run("should parse inactive entry with date", func(t *testing.T) {
		input := "ActiveBot|1||0|0|0\nOldBot|0||1|2|0|03/15/2023\n"
		entries, err := ParseBlacklist(strings.NewReader(input))
		require.NoError(t, err)
		require.Len(t, entries, 2)

		require.True(t, entries[0].Active)
		require.True(t, entries[0].InactiveDate.IsZero())

		require.False(t, entries[1].Active)
		require.Equal(t, time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC), entries[1].InactiveDate)
	})

	t.Run("should skip comments and blank lines", func(t *testing.T) {
		input := "# This is a comment\n\n  # Indented comment\n  \nGooglebot/|1||0|1|0\n"
		entries, err := ParseBlacklist(strings.NewReader(input))
		require.NoError(t, err)
		require.Len(t, entries, 1)
		e := entries[0]
		require.Equal(t, e.Pattern, "Googlebot/")
		require.True(t, e.Active)
		require.Empty(t, e.Exceptions)
		require.Equal(t, 0, e.DualPassFlag)
		require.Equal(t, 1, e.ImpactType)
		require.False(t, e.StartOfString)
		require.True(t, e.InactiveDate.IsZero())
	})

	t.Run("should return empty for empty file", func(t *testing.T) {
		entries, err := ParseBlacklist(strings.NewReader(""))
		require.NoError(t, err)
		require.Empty(t, entries)
	})

	t.Run("should return empty for file with only comments and blank lines", func(t *testing.T) {
		input := "# comment\n\n# another comment\n"
		entries, err := ParseBlacklist(strings.NewReader(input))
		require.NoError(t, err)
		require.Empty(t, entries)
	})

	t.Run("should reject wrong field count", func(t *testing.T) {
		input := "Googlebot/|1|0\n"
		_, err := ParseBlacklist(strings.NewReader(input))
		require.ErrorContains(t, err, "line 1")
		require.ErrorContains(t, err, "expected 6 or 7")
	})

	t.Run("should reject too many fields", func(t *testing.T) {
		input := "Bot|1||0|0|0|01/01/2020|extra\n"
		_, err := ParseBlacklist(strings.NewReader(input))
		require.ErrorContains(t, err, "expected 6 or 7")
	})

	t.Run("should reject empty pattern", func(t *testing.T) {
		input := "|1||0|0|0\n"
		_, err := ParseBlacklist(strings.NewReader(input))
		require.ErrorContains(t, err, "pattern (field 1) must be non-empty")
	})

	t.Run("should reject invalid active flag", func(t *testing.T) {
		input := "Bot|2||0|0|0\n"
		_, err := ParseBlacklist(strings.NewReader(input))
		require.ErrorContains(t, err, "active flag (field 2)")
	})

	t.Run("should reject invalid dual-pass flag", func(t *testing.T) {
		input := "Bot|1||3|0|0\n"
		_, err := ParseBlacklist(strings.NewReader(input))
		require.ErrorContains(t, err, "dual-pass flag (field 4)")
	})

	t.Run("should reject invalid impact type", func(t *testing.T) {
		input := "Bot|1||0|5|0\n"
		_, err := ParseBlacklist(strings.NewReader(input))
		require.ErrorContains(t, err, "impact type (field 5)")
	})

	t.Run("should reject invalid start-of-string flag", func(t *testing.T) {
		input := "Bot|1||0|0|9\n"
		_, err := ParseBlacklist(strings.NewReader(input))
		require.ErrorContains(t, err, "start-of-string flag (field 6)")
	})

	t.Run("should reject invalid inactive date format", func(t *testing.T) {
		input := "Bot|0||0|0|0|2023-03-15\n"
		_, err := ParseBlacklist(strings.NewReader(input))
		require.ErrorContains(t, err, "inactive date must be mm/dd/yyyy")
	})

	t.Run("should stop at first invalid line in mixed input", func(t *testing.T) {
		input := "GoodBot|1||0|0|0\nBadBot|1||0|BAD|0\nAnotherGood|1||0|0|0\n"
		_, err := ParseBlacklist(strings.NewReader(input))
		require.ErrorContains(t, err, "line 2")
	})
}

func TestParseWhitelist(t *testing.T) {
	t.Run("should parse valid active entries", func(t *testing.T) {
		input := "Mozilla/|1|1\nChrome|1|0\n"
		entries, err := ParseWhitelist(strings.NewReader(input))
		require.NoError(t, err)
		require.Len(t, entries, 2)

		require.Equal(t, "Mozilla/", entries[0].Pattern)
		require.True(t, entries[0].Active)
		require.True(t, entries[0].StartOfString)
		require.True(t, entries[0].InactiveDate.IsZero())

		require.Equal(t, "Chrome", entries[1].Pattern)
		require.True(t, entries[1].Active)
		require.False(t, entries[1].StartOfString)
		require.True(t, entries[1].InactiveDate.IsZero())
	})

	t.Run("should parse inactive entry with date", func(t *testing.T) {
		input := "Mozilla/|1|1\nOldBrowser|0|0|06/01/2022\n"
		entries, err := ParseWhitelist(strings.NewReader(input))
		require.NoError(t, err)
		require.Len(t, entries, 2)

		require.False(t, entries[1].Active)
		require.Equal(t, time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC), entries[1].InactiveDate)
	})

	t.Run("should skip comments and blank lines", func(t *testing.T) {
		input := "# comment\n\n  \nMozilla/|1|1\n"
		entries, err := ParseWhitelist(strings.NewReader(input))
		require.NoError(t, err)
		require.Len(t, entries, 1)
	})

	t.Run("should return empty for empty file", func(t *testing.T) {
		entries, err := ParseWhitelist(strings.NewReader(""))
		require.NoError(t, err)
		require.Empty(t, entries)
	})

	t.Run("should return empty for file with only comments and blank lines", func(t *testing.T) {
		input := "# comment\n\n"
		entries, err := ParseWhitelist(strings.NewReader(input))
		require.NoError(t, err)
		require.Empty(t, entries)
	})

	t.Run("should reject wrong field count", func(t *testing.T) {
		input := "Mozilla/|1\n"
		_, err := ParseWhitelist(strings.NewReader(input))
		require.ErrorContains(t, err, "expected 3 or 4")
	})

	t.Run("should reject too many fields", func(t *testing.T) {
		input := "Mozilla/|1|1|01/01/2020|extra\n"
		_, err := ParseWhitelist(strings.NewReader(input))
		require.ErrorContains(t, err, "expected 3 or 4")
	})

	t.Run("should reject empty pattern", func(t *testing.T) {
		input := "|1|1\n"
		_, err := ParseWhitelist(strings.NewReader(input))
		require.ErrorContains(t, err, "pattern (field 1) must be non-empty")
	})

	t.Run("should reject invalid active flag", func(t *testing.T) {
		input := "Mozilla/|yes|1\n"
		_, err := ParseWhitelist(strings.NewReader(input))
		require.ErrorContains(t, err, "active flag (field 2)")
	})

	t.Run("should reject invalid start-of-string flag", func(t *testing.T) {
		input := "Mozilla/|1|X\n"
		_, err := ParseWhitelist(strings.NewReader(input))
		require.ErrorContains(t, err, "start-of-string flag (field 3)")
	})

	t.Run("should reject invalid inactive date format", func(t *testing.T) {
		input := "Mozilla/|0|0|15-03-2023\n"
		_, err := ParseWhitelist(strings.NewReader(input))
		require.ErrorContains(t, err, "inactive date must be mm/dd/yyyy")
	})

	t.Run("should stop at first invalid line in mixed input", func(t *testing.T) {
		input := "Good|1|0\n|1|0\nAlsoGood|1|1\n"
		_, err := ParseWhitelist(strings.NewReader(input))
		require.ErrorContains(t, err, "line 2")
	})
}
