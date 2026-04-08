package protocol

import (
	"fmt"
	"strings"
)

// Command is the parsed form of a raw line from a client connection.
type Command struct {
	Name string
	Args []string
	Raw  string
}

// Parser converts raw request lines into command values.
type Parser interface {
	ParseLine(line string) (Command, error)
}

// LineParser parses one whitespace-delimited command line at a time.
type LineParser struct{}

// NewLineParser constructs the default text protocol parser.
func NewLineParser() *LineParser {
	return &LineParser{}
}

// ParseLine converts a raw request line into a command value.
func (p *LineParser) ParseLine(line string) (Command, error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return Command{}, fmt.Errorf("empty command")
	}

	fields := strings.Fields(trimmed)
	return Command{
		Name: strings.ToUpper(fields[0]),
		Args: fields[1:],
		Raw:  trimmed,
	}, nil
}
