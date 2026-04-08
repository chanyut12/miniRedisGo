package protocol

import "testing"

func TestLineParserParseLine(t *testing.T) {
	parser := NewLineParser()

	cmd, err := parser.ParseLine("  ping   ")
	if err != nil {
		t.Fatalf("ParseLine() unexpected error: %v", err)
	}

	if cmd.Name != "PING" {
		t.Fatalf("ParseLine() name = %q, want %q", cmd.Name, "PING")
	}

	if len(cmd.Args) != 0 {
		t.Fatalf("ParseLine() args len = %d, want 0", len(cmd.Args))
	}

	if cmd.Raw != "ping" {
		t.Fatalf("ParseLine() raw = %q, want %q", cmd.Raw, "ping")
	}
}

func TestLineParserParseLineEmpty(t *testing.T) {
	parser := NewLineParser()

	if _, err := parser.ParseLine("   "); err == nil {
		t.Fatal("ParseLine() error = nil, want non-nil")
	}
}
