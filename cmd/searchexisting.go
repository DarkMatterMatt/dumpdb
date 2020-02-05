package cmd

/**
 * Author: Matt Moran
 */

import (
	"fmt"

	"github.com/spf13/cobra"
)

// the `search` command
var searchExistingCmd = &cobra.Command{
	Use:   "existing",
	Short: "Search multiple databases simultaneously using an existing MySQL server.",
	Long:  "",
	Run:   runSearchExisting,
}

func init() {
	searchCmd.AddCommand(searchExistingCmd)

	// searchExistingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	v.BindPFlags(searchExistingCmd.Flags())
}

func runSearchExisting(cmd *cobra.Command, args []string) {
	fmt.Println("search called")
}
