package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	ulp "github.com/n0madic/go-ulp"
)

func main() {
	headerFormat := flag.String("header-format", "", `Log header format (e.g., "<Date> <Time> <Level> <Content>")`)
	contentField := flag.String("content-field", "Content", "Header field with log message")
	regexStr := flag.String("regex", "", "Additional regex patterns, comma-separated")
	sampleSize := flag.Int("sample-size", 0, "Max events sampled per group, 0=all")
	workers := flag.Int("workers", 0, "Worker goroutines, 0=auto")
	format := flag.String("format", "csv", "Output format: csv, json, text")
	templatesOnly := flag.Bool("templates-only", false, "Output only unique templates")
	output := flag.String("output", "", "Output file (default stdout)")
	verbose := flag.Bool("verbose", false, "Show parsing statistics to stderr")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: go-ulp [flags] [INPUT_FILE]\n\n")
		fmt.Fprintf(os.Stderr, "ULP (Unified Log Parser) extracts log templates from unstructured log files.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nIf no INPUT_FILE is specified, reads from stdin.\n")
	}

	flag.Parse()

	// Build parser options
	var opts []ulp.Option

	if *contentField != "Content" {
		opts = append(opts, ulp.WithContentField(*contentField))
	}
	if *headerFormat != "" {
		opts = append(opts, ulp.WithHeaderFormat(*headerFormat))
	}
	if *regexStr != "" {
		patterns := strings.Split(*regexStr, ",")
		for i := range patterns {
			patterns[i] = strings.TrimSpace(patterns[i])
		}
		opts = append(opts, ulp.WithCustomRegex(patterns))
	}
	if *sampleSize > 0 {
		opts = append(opts, ulp.WithSampleSize(*sampleSize))
	}
	if *workers > 0 {
		opts = append(opts, ulp.WithMaxWorkers(*workers))
	}

	parser, err := ulp.New(opts...)
	if err != nil {
		log.Fatalf("Error creating parser: %v", err)
	}

	// Determine input source
	var input *os.File
	args := flag.Args()
	if len(args) > 0 {
		input, err = os.Open(args[0])
		if err != nil {
			log.Fatalf("Error opening input file: %v", err)
		}
		defer input.Close()
	} else {
		input = os.Stdin
	}

	// Parse
	result, err := parser.Parse(input)
	if err != nil {
		log.Fatalf("Error parsing: %v", err)
	}

	// Determine output destination
	var out *os.File
	if *output != "" {
		out, err = os.Create(*output)
		if err != nil {
			log.Fatalf("Error creating output file: %v", err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	// Sort templates by frequency (descending)
	sort.Slice(result.Templates, func(i, j int) bool {
		return result.Templates[i].Count > result.Templates[j].Count
	})

	// Write output
	if *templatesOnly {
		err = writeTemplates(out, result, *format)
	} else {
		err = writeEvents(out, result, *format)
	}
	if err != nil {
		log.Fatalf("Error writing output: %v", err)
	}

	// Verbose stats to stderr
	if *verbose {
		fmt.Fprintf(os.Stderr, "Lines:     %d\n", len(result.Events))
		fmt.Fprintf(os.Stderr, "Templates: %d\n", len(result.Templates))
		fmt.Fprintf(os.Stderr, "Groups:    %d\n", len(result.Groups))
		fmt.Fprintf(os.Stderr, "Duration:  %v\n", result.Duration)
	}
}
