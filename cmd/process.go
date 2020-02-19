package cmd

/**
 * Author: Matt Moran
 */

import (
	"bufio"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/darkmattermatt/dumpdb/internal/linescanner"
	"github.com/darkmattermatt/dumpdb/internal/parseline"
	"github.com/darkmattermatt/dumpdb/pkg/pathexists"
	l "github.com/darkmattermatt/dumpdb/pkg/simplelog"
	"github.com/darkmattermatt/dumpdb/pkg/splitfilewriter"
	"github.com/spf13/cobra"
)

// the `process` command
var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Process files or folders into a regularised tab-delimited text file.",
	Long:  "",
	Run:   runProcess,
	PreRun: func(cmd *cobra.Command, args []string) {
		v.BindPFlags(cmd.Flags())
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Missing files to process")
		}
		return pathexists.AssertPathsAreFiles(args)
	},
}

func init() {
	rootCmd.AddCommand(processCmd)

	// Positional args: filesOrFolders: files and/or folders to import
	processCmd.Flags().StringP("parser", "p", "", "the custom line parser to use. Modify the internal/parseline package to add another line parser")
	processCmd.Flags().Int("batchSize", 4e6, "number of lines per temporary file (used for the LOAD FILE INTO command). 1e6 = ~64MB, 16e6 = ~1GB")
	processCmd.Flags().String("filePrefix", time.Now().Format("2006-01-02_1504_05 "), "processed file prefix")

	importCmd.MarkFlagRequired("parser")
}

func loadProcessConfig(cmd *cobra.Command) {
	c.BatchSize = v.GetInt("batchSize")
	c.FilePrefix = v.GetString("filePrefix")

	c.LineParser = v.GetString("parser")
	if !parseline.ParserExists(c.LineParser) {
		showUsage(cmd, "Error: unknown line parser: "+c.LineParser+". Have you made a new parser for your dump in the internal/parseline package?")
	}
}

func runProcess(cmd *cobra.Command, filesOrFolders []string) {
	loadProcessConfig(cmd)

	var err error
	errFile, err = os.OpenFile(c.FilePrefix+"err.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	l.FatalOnErr(err)
	doneFile, err = os.OpenFile(c.FilePrefix+"done.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	l.FatalOnErr(err)
	skipFile, err = os.OpenFile(c.FilePrefix+"skip.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	l.FatalOnErr(err)
	outputFile, err = splitfilewriter.Create(c.FilePrefix+"output", ".csv", c.BatchSize)
	l.FatalOnErr(err)
	outputFile.FullFileCallback = func(s *splitfilewriter.SplitFileWriter) error {
		l.D("Beginning to write to " + s.NextFileName())
		return nil
	}

	for _, path := range filesOrFolders {
		err := linescanner.LineScanner(path, processTextFileScanner)
		if err == errSignalInterrupt {
			return
		}
		l.FatalOnErr(err)
	}
}

func processTextFileScanner(path string, lineScanner *bufio.Scanner) error {
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
			return errSignalInterrupt
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

		parsedArray := []string{r.Source, r.Username, r.Email, r.Hash, r.Password, r.Extra}
		parsedStr := strings.Join(parsedArray, "\t")

		// write string to output file
		_, err = outputFile.WriteString(parsedStr + "\n")
		if err != nil {
			l.FatalOnErr(err)
		}
	}
	doneFile.WriteString(path + "\n")
	return nil
}
