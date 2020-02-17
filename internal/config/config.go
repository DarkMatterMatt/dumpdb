package config

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
