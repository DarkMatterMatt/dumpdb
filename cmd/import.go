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
	importCmd.Flags().StringP("conn", "c", "", "connection string for the SQL database. Like user:pass@tcp(127.0.0.1:3306)")
	importCmd.Flags().StringP("database", "d", "", "database name to import into")
	importCmd.Flags().StringP("sourcesDatabase", "s", "", "database name to store sources in")
	importCmd.Flags().StringP("engine", "e", "aria", "the database engine. Aria is recommended (requires MariaDB), MyISAM is supported for MySQL")
	importCmd.Flags().Bool("compress", false, "pack the database into a compressed, read-only format. Requires the Aria or MyISAM database engine")

	importCmd.Flags().Int("batchSize", 4e6, "number of lines per temporary file (used for the LOAD FILE INTO command). 1e6 = ~32MB, 32e6 = ~1GB")
	importCmd.Flags().StringP("filePrefix", "o", "[database]_", "temporary processed file prefix")

	importCmd.MarkFlagRequired("conn")
	importCmd.MarkFlagRequired("database")
	importCmd.MarkFlagRequired("sourcesDatabase")
}

func loadImportConfig(cmd *cobra.Command) {
	c.Database = v.GetString("database")
	c.SourcesDatabase = v.GetString("sourcesDatabase")
	c.Engine = v.GetString("engine")
	c.Compress = v.GetBool("compress")

	c.BatchSize = v.GetInt("batchSize")
	c.FilePrefix = strings.ReplaceAll(v.GetString("filePrefix"), "[database]", c.Database)

	c.Conn = v.GetString("conn")
	if !config.ValidDSNConn(c.Conn) {
		showUsage(cmd, "Invalid MySQL connection string "+c.Conn+". It must look like user:pass@tcp(127.0.0.1:3306)")
	}
	c.Conn += "/"
}

func runImport(cmd *cobra.Command, filesOrFolders []string) {
	loadImportConfig(cmd)

	var err error
	errFile, err = os.OpenFile(c.FilePrefix+"_err.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0)
	l.FatalOnErr(err)
	doneFile, err = os.OpenFile(c.FilePrefix+"_done.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0)
	l.FatalOnErr(err)
	skipFile, err = os.OpenFile(c.FilePrefix+"_skip.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0)
	l.FatalOnErr(err)
	outputFile, err = splitfilewriter.Create(c.FilePrefix+"_tmp_", ".csv", c.BatchSize)
	l.FatalOnErr(err)
	outputFile.NewFileCallback = func(*splitfilewriter.SplitFileWriter) error {
		// import to mysql
		return nil
	}

	db, err = sql.Open("mysql", c.Conn+c.Database)
	l.FatalOnErr(err)
	sourcesDb, err = sql.Open("mysql", c.Conn+c.SourcesDatabase)
	l.FatalOnErr(err)

	for _, path := range filesOrFolders {
		linescanner.LineScanner(path, importTextFileScanner)
	}

	// final import to mysql
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
		r, err := parseline.ParseLine(line, path)
		if err != nil {
			errFile.WriteString(line + "\n")
			continue
		}

		if r.EmailRev == "" && r.Email != "" {
			r.EmailRev = reverse.Reverse(r.Email)
		}
		r.SourceID, err = sourceid.SourceID(r.Source, sourcesDb, sourcesTable)
		l.FatalOnErr((err))

		arr := []string{strconv.FormatInt(r.SourceID, 10), r.Username, r.EmailRev, r.Hash, r.Password}

		// write string to output file
		_, err = outputFile.WriteString(strings.Join(arr, "\t") + "\n")
		l.FatalOnErr(err)
	}
	doneFile.WriteString(path + "\n")
	return nil
}
