package config

// TODO: everything in this file should be either removed or unexported
import (
	"os"
	"strings"
	"unicode"
)

func isUpperCaseConfigKey(s string) bool {
	for _, ch := range s {
		if !(ch == '_' || unicode.IsUpper(ch) || unicode.IsDigit(ch)) {
			return false
		}
	}
	return true
}

// ConfigKeyToEnv gets the env variable name from a given config key
func ConfigKeyToEnv(envPrefix, s string) string {
	// Check if the string is already in upper case format
	if isUpperCaseConfigKey(s) {
		return s
	}

	// convert camelCase to snake_case
	var builder strings.Builder

	// Add the prefix
	builder.WriteString(envPrefix)
	builder.WriteByte('_')

	// Transform the input string to the desired format
	for i, r := range s {
		if r >= 'A' && r <= 'Z' && i > 0 && (s[i-1] >= 'a' && s[i-1] <= 'z' || s[i-1] >= '0' && s[i-1] <= '9') {
			builder.WriteByte('_')
		} else if r == '.' {
			r = '_'
		}
		if unicode.IsLetter(r) {
			r = unicode.ToUpper(r)
		}
		builder.WriteRune(r)
	}

	return builder.String()
}

// getEnv returns the environment value stored in key variable
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
