package config

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
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

// ShowUsage exits after printing an error message followed by the command's usage
func ShowUsage(cmd *cobra.Command, s string) {
	fmt.Println(s)
	cmd.Usage()
	os.Exit(1)
}

// DsnPattern from https://github.com/go-sql-driver/mysql/blob/f4bf8e8e0aa93d4ead0c6473503ca2f5d5eb65a8/utils.go#L34
var DsnPattern = regexp.MustCompile(
	`^(?:(?P<user>.*?)(?::(?P<passwd>.*))?@)?` + // [user[:password]@]
		`(?:(?P<net>[^\(]*)(?:\((?P<addr>[^\)]*)\))?)?` + // [net[(addr)]]
		`\/(?P<dbname>.*?)` + // /dbname
		`(?:\?(?P<params>[^\?]*))?$`) // [?param1=value1&paramN=valueN]

// ValidDSNConn checks that a string is in the format `user:pass@tcp(127.0.0.1:3306)`
func ValidDSNConn(s string) bool {
	return DsnPattern.MatchString(s)
}
