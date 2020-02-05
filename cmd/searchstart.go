package cmd

/**
 * Author: Matt Moran
 */

import (
	"fmt"

	"github.com/spf13/cobra"
)

// the `search` command
var searchStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Search multiple databases simultaneously after starting a MySQL server.",
	Long:  "",
	Run:   runSearchStart,
}

func init() {
	searchCmd.AddCommand(searchStartCmd)

	// searchStartCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	v.BindPFlags(searchStartCmd.Flags())
}

func runSearchStart(cmd *cobra.Command, args []string) {
	fmt.Println("search called")
}
