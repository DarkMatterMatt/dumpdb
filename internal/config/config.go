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

	Conn          string
	Table         string
	Engine        string
	Compress      bool
	TmpFileLines  int
	TmpFilePrefix string
	TmpFileSuffix string
	ErrLog        string
	DoneLog       string
	skipLog       string
}
