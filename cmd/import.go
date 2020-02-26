package cmd

/**
 * Author: Matt Moran
 */

import (
	"bufio"
	"database/sql"
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/darkmattermatt/dumpdb/internal/config"
	"github.com/darkmattermatt/dumpdb/internal/linescanner"

	"github.com/darkmattermatt/dumpdb/internal/parseline"
	"github.com/darkmattermatt/dumpdb/pkg/pathexists"
	l "github.com/darkmattermatt/dumpdb/pkg/simplelog"
	"github.com/darkmattermatt/dumpdb/pkg/splitfilewriter"
	"github.com/darkmattermatt/dumpdb/pkg/stringinslice"
	"github.com/spf13/cobra"
)

// the `import` command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import files or folders into a database.",
	Long:  "",
	Run:   runImport,
	PreRun: func(cmd *cobra.Command, args []string) {
		v.BindPFlags(cmd.Flags())
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Missing files to import")
		}
		return pathexists.AssertPathsAreFiles(args)
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	// Positional args: filesOrFolders: files and/or folders to import
	importCmd.Flags().StringP("parser", "p", "", "the custom line parser to use. Modify the internal/parseline package to add another line parser")
	importCmd.Flags().StringP("conn", "c", "", "connection string for the SQL database. Like user:pass@tcp(127.0.0.1:3306)")
	importCmd.Flags().StringP("database", "d", "", "database name to import into")
	importCmd.Flags().StringP("sourcesDatabase", "s", "", "database name to store sources in")
	importCmd.Flags().Bool("compress", false, "pack the database into a compressed, read-only format. Requires the Aria or MyISAM database engine")

	importCmd.Flags().Int("batchSize", 4e6, "number of lines per temporary file (used for the LOAD FILE INTO command). 1e6 = ~64MB, 16e6 = ~1GB")
	importCmd.Flags().StringP("filePrefix", "o", "[database]_", "temporary processed file prefix")

	importCmd.MarkFlagRequired("parser")
	importCmd.MarkFlagRequired("conn")
	importCmd.MarkFlagRequired("database")
	importCmd.MarkFlagRequired("sourcesDatabase")
}

func loadImportConfig(cmd *cobra.Command) {
	c.Database = v.GetString("database")
	c.SourcesDatabase = v.GetString("sourcesDatabase")
	c.Compress = v.GetBool("compress")

	c.BatchSize = v.GetInt("batchSize")
	c.FilePrefix = strings.ReplaceAll(v.GetString("filePrefix"), "[database]", c.Database)

	c.LineParser = v.GetString("parser")
	if !parseline.ParserExists(c.LineParser) {
		showUsage(cmd, "Error: unknown line parser: "+c.LineParser+". Have you made a new parser for your dump in the internal/parseline package?")
	}

	c.Conn = v.GetString("conn")
	if !config.ValidDSNConn(c.Conn) {
		showUsage(cmd, "Invalid MySQL connection string "+c.Conn+". It must look like user:pass@tcp(127.0.0.1:3306)")
	}
	c.Conn += "/"
}

func checkDatabaseToolsExist() {
	if c.Engine == "aria" {
		_, err := exec.LookPath("aria_chk")
		l.FatalOnErr("Checking that the required database tools are in PATH", err)

		if c.Compress {
			_, err = exec.LookPath("aria_pack")
			l.FatalOnErr("Checking that the required database tools are in PATH", err)
		}
	} else if c.Engine == "myisam" {
		_, err := exec.LookPath("myisamchk")
		l.FatalOnErr("Checking that the required database tools are in PATH", err)

		if c.Compress {
			_, err = exec.LookPath("myisampack")
			l.FatalOnErr("Checking that the required database tools are in PATH", err)
		}
	}
}

func checkDatabaseFilePermissions(dataDir string) {
	fname := dataDir + c.Database + "/" + mainTable
	if c.Engine == "aria" {
		f, err := os.OpenFile(fname+".MAD", os.O_RDWR, 0)
		l.FatalOnErr("Checking read/write permissions", err)
		f.Close()
		f, err = os.OpenFile(fname+".MAI", os.O_RDWR, 0)
		l.FatalOnErr("Checking read/write permissions", err)
		f.Close()
	} else if c.Engine == "myisam" {
		f, err := os.OpenFile(fname+".MYD", os.O_RDWR, 0)
		l.FatalOnErr("Checking read/write permissions", err)
		f.Close()
		f, err = os.OpenFile(fname+".MYI", os.O_RDWR, 0)
		l.FatalOnErr("Checking read/write permissions", err)
		f.Close()
	}
}

func runImport(cmd *cobra.Command, filesOrFolders []string) {
	loadImportConfig(cmd)

	importDone := make(chan bool, 1)
	importDone <- true

	var err error
	errFile, err = os.OpenFile(c.FilePrefix+"err.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	l.FatalOnErr("Opening error log", err)
	doneFile, err = os.OpenFile(c.FilePrefix+"done.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	l.FatalOnErr("Opening done log", err)
	skipFile, err = os.OpenFile(c.FilePrefix+"skip.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	l.FatalOnErr("Opening skip log", err)
	outputFile, err = splitfilewriter.Create(c.FilePrefix+"tmp", ".csv", c.BatchSize)
	l.FatalOnErr("Opening first output file", err)
	outputFile.FullFileCallback = func(s *splitfilewriter.SplitFileWriter) error {
		waitForImport(importDone)
		go importToDatabase(s.CurrentFileName(), importDone)
		return nil
	}

	db, err = sql.Open("mysql", c.Conn+c.Database)
	l.FatalOnErr("Opening main database connection", err)
	sourcesDb, err = sql.Open("mysql", c.Conn+c.SourcesDatabase)
	l.FatalOnErr("Opening sources database connection", err)

	c.Engine = queryDatabaseEngine()
	validEngines := []string{"aria", "myisam"}
	if !stringinslice.StringInSlice(c.Engine, validEngines) {
		showUsage(cmd, "Error: unknown database engine: "+c.Engine+". Use `dumpdb init` to create a new database with a valid database engine.")
	}

	dataDir := getDataDir()
	checkDatabaseToolsExist()
	checkDatabaseFilePermissions(dataDir)

	disableDatabaseIndexes(dataDir)

	for _, path := range filesOrFolders {
		err := linescanner.LineScanner(path, func(a string, b *bufio.Scanner) error {
			return processTextFileScanner(a, b, true)
		})
		if err == errSignalInterrupt {
			return
		}
		l.FatalOnErr("Importing "+path, err)
	}

	// final import to mysql
	waitForImport(importDone)
	importToDatabase(outputFile.CurrentFileName(), importDone)

	flushAndLockTables()

	// TODO: customisable tmpDir
	tmpDir := os.TempDir()
	if c.Compress {
		compressDatabase(dataDir, tmpDir)
	}
	restoreDatabaseIndexes(dataDir, tmpDir)

	unlockTables()
	l.I("Please restart the MySQL server to allow using databases indexes")
}
