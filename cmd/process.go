package cmd

/**
 * Author: Matt Moran
 */

import (
	"errors"
	"fmt"

	"github.com/darkmattermatt/dumpdb/pkg/pathexists"
	"github.com/spf13/cobra"
)

// the `process` command
var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Process a file or folder into tab-delimited files.",
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

	importCmd.Flags().Int("outputFileLines", 4e6, "Output file suffix")
	importCmd.Flags().String("doneLog", "[outFilePrefix]done.log", "Output log file")
	importCmd.Flags().String("skipLog", "[outFilePrefix]skip.log", "Skipped log file")
	importCmd.Flags().String("errLog", "[outFilePrefix]err.log", "Error log file")
	importCmd.Flags().String("outFilePrefix", "", "Output file prefix")
	importCmd.Flags().String("outFileSuffix", ".txt", "Output file suffix")

	v.BindPFlags(importCmd.Flags())
}

func runProcess(cmd *cobra.Command, args []string) {
	fmt.Println("process called")
}
