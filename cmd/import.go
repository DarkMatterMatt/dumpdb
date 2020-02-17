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
	importCmd.Flags().StringP("conn", "c", "", "connection string for the SQL database. Like user:pass@tcp(127.0.0.1:3306)/collection1")
	importCmd.Flags().StringP("sourcesConn", "C", "", "connection string for the sources database. Like user:pass@tcp(127.0.0.1:3306)/sources")
	importCmd.Flags().String("engine", "Aria", "the database engine. Aria is recommended (requires MariaDB), MyISAM is supported for MySQL")
	importCmd.Flags().Bool("compress", false, "pack the database into a compressed, read-only format. Requires the Aria or MyISAM database engine")

	importCmd.Flags().Int("batchSize", 4e6, "number of lines per temporary file (used for the LOAD FILE INTO command). 1e6 = ~32MB, 32e6 = ~1GB")
	importCmd.Flags().String("filePrefix", "", "temporary processed file prefix")

	importCmd.MarkFlagRequired("conn")
	importCmd.MarkFlagRequired("sourcesConn")
	v.BindPFlags(importCmd.Flags())
}

func loadImportConfig() error {
	c.Conn = v.GetString("conn")
	c.SourcesConn = v.GetString("sourcesConn")
	c.Engine = v.GetString("engine")
	c.Compress = v.GetBool("compress")

	c.BatchSize = v.GetInt("batchSize")
	c.FilePrefix = v.GetString("filePrefix")

	return nil
}

func runImport(cmd *cobra.Command, filesOrFolders []string) {
	l.I("import called")
	err := loadImportConfig()
	l.FatalOnErr(err)

	errFile, err = os.OpenFile(c.FilePrefix+"_err.log", os.O_CREATE|os.O_APPEND, 0)
	l.FatalOnErr(err)
	doneFile, err = os.OpenFile(c.FilePrefix+"_done.log", os.O_CREATE|os.O_APPEND, 0)
	l.FatalOnErr(err)
	skipFile, err = os.OpenFile(c.FilePrefix+"_skip.log", os.O_CREATE|os.O_APPEND, 0)
	l.FatalOnErr(err)
	outputFile, err = splitfilewriter.Create(c.FilePrefix+"_tmp_", ".csv", c.BatchSize)
	l.FatalOnErr(err)
	outputFile.NewFileCallback = func(*splitfilewriter.SplitFileWriter) error {
		// import to mysql
		return nil
	}

	db, err = sql.Open("mysql", c.Conn)
	l.FatalOnErr(err)
	sourcesDb, err = sql.Open("mysql", c.SourcesConn)
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
