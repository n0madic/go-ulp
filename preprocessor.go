package ulp

import (
	"fmt"
	"regexp"
	"strings"
)

// defaultRegexPatterns are built-in patterns for obvious dynamic tokens,
// ordered from most specific to most general.
var defaultRegexPatterns = []*regexp.Regexp{
	regexp.MustCompile(`([\da-fA-F]{2}:){5}[\da-fA-F]{2}`),            // MAC address
	regexp.MustCompile(`\d{4}-\d{2}-\d{2}`),                           // Date YYYY-MM-DD
	regexp.MustCompile(`\d{4}/\d{2}/\d{2}`),                           // Date YYYY/MM/DD
	regexp.MustCompile(`[0-9]{2}:[0-9]{2}:[0-9]{2}(?:[.,][0-9]{3})?`), // Time HH:MM:SS[.mmm]
	regexp.MustCompile(`0[xX][0-9a-fA-F]+`),                           // Hex value
	regexp.MustCompile(`([0-9a-fA-F]*:){8,}`),                         // IPv6-like
	regexp.MustCompile(`https?://\S+`),                                // HTTP/HTTPS URL
	regexp.MustCompile(`(/?)([a-zA-Z0-9-]+\.){2,}([a-zA-Z0-9-]+)?`),   // Domain/host
}

// punctuationToRemove are characters stripped during preprocessing.
var punctuationToRemove = "!@#$%^&{}<>?\\|`~"

// bracketNormRe matches brackets and equals signs that need surrounding spaces.
var bracketNormRe = regexp.MustCompile(`([=\(\)\[\]])`)

// parseHeaderFormat parses a header format string like "<Date> <Time> <Level> <Content>"
// into a structured HeaderFormat.
func parseHeaderFormat(format, contentField string) (*HeaderFormat, error) {
	if format == "" {
		return nil, fmt.Errorf("header format cannot be empty")
	}

	hf := &HeaderFormat{
		Format:       format,
		ContentField: contentField,
	}

	remaining := format
	for remaining != "" {
		// Find the next field marker <FieldName>
		start := strings.Index(remaining, "<")
		if start == -1 {
			break
		}
		end := strings.Index(remaining[start+1:], ">")
		if end == -1 {
			return nil, fmt.Errorf("unclosed field marker in format: %s", format)
		}
		end += start + 1

		fieldName := remaining[start+1 : end]
		if strings.Contains(fieldName, "<") {
			return nil, fmt.Errorf("unclosed field marker in format: %s", format)
		}
		remaining = remaining[end+1:]

		field := headerField{name: fieldName}

		// Find separator: everything up to the next "<" or end of string
		nextField := strings.Index(remaining, "<")
		if nextField == -1 {
			field.separator = remaining
			remaining = ""
		} else {
			field.separator = remaining[:nextField]
			remaining = remaining[nextField:]
		}

		hf.fields = append(hf.fields, field)
	}

	if len(hf.fields) == 0 {
		return nil, fmt.Errorf("no fields found in header format: %s", format)
	}

	return hf, nil
}

// extractContent parses a log line using the header format and returns
// the content field value. If no header format is set, returns the whole line.
func (p *Parser) extractContent(line string) string {
	if p.headerFormat == nil {
		return line
	}

	remaining := line
	for i, field := range p.headerFormat.fields {
		if field.name == p.contentField {
			// This is the content field — return everything remaining
			return strings.TrimSpace(remaining)
		}

		// For the last field (or if no separator), consume the rest
		if field.separator == "" || i == len(p.headerFormat.fields)-1 {
			break
		}

		// Find the separator in the remaining text
		sepIdx := strings.Index(remaining, field.separator)
		if sepIdx == -1 {
			// Separator not found; fall back to returning everything
			return strings.TrimSpace(remaining)
		}
		remaining = remaining[sepIdx+len(field.separator):]
	}

	return strings.TrimSpace(remaining)
}

// preprocess applies all preprocessing steps to a raw content string:
// 1. Remove punctuation
// 2. Replace obvious dynamic tokens via regex
// 3. Normalize brackets
func (p *Parser) preprocess(content string) string {
	// Step 1: Remove punctuation characters
	s := removePunctuation(content)

	// Step 2: Replace obvious dynamic tokens with wildcard
	s = p.replaceByRegex(s)

	// Step 3: Normalize brackets — add spaces around = ( ) [ ]
	s = bracketNormRe.ReplaceAllString(s, " $1 ")

	// Collapse multiple spaces
	s = collapseSpaces(s)

	return strings.TrimSpace(s)
}

// removePunctuation strips characters in punctuationToRemove from the string.
func removePunctuation(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if !strings.ContainsRune(punctuationToRemove, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// replaceByRegex replaces tokens matching default and custom regex patterns
// with the dynamic wildcard.
func (p *Parser) replaceByRegex(s string) string {
	for _, re := range defaultRegexPatterns {
		s = re.ReplaceAllString(s, p.dynamicWildcard)
	}
	for _, re := range p.customRegex {
		s = re.ReplaceAllString(s, p.dynamicWildcard)
	}
	return s
}

// collapseSpaces replaces runs of whitespace with a single space.
func collapseSpaces(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
		} else {
			b.WriteRune(r)
			prevSpace = false
		}
	}
	return b.String()
}
