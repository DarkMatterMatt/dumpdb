package cmd

/**
 * Author: Matt Moran
 */

import (
	"github.com/spf13/cobra"
)

// the `search` command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search multiple databases simultaneously.",
	Long:  "",
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// searchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	v.BindPFlags(searchCmd.Flags())
}
