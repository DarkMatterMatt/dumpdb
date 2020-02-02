package cmd

/**
 * Author: Matt Moran
 */

import (
	"fmt"

	"github.com/spf13/cobra"
)

// the `search` command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search multiple databases simultaneously.",
	Long:  "",
	Run:   runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// searchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	v.BindPFlags(searchCmd.Flags())
}

func runSearch(cmd *cobra.Command, args []string) {
	fmt.Println("search called")
}
