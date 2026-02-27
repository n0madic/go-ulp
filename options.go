package ulp

import (
	"fmt"
	"regexp"
	"runtime"
)

// Parser is the main ULP log parser.
type Parser struct {
	headerFormat    *HeaderFormat
	contentField    string
	customRegex     []*regexp.Regexp
	sampleSize      int
	maxWorkers      int
	dynamicWildcard string
	replaceNumbers  bool
}

// Option configures the Parser.
type Option func(*Parser) error

// New creates a new Parser with the given options.
func New(opts ...Option) (*Parser, error) {
	p := &Parser{
		contentField:    "Content",
		sampleSize:      0,
		maxWorkers:      runtime.NumCPU(),
		dynamicWildcard: "<*>",
	}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// WithHeaderFormat sets the log header format string.
// Example: "<Date> <Time> <Pid> <Level> <Component>: <Content>"
func WithHeaderFormat(format string) Option {
	return func(p *Parser) error {
		hf, err := parseHeaderFormat(format, p.contentField)
		if err != nil {
			return err
		}
		p.headerFormat = hf
		return nil
	}
}

// WithContentField sets the name of the header field that contains
// the log message body. Default is "Content".
func WithContentField(field string) Option {
	return func(p *Parser) error {
		if field == "" {
			return fmt.Errorf("content field cannot be empty")
		}
		p.contentField = field
		return nil
	}
}

// WithCustomRegex adds additional regex patterns for preprocessing.
// Matched tokens are replaced with the dynamic wildcard.
func WithCustomRegex(patterns []string) Option {
	return func(p *Parser) error {
		for _, pat := range patterns {
			re, err := regexp.Compile(pat)
			if err != nil {
				return fmt.Errorf("invalid regex pattern %q: %w", pat, err)
			}
			p.customRegex = append(p.customRegex, re)
		}
		return nil
	}
}

// WithSampleSize sets the maximum number of events sampled per group
// for frequency analysis. 0 means use all events.
func WithSampleSize(n int) Option {
	return func(p *Parser) error {
		if n < 0 {
			return fmt.Errorf("sample size cannot be negative")
		}
		p.sampleSize = n
		return nil
	}
}

// WithMaxWorkers sets the number of worker goroutines for parallel
// group processing. 0 or negative values default to runtime.NumCPU().
func WithMaxWorkers(n int) Option {
	return func(p *Parser) error {
		if n <= 0 {
			p.maxWorkers = runtime.NumCPU()
		} else {
			p.maxWorkers = n
		}
		return nil
	}
}

// WithDynamicWildcard sets the placeholder string for dynamic tokens.
// Default is "<*>".
func WithDynamicWildcard(w string) Option {
	return func(p *Parser) error {
		if w == "" {
			return fmt.Errorf("dynamic wildcard cannot be empty")
		}
		p.dynamicWildcard = w
		return nil
	}
}

// WithReplaceNumbers controls whether standalone numbers in templates are
// replaced with the dynamic wildcard. Default is false (per the paper,
// dynamic tokens are determined only via frequency analysis).
func WithReplaceNumbers(enable bool) Option {
	return func(p *Parser) error {
		p.replaceNumbers = enable
		return nil
	}
}
