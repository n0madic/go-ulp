package ulp

import "strings"

// getVocabulary builds a frequency map of tokens across a set of events,
// counting each token at most once per event (deduplication within an event).
// This implements the paper's CountTokensAvoidDuplicatePerEvent approach.
func getVocabulary(events []*LogEvent) map[string]int {
	vocab := make(map[string]int)
	for _, ev := range events {
		tokens := strings.Fields(ev.TokenString)
		seen := make(map[string]struct{}, len(tokens))
		for _, tok := range tokens {
			if _, ok := seen[tok]; !ok {
				seen[tok] = struct{}{}
				vocab[tok]++
			}
		}
	}
	return vocab
}

// findDynamicTokens identifies tokens in a template that appear in fewer events
// than the group size, meaning they are dynamic (variable) tokens.
// Per the paper's Algorithm 1: if token_count < group_length => dynamic.
func findDynamicTokens(vocab map[string]int, groupSize int) map[string]bool {
	dynamic := make(map[string]bool)
	for token, count := range vocab {
		if count < groupSize {
			dynamic[token] = true
		}
	}
	return dynamic
}

// sampleEvents selects a uniform sample of events from a group.
// If sampleSize is 0 or >= len(events), returns all events.
// Uses deterministic uniform spacing (not random) for reproducibility.
func sampleEvents(events []*LogEvent, sampleSize int) []*LogEvent {
	n := len(events)
	if sampleSize <= 0 || sampleSize >= n {
		return events
	}

	sampled := make([]*LogEvent, 0, sampleSize)
	step := float64(n) / float64(sampleSize)
	for i := 0; i < sampleSize; i++ {
		idx := int(float64(i) * step)
		sampled = append(sampled, events[idx])
	}
	return sampled
}
