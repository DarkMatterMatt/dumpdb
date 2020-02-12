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
	processCmd.Flags().Int("outputFileLines", 4e6, "number of lines per temporary file (used for the LOAD FILE INTO command). 1e6 = ~32MB, 32e6 = ~1GB")
	processCmd.Flags().String("outputFilePrefix", "[filesPrefix]_", "temporary processed file prefix")
	processCmd.Flags().String("outputFileSuffix", ".txt", "temporary processed file suffix")

	processCmd.Flags().String("errLog", "[filesPrefix]_err.log", "log file for unparsed lines")
	processCmd.Flags().String("doneLog", "[filesPrefix]_done.log", "log file for processed input files")
	processCmd.Flags().String("skipLog", "[filesPrefix]_skip.log", "log file for skipped input files")

	processCmd.Flags().String("filesPrefix", time.Now().Format("2006-01-02_1504_05"), "processed file prefix")

	v.BindPFlags(processCmd.Flags())
}

func getProcessFilename(s string) string {
	return strings.ReplaceAll(s, "[filesPrefix]", c.FilesPrefix)
}

func loadProcessConfig() error {
	c.FilesPrefix = v.GetString("filesPrefix")
	c.OutFileLines = v.GetInt("outputFileLines")
	c.OutFilePrefix = getProcessFilename(v.GetString("outputFilePrefix"))
	c.OutFileSuffix = v.GetString("outputFileSuffix")
	c.ErrLog = getProcessFilename(v.GetString("errLog"))
	c.DoneLog = getProcessFilename(v.GetString("doneLog"))
	c.SkipLog = getProcessFilename(v.GetString("skipLog"))
	return nil
}

func runProcess(cmd *cobra.Command, filesOrFolders []string) {
	err := loadProcessConfig()
	l.FatalOnErr(err)

	errFile, err = os.OpenFile(c.ErrLog, os.O_CREATE|os.O_APPEND, 0)
	l.FatalOnErr(err)
	doneFile, err = os.OpenFile(c.DoneLog, os.O_CREATE|os.O_APPEND, 0)
	l.FatalOnErr(err)
	skipFile, err = os.OpenFile(c.SkipLog, os.O_CREATE|os.O_APPEND, 0)
	l.FatalOnErr(err)
	outputFile, err = splitfilewriter.OpenFileNewWriter(c.OutFilePrefix, c.OutFileSuffix, c.OutFileLines, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	l.FatalOnErr(err)

	for _, path := range filesOrFolders {
		linescanner.LineScanner(path, processTextFileScanner)
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

		parsedArray := []string{r.Source, r.Username, r.Email, r.Hash, r.Password}
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
