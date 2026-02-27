package ulp

import (
	"strings"
	"unicode"
)

// specialCharsForEventID are characters removed from tokens when generating EventID.
const specialCharsForEventID = "!@#$%^&*()[]{};:,/<>?\\|`~-=+"

// generateEventID creates a group identifier for a preprocessed log message.
// Algorithm:
//  1. Count total words from the original preprocessed string
//  2. Remove special characters
//  3. Keep only purely alphabetic words
//  4. EventID = concatenation of filtered words + string(length)
func generateEventID(tokenString string) string {
	// Count words from ORIGINAL string (per paper Algorithm 1)
	length := len(strings.Fields(tokenString))

	// Remove special characters, then split into words
	cleaned := removeSpecialChars(tokenString)
	words := strings.Fields(cleaned)

	// Filter: keep only purely alphabetic words
	var filtered []string
	for _, w := range words {
		if isAlpha(w) {
			filtered = append(filtered, w)
		}
	}

	// Build EventID
	var b strings.Builder
	for _, w := range filtered {
		b.WriteString(w)
	}
	b.WriteString(intToStr(length))

	return b.String()
}

// removeSpecialChars strips specialCharsForEventID characters from s.
func removeSpecialChars(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if !strings.ContainsRune(specialCharsForEventID, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// isAlpha returns true if the string contains only ASCII letters.
func isAlpha(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// groupEvents groups events by their EventID.
func groupEvents(events []*LogEvent) map[string]*LogGroup {
	groups := make(map[string]*LogGroup)
	for _, ev := range events {
		g, ok := groups[ev.EventID]
		if !ok {
			g = &LogGroup{
				EventID: ev.EventID,
			}
			groups[ev.EventID] = g
		}
		g.Events = append(g.Events, ev)
	}
	return groups
}

// intToStr converts a small non-negative int to string without fmt.Sprintf.
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + intToStr(-n)
	}
	var digits []byte
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	// Reverse
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}
