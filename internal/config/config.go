package config

// Config contains the configuration options for DumpDB
type Config struct {
	// root
	Verbosity  int
	ConfigFile string

	// search
	Databases    []string
	ConnPrefix   string
	SourcesConn  string
	SourcesTable string
	Query        string
	Columns      []string

	// process
	OutFileLines  int
	OutFilePrefix string
	OutFileSuffix string
	FilesPrefix   string
	ErrLog        string
	DoneLog       string
	SkipLog       string

	// import
	Conn     string
	Table    string
	Engine   string
	Compress bool
}
