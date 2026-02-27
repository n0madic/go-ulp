package ulp

import "testing"

func TestGenerateEventID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "PacketResponder line",
			input: "PacketResponder 0 for block blk_38865049064139660 terminating",
			want:  "PacketResponderforblockterminating6",
		},
		{
			name:  "single word",
			input: "error",
			want:  "error1",
		},
		{
			name:  "all numbers",
			input: "123 456 789",
			want:  "3",
		},
		{
			name:  "mixed tokens",
			input: "GET /api/v2 200 OK",
			want:  "GETOK4",
		},
		{
			name:  "empty string",
			input: "",
			want:  "0",
		},
		{
			name:  "single char words included",
			input: "a b c test",
			want:  "abctest4",
		},
		{
			name:  "wildcard tokens counted in length",
			input: "error <*> at line <*>",
			want:  "erroratline5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateEventID(tt.input)
			if got != tt.want {
				t.Errorf("generateEventID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsAlpha(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"hello", true},
		{"Hello", true},
		{"test123", false},
		{"123", false},
		{"blk_123", false},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isAlpha(tt.input)
			if got != tt.want {
				t.Errorf("isAlpha(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGroupEvents(t *testing.T) {
	events := []*LogEvent{
		{LineID: 1, EventID: "A"},
		{LineID: 2, EventID: "B"},
		{LineID: 3, EventID: "A"},
		{LineID: 4, EventID: "A"},
		{LineID: 5, EventID: "B"},
	}

	groups := groupEvents(events)

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if len(groups["A"].Events) != 3 {
		t.Errorf("group A: expected 3 events, got %d", len(groups["A"].Events))
	}
	if len(groups["B"].Events) != 2 {
		t.Errorf("group B: expected 2 events, got %d", len(groups["B"].Events))
	}
}

func TestRemoveSpecialChars(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello(world)", "helloworld"},
		{"key=value", "keyvalue"},
		{"no-dash", "nodash"},
		{"test+1", "test1"},
		{"plain", "plain"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := removeSpecialChars(tt.input)
			if got != tt.want {
				t.Errorf("removeSpecialChars(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIntToStr(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{100, "100"},
		{12345, "12345"},
	}

	for _, tt := range tests {
		got := intToStr(tt.input)
		if got != tt.want {
			t.Errorf("intToStr(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
