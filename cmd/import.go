/**
 * Author: Matt Moran
 */
package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// the `import` command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a file or folder into a database.",
	Long:  "",
	Run:   runImport,
	Args:  validateArgs,
}

func validateArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("requires a file or directory to import recursively")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().StringP("connection", "C", "", "SQL connection string to enable automatic loading into MariaDB. Like user:pass@tcp(127.0.0.1:3306)/collection1")
	importCmd.Flags().Int("outputFileLines", 4e6, "(Temp) output file suffix")
	importCmd.Flags().String("doneLog", "[dbName]_done.log", "Output log file")
	importCmd.Flags().String("skipLog", "[dbName]_skip.log", "Skipped log file")
	importCmd.Flags().String("errLog", "[dbName]_err.log", "Error log file")
	importCmd.Flags().String("outFilePrefix", "[dbName]_", "(Temp) output file prefix")
	importCmd.Flags().String("outFileSuffix", ".txt", "(Temp) output file suffix")
	importCmd.Flags().String("tableName", "main", "Database table name to insert into")
	importCmd.Flags().String("sourcesConnection", "", "Optional SQL connection string to enable automatic loading into MariaDB. Like user:pass@tcp(127.0.0.1:3306)/sources")
	importCmd.Flags().String("sourcesTableName", "sources", "Database table name to store sources in")

	importCmd.MarkFlagRequired("connection")
	v.BindPFlags(importCmd.Flags())
}

func runImport(cmd *cobra.Command, args []string) {
	fmt.Println("import called")
}
