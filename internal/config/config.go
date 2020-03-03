package config

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/darkmattermatt/dumpdb/internal/parseline"
	"github.com/darkmattermatt/dumpdb/pkg/pathexists"
	"github.com/darkmattermatt/dumpdb/pkg/simplelog"
	"github.com/darkmattermatt/dumpdb/pkg/stringinslice"
)

// Config contains the configuration options for DumpDB
type Config struct {
	// root
	Verbosity int

	// init
	Databases       []string
	Conn            string
	SourcesDatabase string
	Engine          string
	Indexes         []string

	// search
	Query        string
	OutputFormat string
	Columns      []string

	// import
	FilesOrFolders []string
	LineParser     string
	Database       string
	Compress       bool
	BatchSize      int
	FilePrefix     string
}

// SetVerbosity sets the Config verbosity
func (c *Config) SetVerbosity(v int) error {
	if v < simplelog.FATAL {
		return fmt.Errorf("Invalid verbosity: is %d, must be greater than or equal to %d", v, simplelog.FATAL)
	} else if v > simplelog.DEBUG {
		return fmt.Errorf("Invalid verbosity: is %d, must be less than or equal to %d", v, simplelog.DEBUG)
	}
	c.Verbosity = v
	return nil
}

// SetDatabases sets the databases to use, SetConn must be called first
func (c *Config) SetDatabases(dbs []string) error {
	if c.Conn == "" {
		return errors.New("Programming error: SetConn must be called before SetDatabases")
	}

	conn, err := sql.Open("mysql", c.Conn)
	if err != nil {
		return err
	}

	for _, db := range dbs {
		_, err = conn.Exec("USE " + db)
		if err != nil {
			return err
		}

		var dbType string
		err = conn.QueryRow("SELECT v FROM metadata WHERE k='type'").Scan(&dbType)
		if err != nil {
			return err
		}

		if dbType != "main" {
			return errors.New("The specified database is not a DumpDB 'main' database type")
		}
	}

	c.Databases = dbs
	return nil
}

// connPattern matches a string in the format `user:pass@tcp(127.0.0.1:3306)`
var connPattern = regexp.MustCompile(`^\w+:\w*@tcp\([\w\.]+:\d+\)$`)

// SetConn sets the database connection string, in the format `user:pass@tcp(127.0.0.1:3306)`
func (c *Config) SetConn(s string) error {
	if !connPattern.MatchString(s) {
		return errors.New("Invalid MySQL connection string: must be in the format `user:pass@tcp(127.0.0.1:3306)`")
	}
	c.Conn = s + "/"
	return nil
}

// SetSourcesDatabase sets the sources database name, SetConn must be called first
func (c *Config) SetSourcesDatabase(s string) error {
	if c.Conn == "" {
		return errors.New("Programming error: SetConn must be called before SetDatabases")
	}

	conn, err := sql.Open("mysql", c.Conn+s)
	if err != nil {
		return err
	}

	var dbType string
	err = conn.QueryRow("SELECT v FROM metadata WHERE k='type'").Scan(&dbType)
	if err != nil {
		return err
	}

	if dbType != "sources" {
		return errors.New("The specified database is not a DumpDB 'sources' database type")
	}

	c.SourcesDatabase = s
	return nil
}

// SetEngine sets the database storage engine
func (c *Config) SetEngine(e string) error {
	supportedEngines := []string{"aria", "myisam"}
	if !stringinslice.StringInSlice(strings.ToLower(e), supportedEngines) {
		return errors.New("Error: unknown database engine: " + e + ". Supported engines are: " + strings.Join(supportedEngines, ", "))
	}
	c.Engine = e
	return nil
}

// SetIndexes sets the 'main' table indexes
func (c *Config) SetIndexes(indexes []string) error {
	columns := []string{"email_rev", "hash", "password", "sourceid", "username", "extra"}
	for _, index := range indexes {
		index = strings.ToLower(index)

		if index == "email" {
			return errors.New("Sorry, the `email` column cannot be indexed due to the way the email is stored (it is stored reversed so that emails can be searched from the end to the start)")
		}

		if !stringinslice.StringInSlice(index, columns) {
			return errors.New("Cannot index by column '" + index + "'. Valid columns are: " + strings.Join(columns, ", "))
		}
	}
	c.Indexes = indexes
	return nil
}

