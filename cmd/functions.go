package cmd

/**
 * Author: Matt Moran
 */

import (
	l "github.com/darkmattermatt/dumpdb/pkg/simplelog"
	"github.com/spf13/cobra"
)

// showUsage exits after printing an error message followed by the command's usage
func showUsage(cmd *cobra.Command, s string) {
	l.R(s)
	cmd.Usage()
	l.F(s)
}
