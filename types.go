package ulp

import "time"

// LogEvent represents a single log line after preprocessing.
type LogEvent struct {
	LineID      int
	RawContent  string // original message content (header removed)
	TokenString string // preprocessed token string
	EventID     string // group identifier
	TemplateID  string // final template identifier
}

// LogGroup represents a cluster of events sharing the same EventID.
type LogGroup struct {
	EventID  string
	Events   []*LogEvent
	Template string
}

// LogTemplate represents a unique log template after merging groups.
type LogTemplate struct {
	TemplateID string
	Template   string
	EventIDs   []string // EventIDs that share this template
	Count      int      // total number of events matching this template
}

// ParseResult holds the complete output of the parsing process.
type ParseResult struct {
	Events    []*LogEvent
	Templates []*LogTemplate
	Groups    []*LogGroup
	Duration  time.Duration
}

// HeaderFormat describes how to parse the log header.
// Fields are extracted by name from the format string, e.g. "<Date> <Time> <Pid> <Level> <Component>: <Content>".
type HeaderFormat struct {
	Format       string
	ContentField string
	fields       []headerField
}

type headerField struct {
	name      string
	separator string // separator after this field (empty for last field)
}
