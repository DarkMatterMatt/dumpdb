package config

import (
	"regexp"
)

// Config contains the configuration options for DumpDB
type Config struct {
	// root
	Verbosity  int
	ConfigFile string

	// init
	Sources string

	// search
	Databases   []string
	SourcesConn string
	Query       string
	Columns     []string

	// process
	BatchSize  int
	FilePrefix string

	// import
	Conn     string
	Engine   string
	Compress bool
}

// DsnPattern matches a string beginning with `user:pass@tcp(127.0.0.1:3306)`
var DsnPattern = regexp.MustCompile(`^\w+:\w*@tcp\([\w\.]+:\d+\)`)

// ValidDSNConn checks that a string is in the format `user:pass@tcp(127.0.0.1:3306)`
func ValidDSNConn(s string) bool {
	return DsnPattern.MatchString(s)
}
