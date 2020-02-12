package config

// Config contains the configuration options for DumpDB
type Config struct {
	// root
	Verbosity  int
	ConfigFile string

	// search
	Databases    []string
	ConnPrefix   string
	DbTable      string
	SourcesConn  string
	SourcesTable string
	Query        string
	Columns      []string

	// process
	OutFileLines  int
	OutFilePrefix string
	OutFileSuffix string
	ErrLog        string
	DoneLog       string
	skipLog       string

	// import
	Conn     string
	Table    string
	Engine   string
	Compress bool
}
