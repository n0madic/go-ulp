package ulp

import "testing"

func TestGetVocabulary(t *testing.T) {
	events := []*LogEvent{
		{TokenString: "PacketResponder 0 for block blk_111 terminating"},
		{TokenString: "PacketResponder 2 for block blk_222 terminating"},
		{TokenString: "PacketResponder 1 for block blk_333 terminating"},
	}

	vocab := getVocabulary(events)

	// Static tokens should appear in all 3 events
	for _, tok := range []string{"PacketResponder", "for", "block", "terminating"} {
		if vocab[tok] != 3 {
			t.Errorf("vocab[%q] = %d, want 3", tok, vocab[tok])
		}
	}

	// Dynamic tokens appear in 1 event each
	for _, tok := range []string{"0", "2", "1", "blk_111", "blk_222", "blk_333"} {
		if vocab[tok] != 1 {
			t.Errorf("vocab[%q] = %d, want 1", tok, vocab[tok])
		}
	}
}

func TestGetVocabularyDedup(t *testing.T) {
	// Token "error" appears twice in the same event — should count once
	events := []*LogEvent{
		{TokenString: "error error occurred"},
		{TokenString: "error in module"},
	}

	vocab := getVocabulary(events)

	if vocab["error"] != 2 {
		t.Errorf("vocab[error] = %d, want 2 (dedup per event)", vocab["error"])
	}
}

func TestFindDynamicTokens(t *testing.T) {
	vocab := map[string]int{
		"PacketResponder": 5,
		"for":             5,
		"block":           5,
		"terminating":     5,
		"0":               1,
		"blk_111":         1,
		"2":               2,
	}

	dynamic := findDynamicTokens(vocab, 5)

	// Tokens with count < 5 are dynamic
	if !dynamic["0"] {
		t.Error("expected '0' to be dynamic")
	}
	if !dynamic["blk_111"] {
		t.Error("expected 'blk_111' to be dynamic")
	}
	if !dynamic["2"] {
		t.Error("expected '2' to be dynamic")
	}

	// Tokens with count == groupSize are static
	if dynamic["PacketResponder"] {
		t.Error("expected 'PacketResponder' to be static")
	}
	if dynamic["for"] {
		t.Error("expected 'for' to be static")
	}
}

func TestSampleEvents(t *testing.T) {
	events := make([]*LogEvent, 100)
	for i := range events {
		events[i] = &LogEvent{LineID: i}
	}

	// Sample size 0 returns all
	sampled := sampleEvents(events, 0)
	if len(sampled) != 100 {
		t.Errorf("sampleSize=0: got %d, want 100", len(sampled))
	}

	// Sample size larger than events returns all
	sampled = sampleEvents(events, 200)
	if len(sampled) != 100 {
		t.Errorf("sampleSize=200: got %d, want 100", len(sampled))
	}

	// Sample size 10 returns exactly 10
	sampled = sampleEvents(events, 10)
	if len(sampled) != 10 {
		t.Errorf("sampleSize=10: got %d, want 10", len(sampled))
	}

	// Verify uniform distribution (first and roughly last elements)
	if sampled[0].LineID != 0 {
		t.Errorf("first sample should be index 0, got %d", sampled[0].LineID)
	}
	if sampled[9].LineID != 90 {
		t.Errorf("last sample should be index 90, got %d", sampled[9].LineID)
	}
}

func TestSampleEventsSmall(t *testing.T) {
	events := []*LogEvent{
		{LineID: 0},
		{LineID: 1},
		{LineID: 2},
	}

	// Sample size equal to events returns all
	sampled := sampleEvents(events, 3)
	if len(sampled) != 3 {
		t.Errorf("sampleSize=3: got %d, want 3", len(sampled))
	}
}
