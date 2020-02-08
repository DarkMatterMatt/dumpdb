package cmd

/**
 * Author: Matt Moran
 */

import (
	"errors"
	"time"

	"github.com/darkmattermatt/dumpdb/pkg/pathexists"
	l "github.com/darkmattermatt/dumpdb/pkg/simplelog"
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
			return errors.New("Missing at least one parameter: file or directory to import recursively")
		}
		return pathexists.PathsAllExist(args)
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

	processCmd.Flags().String("filesPrefix", time.Now().Format("2006-01-02_15-04-05_"), "temporary processed file prefix")

	v.BindPFlags(processCmd.Flags())
}

func runProcess(cmd *cobra.Command, args []string) {
	l.I("process called")
}
