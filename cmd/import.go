package cmd

/**
 * Author: Matt Moran
 */

import (
	"bufio"
	"database/sql"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/darkmattermatt/dumpdb/internal/config"
	"github.com/darkmattermatt/dumpdb/internal/sourceid"

	"github.com/darkmattermatt/dumpdb/internal/linescanner"
	"github.com/darkmattermatt/dumpdb/internal/parseline"
	"github.com/darkmattermatt/dumpdb/pkg/pathexists"
	"github.com/darkmattermatt/dumpdb/pkg/reverse"
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
	importCmd.Flags().String("engine", "aria", "the database engine. Aria is recommended (requires MariaDB), MyISAM is supported for MySQL")
	importCmd.Flags().Bool("compress", false, "pack the database into a compressed, read-only format. Requires the Aria or MyISAM database engine")

	importCmd.Flags().Int("batchSize", 4e6, "number of lines per temporary file (used for the LOAD FILE INTO command). 1e6 = ~64MB, 16e6 = ~1GB")
	importCmd.Flags().StringP("filePrefix", "o", "[database]_", "temporary processed file prefix")

	importCmd.MarkFlagRequired("parser")
	importCmd.MarkFlagRequired("conn")
	importCmd.MarkFlagRequired("database")
	importCmd.MarkFlagRequired("sourcesDatabase")
}

func loadImportConfig(cmd *cobra.Command) {
	c.LineParser = v.GetString("parser")
	c.Database = v.GetString("database")
	c.SourcesDatabase = v.GetString("sourcesDatabase")
	c.Compress = v.GetBool("compress")

	c.BatchSize = v.GetInt("batchSize")
	c.FilePrefix = strings.ReplaceAll(v.GetString("filePrefix"), "[database]", c.Database)

	c.Engine = strings.ToLower(v.GetString("engine"))
	validEngines := []string{"aria", "myisam"}
	if !stringinslice.StringInSlice(c.Engine, validEngines) {
		showUsage(cmd, "Error: unknown database engine: "+c.Engine+". Valid options are: "+strings.Join(validEngines, ", ")+"\n")
	}

	c.Conn = v.GetString("conn")
	if !config.ValidDSNConn(c.Conn) {
		showUsage(cmd, "Invalid MySQL connection string "+c.Conn+". It must look like user:pass@tcp(127.0.0.1:3306)")
	}
	c.Conn += "/"
}

func runImport(cmd *cobra.Command, filesOrFolders []string) {
	loadImportConfig(cmd)

	importDone := make(chan bool, 1)
	importDone <- true

	var err error
	errFile, err = os.OpenFile(c.FilePrefix+"err.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	l.FatalOnErr(err)
	doneFile, err = os.OpenFile(c.FilePrefix+"done.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	l.FatalOnErr(err)
	skipFile, err = os.OpenFile(c.FilePrefix+"skip.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	l.FatalOnErr(err)
	outputFile, err = splitfilewriter.Create(c.FilePrefix+"tmp", ".csv", c.BatchSize)
	l.FatalOnErr(err)
	outputFile.FullFileCallback = func(s *splitfilewriter.SplitFileWriter) error {
		waitForImport(importDone)
		go importToDatabase(s.CurrentFileName(), importDone)
		return nil
	}

	db, err = sql.Open("mysql", c.Conn+c.Database)
	l.FatalOnErr(err)
	sourcesDb, err = sql.Open("mysql", c.Conn+c.SourcesDatabase)
	l.FatalOnErr(err)

	dataDir := getDataDir()
	disableDatabaseIndexes(dataDir)

	for _, path := range filesOrFolders {
		err := linescanner.LineScanner(path, processTextFileScanner)
		l.FatalOnErr(err)
	}

	// final import to mysql
	waitForImport(importDone)
	importToDatabase(outputFile.CurrentFileName(), importDone)

	// TODO: customisable tmpDir
	tmpDir := os.TempDir()
	if c.Compress {
		compressDatabase(dataDir, tmpDir)
	}
	restoreDatabaseIndexes(dataDir, tmpDir)
}

func importTextFileScanner(path string, lineScanner *bufio.Scanner) error {
	if !strings.HasSuffix(path, ".txt") && !strings.HasSuffix(path, ".csv") {
		l.V("Skipping: " + path)
		_, err := skipFile.WriteString(path + "\n")
		l.FatalOnErr(err)
		return nil
	}

	l.V("Processing: " + path)

	for lineScanner.Scan() {
		// CTRL+C means stop
		if signalInterrupt {
			return errors.New("Signal Interrupt")
		}

		line := lineScanner.Text()
		// skip blank lines
		if line == "" {
			continue
		}

		// parse & reformat line
		r, err := parseline.ParseLine(c.LineParser, line, path)
		if err != nil {
			if err == parseline.ErrInvalidLineParser {
				return errors.New(err.Error() + ": " + c.LineParser)
			}
			errFile.WriteString(line + "\n")
			continue
		}

		if r.EmailRev == "" && r.Email != "" {
			r.EmailRev = reverse.Reverse(r.Email)
		}
		r.SourceID, err = sourceid.SourceID(r.Source, sourcesDb, sourcesTable)
		l.FatalOnErr((err))

		arr := []string{strconv.FormatInt(r.SourceID, 10), r.Username, r.EmailRev, r.Hash, r.Password, r.Extra}

		// write string to output file
		_, err = outputFile.WriteString(strings.Join(arr, "\t") + "\n")
		l.FatalOnErr(err)
	}
	doneFile.WriteString(path + "\n")
	return nil
}