// SetQuery sets the WHERE ... clause of the SQL query
func (c *Config) SetQuery(q string) error {
	if strings.Contains(q, ";") {
		return errors.New("Query string must not contain a semicolon")
	}
	if strings.HasPrefix(strings.ToUpper(q), "WHERE ") {
		q = q[6:]
	}
	c.Query = q
	return nil
}

// SetOutputFormat sets the search output format
func (c *Config) SetOutputFormat(o string) error {
	supportedFormats := []string{"text", "jsonl"}
	if !stringinslice.StringInSlice(strings.ToLower(o), supportedFormats) {
		return errors.New("Unsupported output format: '" + o + "'. Supported formats are: " + strings.Join(supportedFormats, ", "))
	}
	c.OutputFormat = o
	return nil
}

// SetColumns sets the columns to output when searching
func (c *Config) SetColumns(cols []string) error {
	if len(cols) < 1 {
		c.Columns = []string{"email", "hash", "password", "sourceid", "username", "extra"}
		if c.SourcesDatabase != "" {
			c.Columns = append(c.Columns, "source")
		}
		return nil
	}

	supportedCols := []string{"email", "email_rev", "hash", "password", "sourceid", "username", "extra", "source"}

	if c.SourcesDatabase == "" && stringinslice.StringInSlice("source", cols) {
		return errors.New("The sources database must be set to load the `source` column")
	}

	for _, col := range cols {
		if !stringinslice.StringInSlice(strings.ToLower(col), supportedCols) {
			return errors.New("Cannot index by column '" + col + "'. Valid columns are: " + strings.Join(supportedCols, ", "))
		}
	}
	c.Columns = cols
	return nil
}

// SetFilesOrFolders sets the files or folders to import
func (c *Config) SetFilesOrFolders(paths []string) error {
	if len(paths) < 1 {
		return errors.New("Missing files or folders to import")
	}
	err := pathexists.AssertPathsAllExist(paths)
	if err != nil {
		return err
	}
	c.FilesOrFolders = paths
	return nil
}

// SetLineParser sets the function to parse lines when importing
func (c *Config) SetLineParser(p string) error {
	if !parseline.ParserExists(p) {
		return errors.New("Error: unknown line parser: " + p + ". Have you made a new parser for your dump in the internal/parseline package?")
	}
	c.LineParser = p
	return nil
}

// SetDatabase sets the 'main' database name, SetConn must be called first
func (c *Config) SetDatabase(s string) error {
	if c.Conn == "" {
		return errors.New("Programming error: SetConn must be called before SetDatabases")
	}

	conn, err := sql.Open("mysql", c.Conn+s)
	if err != nil {
		return err
	}

	var dbType string
	err = conn.QueryRow("SELECT v FROM metadata WHERE k='type'").Scan(&dbType)
	if err != nil {
		return err
	}

	if dbType != "main" {
		return errors.New("The specified database is not a DumpDB 'main' database type")
	}

	c.Database = s
	return nil
}

// SetCompress sets whether or not to compress the database after importing
func (c *Config) SetCompress(compress bool) error {
	c.Compress = compress
	return nil
}

// SetBatchSize sets the number of records that are imported to the database together
func (c *Config) SetBatchSize(size int) error {
	c.BatchSize = size
	return nil
}

// SetFilePrefix sets the number of records that are imported to the database together
func (c *Config) SetFilePrefix(prefix string) error {
	prefix = strings.ReplaceAll(prefix, "[database]", c.Database)

	f, err := os.OpenFile(prefix+"__dumpdb__test_write", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	f.Close()

	err = os.Remove(prefix + "__dumpdb__test_write")
	if err != nil {
		return err
	}

	c.FilePrefix = prefix
	return nil
}
