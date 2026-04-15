package iabparser

import "time"

// BlacklistEntry represents a single parsed and validated blacklist entry.
type BlacklistEntry struct {
	Pattern       string    // case-insensitive string to match against the user agent
	Active        bool      // whether this entry is currently active
	Exceptions    []string  // comma-and-space separated list of exception patterns
	DualPassFlag  int       // 1 = redundant if using two-pass detection, 0 = always needed
	ImpactType    int       // 0 = page impressions, 1 = ad impressions, 2 = both
	StartOfString bool      // true = must match from start of UA, false = match anywhere
	InactiveDate  time.Time // zero value for active entries, set for inactive
}

// WhitelistEntry represents a single parsed and validated whitelist entry.
type WhitelistEntry struct {
	Pattern       string    // case-insensitive string to match against the user agent
	Active        bool      // whether this entry is currently active
	StartOfString bool      // true = must match from start of UA, false = match anywhere
	InactiveDate  time.Time // zero value for active entries, set for inactive
}
