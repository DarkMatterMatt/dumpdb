package parseline

import "errors"

// ErrInvalidLineParser occurs when a line parser that does not exists is requested
var ErrInvalidLineParser = errors.New("The requested line parser does not exist")

var lineParsers = make(map[string]func(line, source string) (Record, error))

// ParseLine parses a single line with the requested line parser
func ParseLine(name, line, source string) (Record, error) {
	parser, ok := lineParsers[name]
	if ok {
		return parser(line, source)
	}
	return Record{}, ErrInvalidLineParser
}
