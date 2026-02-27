package ulp

import "testing"

func TestParseHeaderFormat(t *testing.T) {
	tests := []struct {
		name         string
		format       string
		contentField string
		wantFields   int
		wantErr      bool
	}{
		{
			name:         "HDFS format",
			format:       "<Date> <Time> <Pid> <Level> <Component>: <Content>",
			contentField: "Content",
			wantFields:   6,
		},
		{
			name:         "simple format",
			format:       "<Level> <Content>",
			contentField: "Content",
			wantFields:   2,
		},
		{
			name:         "empty format",
			format:       "",
			contentField: "Content",
			wantErr:      true,
		},
		{
			name:         "unclosed bracket",
			format:       "<Date <Time>",
			contentField: "Content",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hf, err := parseHeaderFormat(tt.format, tt.contentField)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHeaderFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(hf.fields) != tt.wantFields {
				t.Errorf("got %d fields, want %d", len(hf.fields), tt.wantFields)
			}
		})
	}
}

func TestExtractContent(t *testing.T) {
	tests := []struct {
		name   string
		format string
		line   string
		want   string
	}{
		{
			name:   "HDFS log line",
			format: "<Date> <Time> <Pid> <Level> <Component>: <Content>",
			line:   "081109 203615 148 INFO dfs.DataNode$PacketResponder: PacketResponder 0 for block blk_38865049064139660 terminating",
			want:   "PacketResponder 0 for block blk_38865049064139660 terminating",
		},
		{
			name:   "no header format",
			format: "",
			line:   "some raw log line",
			want:   "some raw log line",
		},
		{
			name:   "simple level+content",
			format: "<Level> <Content>",
			line:   "INFO Something happened here",
			want:   "Something happened here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []Option
			if tt.format != "" {
				opts = append(opts, WithHeaderFormat(tt.format))
			}
			p, err := New(opts...)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			got := p.extractContent(tt.line)
			if got != tt.want {
				t.Errorf("extractContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRemovePunctuation(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello!world", "helloworld"},
		{"test@email#hash", "testemailhash"},
		{"no_punct_here-ok", "no_punct_here-ok"},
		{"curly{brace}", "curlybrace"},
		{"normal text", "normal text"},
		{"$100%", "100"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := removePunctuation(tt.input)
			if got != tt.want {
				t.Errorf("removePunctuation(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPreprocess(t *testing.T) {
	p, _ := New()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "hex value",
			input: "block 0xDEADBEEF loaded",
			want:  "block <*> loaded",
		},
		{
			name:  "date replacement",
			input: "event on 2023-01-15 occurred",
			want:  "event on <*> occurred",
		},
		{
			name:  "time replacement",
			input: "at 14:30:45 happened",
			want:  "at <*> happened",
		},
		{
			name:  "bracket normalization",
			input: "key=value func(arg)",
			want:  "key = value func ( arg )",
		},
		{
			name:  "punctuation removal",
			input: "error! at {line} <pos>",
			want:  "error at line pos",
		},
		{
			name:  "MAC address",
			input: "device aa:bb:cc:dd:ee:ff connected",
			want:  "device <*> connected",
		},
		{
			name:  "HTTP URL",
			input: "request to http://example.com/api/path",
			want:  "request to <*>",
		},
		{
			name:  "HTTPS URL",
			input: "visit https://secure.example.com:8443/resource",
			want:  "visit <*>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.preprocess(tt.input)
			if got != tt.want {
				t.Errorf("preprocess(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCollapseSpaces(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"a  b", "a b"},
		{"a   b   c", "a b c"},
		{"no extra spaces", "no extra spaces"},
		{"tab\there", "tab here"},
	}

	for _, tt := range tests {
		got := collapseSpaces(tt.input)
		if got != tt.want {
			t.Errorf("collapseSpaces(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
