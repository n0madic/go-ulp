package ulp

import (
	"bufio"
	"io"
	"sort"
	"sync"
	"time"
)

// Parse reads log lines from r and returns parsed results with templates.
// This implements Algorithm 1 from the ULP paper (ICSME 2022).
func (p *Parser) Parse(r io.Reader) (*ParseResult, error) {
	start := time.Now()

	// Step 1: Read and preprocess all lines
	events, err := p.readAndPreprocess(r)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return &ParseResult{Duration: time.Since(start)}, nil
	}

	// Step 2: Generate EventIDs and group events
	for _, ev := range events {
		ev.EventID = generateEventID(ev.TokenString)
	}
	groupMap := groupEvents(events)

	// Convert to slice for deterministic ordering
	groups := make([]*LogGroup, 0, len(groupMap))
	for _, g := range groupMap {
		groups = append(groups, g)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Events[0].LineID < groups[j].Events[0].LineID
	})

	// Step 3: Generate templates (parallel via worker pool)
	p.generateTemplatesParallel(groups)

	// Step 4: Merge groups with similar templates
	templates := mergeGroupsWithSimilarTemplates(groups, p.dynamicWildcard)

	// Assign TemplateIDs back to events
	templateByEventID := make(map[string]string)
	for _, tmpl := range templates {
		for _, eid := range tmpl.EventIDs {
			templateByEventID[eid] = tmpl.TemplateID
		}
	}
	for _, ev := range events {
		ev.TemplateID = templateByEventID[ev.EventID]
	}

	return &ParseResult{
		Events:    events,
		Templates: templates,
		Groups:    groups,
		Duration:  time.Since(start),
	}, nil
}

// readAndPreprocess reads log lines from r and creates preprocessed LogEvents.
func (p *Parser) readAndPreprocess(r io.Reader) ([]*LogEvent, error) {
	scanner := bufio.NewScanner(r)
	// Allow long lines (up to 1MB)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var events []*LogEvent
	lineID := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		lineID++

		content := p.extractContent(line)
		tokenString := p.preprocess(content)

		events = append(events, &LogEvent{
			LineID:      lineID,
			RawContent:  content,
			TokenString: tokenString,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// generateTemplatesParallel processes groups through a worker pool.
func (p *Parser) generateTemplatesParallel(groups []*LogGroup) {
	if len(groups) == 0 {
		return
	}

	workers := p.maxWorkers
	if workers > len(groups) {
		workers = len(groups)
	}

	ch := make(chan *LogGroup, len(groups))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Go(func() {
			for g := range ch {
				g.Template = generateTemplate(g, p.dynamicWildcard, p.sampleSize, p.replaceNumbers)
			}
		})
	}

	// Send groups to workers
	for _, g := range groups {
		ch <- g
	}
	close(ch)

	wg.Wait()
}
