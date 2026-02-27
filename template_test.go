package ulp

import "testing"

func TestGenerateTemplate(t *testing.T) {
	tests := []struct {
		name           string
		events         []*LogEvent
		wildcard       string
		replaceNumbers bool
		want           string
	}{
		{
			name: "PacketResponder group",
			events: []*LogEvent{
				{TokenString: "PacketResponder 0 for block blk_111 terminating"},
				{TokenString: "PacketResponder 2 for block blk_222 terminating"},
				{TokenString: "PacketResponder 1 for block blk_333 terminating"},
			},
			wildcard: "<*>",
			want:     "PacketResponder <*> for block <*> terminating",
		},
		{
			name: "single event",
			events: []*LogEvent{
				{TokenString: "error occurred at line 42"},
			},
			wildcard: "<*>",
			want:     "error occurred at line 42",
		},
		{
			name:     "empty group",
			events:   []*LogEvent{},
			wildcard: "<*>",
			want:     "",
		},
		{
			name: "all static tokens",
			events: []*LogEvent{
				{TokenString: "server started successfully"},
				{TokenString: "server started successfully"},
			},
			wildcard: "<*>",
			want:     "server started successfully",
		},
		{
			name: "single event with replaceNumbers enabled",
			events: []*LogEvent{
				{TokenString: "error occurred at line 42"},
			},
			wildcard:       "<*>",
			replaceNumbers: true,
			want:           "error occurred at line <*>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &LogGroup{Events: tt.events}
			got := generateTemplate(group, tt.wildcard, 0, tt.replaceNumbers)
			if got != tt.want {
				t.Errorf("generateTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsNumericToken(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"42", true},
		{"3.14", true},
		{"-1", true},
		{"100_000", true},
		{"abc", false},
		{"12abc", false},
		{"", false},
		{"-", false},
		{"0x1F", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isNumericToken(tt.input)
			if got != tt.want {
				t.Errorf("isNumericToken(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeTemplate(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello <*> <*> world", "hello <*> world"},
		{"<*> <*> <*>", "<*>"},
		{"no wildcards here", "no wildcards here"},
		{"  extra  spaces  ", "extra  spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeTemplate(tt.input, "<*>")
			if got != tt.want {
				t.Errorf("normalizeTemplate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMergeGroupsWithSimilarTemplates(t *testing.T) {
	groups := []*LogGroup{
		{EventID: "A", Template: "error <*> at line <*>", Events: make([]*LogEvent, 5)},
		{EventID: "B", Template: "error <*> at line <*>", Events: make([]*LogEvent, 3)},
		{EventID: "C", Template: "server started", Events: make([]*LogEvent, 2)},
	}

	templates := mergeGroupsWithSimilarTemplates(groups, "<*>")

	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}

	// First template should merge A and B
	if templates[0].Count != 8 {
		t.Errorf("merged template count = %d, want 8", templates[0].Count)
	}
	if len(templates[0].EventIDs) != 2 {
		t.Errorf("merged template eventIDs = %d, want 2", len(templates[0].EventIDs))
	}

	// Second template is C
	if templates[1].Count != 2 {
		t.Errorf("second template count = %d, want 2", templates[1].Count)
	}
}

func TestCleanupTemplate(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		replaceNumbers bool
		want           string
	}{
		{
			name:  "standalone number not replaced by default",
			input: "block 123 loaded",
			want:  "block 123 loaded",
		},
		{
			name:           "standalone number replaced when enabled",
			input:          "block 123 loaded",
			replaceNumbers: true,
			want:           "block <*> loaded",
		},
		{
			name:  "consecutive wildcards collapsed",
			input: "error <*> <*> occurred",
			want:  "error <*> occurred",
		},
		{
			name:           "both cleanup steps when replaceNumbers enabled",
			input:          "value 42 <*> 99 end",
			replaceNumbers: true,
			want:           "value <*> end",
		},
		{
			name:  "numbers preserved when replaceNumbers disabled",
			input: "value 42 <*> 99 end",
			want:  "value 42 <*> 99 end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanupTemplate(tt.input, "<*>", tt.replaceNumbers)
			if got != tt.want {
				t.Errorf("cleanupTemplate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
