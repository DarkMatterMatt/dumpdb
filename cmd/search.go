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
	Short: "Search multiple dump databases simultaneously.",
	Long:  "",
	Run:   runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Positional args: filesOrFolders: files and/or folders to import
	searchCmd.Flags().StringP("db", "d", "", "comma separated list of databases to search")
	searchCmd.Flags().StringP("connPrefix", "c", "", "connection string prefix to connect to MySQL databases. Like user:pass@tcp(127.0.0.1:3306)")
	searchCmd.Flags().StringP("dbTable", "t", "main", "database table name to search. Must be the same for all databases")
	searchCmd.Flags().StringP("sourcesConn", "C", "", "connection string for the sources database. Like user:pass@tcp(127.0.0.1:3306)/sources")
	searchCmd.Flags().StringP("sourcesTable", "T", "sources", "SQL connection string for the sources database. Like user:pass@tcp(127.0.0.1:3306)/sources")

	searchCmd.Flags().StringP("query", "Q", "", "the WHERE clause of a SQL query. Yes it's injected, so try not to break your own database")
	searchCmd.Flags().String("columns", "all", "comma separated list of columns to retrieve from the database")

	searchCmd.MarkFlagRequired("db")
	searchCmd.MarkFlagRequired("connPrefix")
	searchCmd.MarkFlagRequired("query")

	v.BindPFlags(searchCmd.Flags())
}

func runSearch(cmd *cobra.Command, args []string) {
	fmt.Println("search called")
}
