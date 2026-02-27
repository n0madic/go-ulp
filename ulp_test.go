package ulp

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
)

func TestParseHDFSSample(t *testing.T) {
	f, err := os.Open("testdata/hdfs_sample.log")
	if err != nil {
		t.Fatalf("failed to open test data: %v", err)
	}
	defer f.Close()

	p, err := New(
		WithHeaderFormat("<Date> <Time> <Pid> <Level> <Component>: <Content>"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := p.Parse(f)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Events) != 12 {
		t.Errorf("expected 12 events, got %d", len(result.Events))
	}

	if len(result.Templates) != 3 {
		t.Errorf("expected 3 templates, got %d", len(result.Templates))
	}

	// Verify expected templates exist
	expectedPatterns := []string{
		"PacketResponder",
		"NameSystem.addStoredBlock",
		"Received block",
	}

	templateTexts := make([]string, 0, len(result.Templates))
	for _, tmpl := range result.Templates {
		templateTexts = append(templateTexts, tmpl.Template)
	}

	for _, pattern := range expectedPatterns {
		found := false
		for _, tmpl := range templateTexts {
			if strings.Contains(tmpl, pattern) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected template containing %q not found in %v", pattern, templateTexts)
		}
	}

	// Verify event counts per template
	sort.Slice(result.Templates, func(i, j int) bool {
		return result.Templates[i].Count > result.Templates[j].Count
	})

	// PacketResponder: 6 events (lines 1,2,4,6,9,12)
	// Received block: 3 events (lines 3,7,10)
	// addStoredBlock: 3 events (lines 5,8,11)
	if result.Templates[0].Count != 6 {
		t.Errorf("largest template count = %d, want 6", result.Templates[0].Count)
	}

	t.Logf("Parse took %v", result.Duration)
	for _, tmpl := range result.Templates {
		t.Logf("Template %s (count=%d): %s", tmpl.TemplateID, tmpl.Count, tmpl.Template)
	}
}

func TestParseNoHeader(t *testing.T) {
	input := `error connecting to database host=db1 port=5432
error connecting to database host=db2 port=5433
error connecting to database host=db3 port=5432
server started on port 8080
server started on port 9090
`

	p, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := p.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Events) != 5 {
		t.Errorf("expected 5 events, got %d", len(result.Events))
	}

	t.Logf("Templates: %d", len(result.Templates))
	for _, tmpl := range result.Templates {
		t.Logf("  [%d events] %s", tmpl.Count, tmpl.Template)
	}
}

func TestParseEmpty(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := p.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Events) != 0 {
		t.Errorf("expected 0 events, got %d", len(result.Events))
	}
	if len(result.Templates) != 0 {
		t.Errorf("expected 0 templates, got %d", len(result.Templates))
	}
}

func TestParseWithSampling(t *testing.T) {
	// Generate many similar lines
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, fmt.Sprintf("request %d from user_%d completed in %dms", i, i%10, i*3))
	}

	p, err := New(WithSampleSize(10))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := p.Parse(strings.NewReader(strings.Join(lines, "\n")))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Events) != 100 {
		t.Errorf("expected 100 events, got %d", len(result.Events))
	}

	t.Logf("Templates with sampling: %d", len(result.Templates))
	for _, tmpl := range result.Templates {
		t.Logf("  [%d events] %s", tmpl.Count, tmpl.Template)
	}
}

func TestParseWithCustomRegex(t *testing.T) {
	input := `session ABC123 started
session DEF456 started
session GHI789 started
`

	p, err := New(WithCustomRegex([]string{`[A-Z]{3}\d{3}`}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := p.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Templates) != 1 {
		t.Errorf("expected 1 template, got %d", len(result.Templates))
	}

	if len(result.Templates) > 0 && !strings.Contains(result.Templates[0].Template, "<*>") {
		t.Errorf("template should contain <*>, got: %s", result.Templates[0].Template)
	}
}

func TestParseWithReplaceNumbers(t *testing.T) {
	input := `error at line 42 in module
error at line 99 in module
`
	t.Run("disabled by default", func(t *testing.T) {
		p, err := New()
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		result, err := p.Parse(strings.NewReader(input))
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		// With frequency analysis, the varying numbers become <*>
		// but standalone numbers in single-event groups are preserved
		for _, tmpl := range result.Templates {
			t.Logf("  [%d events] %s", tmpl.Count, tmpl.Template)
		}
	})

	t.Run("enabled replaces standalone numbers", func(t *testing.T) {
		p, err := New(WithReplaceNumbers(true))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		result, err := p.Parse(strings.NewReader("single line with number 42\n"))
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if len(result.Templates) != 1 {
			t.Fatalf("expected 1 template, got %d", len(result.Templates))
		}

		tmpl := result.Templates[0].Template
		if !strings.Contains(tmpl, "<*>") {
			t.Errorf("expected template to contain <*> for standalone number, got: %s", tmpl)
		}
		if strings.Contains(tmpl, "42") {
			t.Errorf("expected standalone number 42 to be replaced, got: %s", tmpl)
		}
	})
}

func TestParseSingleWorker(t *testing.T) {
	input := `msg A 1
msg B 2
msg A 3
`
	p, err := New(WithMaxWorkers(1))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := p.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Events) != 3 {
		t.Errorf("expected 3 events, got %d", len(result.Events))
	}
}

// Benchmarks

func BenchmarkParse12Lines(b *testing.B) {
	data, err := os.ReadFile("testdata/hdfs_sample.log")
	if err != nil {
		b.Fatalf("failed to read test data: %v", err)
	}

	p, _ := New(
		WithHeaderFormat("<Date> <Time> <Pid> <Level> <Component>: <Content>"),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(strings.NewReader(string(data)))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse1000Lines(b *testing.B) {
	var lines []string
	templates := []string{
		"Connection from %s established on port %d",
		"Request %s processed in %dms",
		"Error: failed to connect to %s:%d",
		"User %s logged in from %s",
		"Cache hit for key %s (ttl=%d)",
	}

	for i := 0; i < 1000; i++ {
		tmpl := templates[i%len(templates)]
		lines = append(lines, fmt.Sprintf(tmpl, fmt.Sprintf("val_%d", i), i*7))
	}
	input := strings.Join(lines, "\n")

	p, _ := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(strings.NewReader(input))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse10000Lines(b *testing.B) {
	var lines []string
	templates := []string{
		"PacketResponder %d for block blk_%d terminating",
		"Received block blk_%d of size 67108864 from /10.251.%d.%d",
		"BLOCK* NameSystem.addStoredBlock: blockMap updated: 10.250.%d.%d:50010 is added to blk_%d size 67108864",
	}

	for i := 0; i < 10000; i++ {
		tmpl := templates[i%len(templates)]
		lines = append(lines, fmt.Sprintf(tmpl, i, i*31, i*17))
	}
	input := strings.Join(lines, "\n")

	p, _ := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(strings.NewReader(input))
		if err != nil {
			b.Fatal(err)
		}
	}
}
