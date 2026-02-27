# go-ulp

A Go implementation of the **ULP (Unified Log Parser)** algorithm from the paper ["An Effective Approach for Parsing Large Log Files"](https://ieeexplore.ieee.org/document/9978308) (ICSME 2022).

ULP extracts log templates from unstructured log files by combining string matching for grouping with local frequency analysis for separating static and dynamic tokens.

## Features

- **Zero external dependencies** — uses only Go standard library
- **Concurrent processing** — worker pool for parallel group analysis
- **Streaming input** — processes logs via `bufio.Scanner`, no full-file loading
- **Configurable header parsing** — flexible log format specification
- **Extensible regex patterns** — built-in + custom patterns for dynamic token detection
- **Multiple output formats** — CSV, JSON, text
- **Library + CLI** — usable as a Go package or standalone command

## Installation

### CLI

```bash
go install github.com/n0madic/go-ulp/cmd/go-ulp@latest
```

### Library

```bash
go get github.com/n0madic/go-ulp
```

## CLI Usage

```
go-ulp [flags] [INPUT_FILE]

Flags:
  -header-format string   Log header format (e.g., "<Date> <Time> <Level> <Content>")
  -content-field string   Header field with log message (default "Content")
  -regex string           Additional regex patterns, comma-separated
  -sample-size int        Max events sampled per group, 0=all (default 0)
  -workers int            Worker goroutines, 0=auto (default 0)
  -format string          Output format: csv, json, text (default "csv")
  -templates-only         Output only unique templates
  -output string          Output file (default stdout)
  -verbose                Show parsing statistics to stderr
```

### Examples

Extract templates from HDFS logs:
```bash
go-ulp -header-format '<Date> <Time> <Pid> <Level> <Component>: <Content>' \
       -templates-only -format text hdfs.log
```

Parse syslog from stdin:
```bash
cat /var/log/syslog | go-ulp -format json -templates-only
```

Output (sorted by frequency):
```
(6 events) PacketResponder <*> for block <*> terminating
(3 events) Received block <*> of size 67108864 from <*>
(3 events) BLOCK* NameSystem.addStoredBlock: blockMap updated: <*>:50010 is added to <*> size 67108864
```

## Library Usage

```go
package main

import (
    "fmt"
    "os"

    ulp "github.com/n0madic/go-ulp"
)

func main() {
    parser, _ := ulp.New(
        ulp.WithHeaderFormat("<Date> <Time> <Pid> <Level> <Component>: <Content>"),
        ulp.WithMaxWorkers(4),
    )

    f, _ := os.Open("hdfs.log")
    defer f.Close()

    result, _ := parser.Parse(f)

    for _, tmpl := range result.Templates {
        fmt.Printf("[%d events] %s\n", tmpl.Count, tmpl.Template)
    }
}
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithHeaderFormat(format)` | Log header format string | none (whole line is content) |
| `WithContentField(field)` | Name of the content field in header | `"Content"` |
| `WithCustomRegex(patterns)` | Additional regex patterns for preprocessing | none |
| `WithSampleSize(n)` | Max events sampled per group (0=all) | `0` |
| `WithMaxWorkers(n)` | Worker goroutines (0=NumCPU) | `runtime.NumCPU()` |
| `WithDynamicWildcard(w)` | Placeholder for dynamic tokens | `"<*>"` |
| `WithReplaceNumbers(bool)` | Replace standalone numbers with wildcard | `false` |

## Algorithm Overview

1. **Preprocessing**: Remove headers, strip punctuation, replace obvious dynamic tokens (IPs, dates, hex values, etc.) with wildcards
2. **Grouping**: Generate EventIDs from alphabetic tokens + word count, group events by EventID
3. **Frequency Analysis**: Within each group, count token frequency (deduplicated per event). Tokens not present in all events are dynamic
4. **Template Generation**: Replace dynamic tokens with wildcards, collapse consecutive wildcards. Optionally replace standalone numbers (`WithReplaceNumbers(true)`)
5. **Merging**: Merge groups that produce identical templates

## Built-in Regex Patterns

The preprocessor automatically detects and replaces:
- MAC addresses (`aa:bb:cc:dd:ee:ff`)
- Dates (`YYYY-MM-DD`, `YYYY/MM/DD`)
- Times (`HH:MM:SS`, with optional milliseconds)
- Hex values (`0xDEADBEEF`)
- IPv6 addresses
- HTTP/HTTPS URLs
- Domain names / hostnames

## License

MIT
