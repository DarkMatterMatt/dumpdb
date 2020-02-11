package config

// Config contains the configuration options for DumpDB
type Config struct {
	Verbosity  int
	ConfigFile string

	Databases    []string
	ConnPrefix   string
	DbTable      string
	SourcesConn  string
	SourcesTable string
	Query        string
	Columns      []string
}
