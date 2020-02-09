package cmd

/**
 * Author: Matt Moran
 */

import (
	"errors"

	"github.com/darkmattermatt/dumpdb/pkg/pathexists"
	l "github.com/darkmattermatt/dumpdb/pkg/simplelog"
	"github.com/spf13/cobra"
)

// the `import` command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import files or folders into a database.",
	Long:  "",
	Run:   runImport,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Missing files to import")
		}
		return pathexists.AssertPathsAreFiles(args)
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	// Positional args: filesOrFolders: files and/or folders to import
	importCmd.Flags().StringP("conn", "c", "", "connection string for the SQL database. Like user:pass@tcp(127.0.0.1:3306)/collection1")
	importCmd.Flags().StringP("table", "t", "main", "database table name to insert into")
	importCmd.Flags().StringP("sourcesConn", "C", "", "connection string for the sources database. Like user:pass@tcp(127.0.0.1:3306)/sources")
	importCmd.Flags().StringP("sourcesTable", "T", "sources", "database table name to store sources in")

	importCmd.Flags().String("engine", "Aria", "the database engine. Aria is recommended (requires MariaDB), MyISAM is supported for MySQL")
	importCmd.Flags().Bool("compress", false, "pack the database into a compressed, read-only format. Requires the Aria or MyISAM database engine")

	importCmd.Flags().Int("tmpFileLines", 4e6, "number of lines per temporary file (used for the LOAD FILE INTO command). 1e6 = ~32MB, 32e6 = ~1GB")
	importCmd.Flags().String("tmpFilePrefix", "[dbName]_", "temporary processed file prefix")
	importCmd.Flags().String("tmpFileSuffix", ".txt", "temporary processed file suffix")

	importCmd.Flags().String("errLog", "[dbName]_err.log", "log file for unparsed lines")
	importCmd.Flags().String("doneLog", "[dbName]_done.log", "log file for processed input files")
	importCmd.Flags().String("skipLog", "[dbName]_skip.log", "log file for skipped input files")

	importCmd.MarkFlagRequired("conn")
	importCmd.MarkFlagRequired("sourcesConn")
	v.BindPFlags(importCmd.Flags())
}

func runImport(cmd *cobra.Command, args []string) {
	l.I("import called")
}
