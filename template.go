package ulp

import (
	"regexp"
	"strings"
)

// initWildcardRe initializes the regex for collapsing consecutive wildcards.
// Must be called after the parser's dynamicWildcard is set.
func initWildcardRe(wildcard string) *regexp.Regexp {
	escaped := regexp.QuoteMeta(wildcard)
	return regexp.MustCompile(escaped + `(\s+` + escaped + `)+`)
}

// generateTemplate creates a template from a group of events.
// Algorithm:
//  1. Take first event's token string as the initial template
//  2. Build vocabulary with per-event deduplication
//  3. Replace tokens that don't appear in all events (dynamic) with wildcard
//  4. Optionally replace standalone numbers with wildcard
//  5. Collapse consecutive wildcards
func generateTemplate(group *LogGroup, wildcard string, sampleSize int, replaceNumbers bool) string {
	if len(group.Events) == 0 {
		return ""
	}

	// Single event — use it as template
	if len(group.Events) == 1 {
		return cleanupTemplate(group.Events[0].TokenString, wildcard, replaceNumbers)
	}

	// Sample events for frequency analysis
	sampled := sampleEvents(group.Events, sampleSize)

	// Build vocabulary
	vocab := getVocabulary(sampled)
	groupSize := len(sampled)

	// Find dynamic tokens
	dynamic := findDynamicTokens(vocab, groupSize)

	// Use first event as template base
	templateTokens := strings.Fields(group.Events[0].TokenString)
	var b strings.Builder
	for i, tok := range templateTokens {
		if i > 0 {
			b.WriteByte(' ')
		}
		if dynamic[tok] {
			b.WriteString(wildcard)
		} else {
			b.WriteString(tok)
		}
	}

	return cleanupTemplate(b.String(), wildcard, replaceNumbers)
}

// cleanupTemplate normalizes a template string:
// 1. Optionally replace standalone numbers with wildcard
// 2. Collapse consecutive wildcards
// 3. Trim whitespace
func cleanupTemplate(tmpl, wildcard string, replaceNumbers bool) string {
	// Replace standalone numbers with wildcard only if enabled
	if replaceNumbers {
		tmpl = replaceStandaloneNumbers(tmpl, wildcard)
	}

	// Collapse consecutive wildcards
	wcRe := initWildcardRe(wildcard)
	tmpl = wcRe.ReplaceAllString(tmpl, wildcard)

	return strings.TrimSpace(tmpl)
}

// replaceStandaloneNumbers replaces tokens that are pure numbers with the wildcard.
func replaceStandaloneNumbers(s, wildcard string) string {
	tokens := strings.Fields(s)
	var b strings.Builder
	for i, tok := range tokens {
		if i > 0 {
			b.WriteByte(' ')
		}
		if isNumericToken(tok) {
			b.WriteString(wildcard)
		} else {
			b.WriteString(tok)
		}
	}
	return b.String()
}

// isNumericToken returns true if the token looks like a standalone number
// (digits, dots, underscores, optional leading minus).
func isNumericToken(s string) bool {
	if s == "" {
		return false
	}
	start := 0
	if s[0] == '-' {
		start = 1
	}
	if start >= len(s) {
		return false
	}
	hasDigit := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			hasDigit = true
		} else if c != '.' && c != '_' {
			return false
		}
	}
	return hasDigit
}

// normalizeTemplate prepares a template for comparison by collapsing
// consecutive wildcards and trimming whitespace.
func normalizeTemplate(tmpl, wildcard string) string {
	wcRe := initWildcardRe(wildcard)
	tmpl = wcRe.ReplaceAllString(tmpl, wildcard)
	return strings.TrimSpace(tmpl)
}

// mergeGroupsWithSimilarTemplates merges groups that produced identical
// (normalized) templates into LogTemplate entries.
func mergeGroupsWithSimilarTemplates(groups []*LogGroup, wildcard string) []*LogTemplate {
	templateMap := make(map[string]*LogTemplate)
	var templateOrder []string

	for _, g := range groups {
		normalized := normalizeTemplate(g.Template, wildcard)

		if lt, ok := templateMap[normalized]; ok {
			lt.EventIDs = append(lt.EventIDs, g.EventID)
			lt.Count += len(g.Events)
		} else {
			lt := &LogTemplate{
				TemplateID: intToStr(len(templateMap) + 1),
				Template:   normalized,
				EventIDs:   []string{g.EventID},
				Count:      len(g.Events),
			}
			templateMap[normalized] = lt
			templateOrder = append(templateOrder, normalized)
		}
	}

	// Return templates in discovery order
	templates := make([]*LogTemplate, 0, len(templateOrder))
	for _, key := range templateOrder {
		templates = append(templates, templateMap[key])
	}

	return templates
}
