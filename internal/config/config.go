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
	Databases       []string
	Conn            string
	SourcesDatabase string
	Engine          string

	// search
	Query   string
	Columns []string

	// import
	FilesOrFolders string
	LineParser     string
	Database       string
	Compress       bool
	BatchSize      int
	FilePrefix         string
}

// DsnPattern matches a string in the format `user:pass@tcp(127.0.0.1:3306)`
var DsnPattern = regexp.MustCompile(`^\w+:\w*@tcp\([\w\.]+:\d+\)$`)

// ValidDSNConn checks that a string is in the format `user:pass@tcp(127.0.0.1:3306)`
func ValidDSNConn(s string) bool {
	return DsnPattern.MatchString(s)
}
