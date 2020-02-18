package cmd

/**
 * Author: Matt Moran
 */

import (
	"database/sql"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/darkmattermatt/dumpdb/internal/config"
	"github.com/darkmattermatt/dumpdb/internal/sourceid"
	l "github.com/darkmattermatt/dumpdb/pkg/simplelog"
	"github.com/darkmattermatt/dumpdb/pkg/stringinslice"
	"github.com/spf13/cobra"
)

// the `search` command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search multiple dump databases simultaneously.",
	Long:  "",
	Run:   runSearch,
	PreRun: func(cmd *cobra.Command, args []string) {
		v.BindPFlags(cmd.Flags())
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Positional args: filesOrFolders: files and/or folders to import
	searchCmd.Flags().StringSliceP("databases", "d", []string{}, "comma separated list of databases to search")
	searchCmd.Flags().StringP("conn", "c", "", "connection string prefix to connect to MySQL databases. Like user:pass@tcp(127.0.0.1:3306)")
	searchCmd.Flags().StringP("sourcesConn", "C", "", "connection string for the sources database. Like user:pass@tcp(127.0.0.1:3306)/sources")

	searchCmd.Flags().StringP("query", "Q", "", "the WHERE clause of a SQL query. Yes it's injected, so try not to break your own database")
	searchCmd.Flags().StringSlice("columns", []string{}, "columns to retrieve from the database")

	searchCmd.MarkFlagRequired("databases")
	searchCmd.MarkFlagRequired("conn")
	searchCmd.MarkFlagRequired("query")
}

func loadSearchConfig(cmd *cobra.Command) {
	c.Databases = v.GetStringSlice("databases")
	c.SourcesConn = v.GetString("sourcesConn")
	c.Query = v.GetString("query")

	// c.Columns = v.GetStringSlice("columns")
	dbCols := []string{"email", "hash", "password", "sourceid", "username"}
	if len(c.Columns) == 0 {
		c.Columns = dbCols
		if c.SourcesConn != "" {
			c.Columns = append(c.Columns, "source")
		}
	} else {
		for _, col := range v.GetStringSlice("columns") {
			col = strings.ToLower(col)
			if !stringinslice.StringInSlice(col, dbCols) && col != "source" {
				config.ShowUsage(cmd, "Invalid column name: "+col)
			}
			c.Columns = append(c.Columns, col)
		}
	}

	c.Conn = v.GetString("conn")
	if !config.ValidDSNConn(c.Conn) {
		config.ShowUsage(cmd, "Invalid MySQL connection string "+c.Conn+". It must look like user:pass@tcp(127.0.0.1:3306)")
	}
	if strings.HasSuffix(c.Conn, ")") {
		c.Conn += "/"
	}
}

func runSearch(cmd *cobra.Command, args []string) {
	loadSearchConfig(cmd)

	if stringinslice.StringInSlice("source", c.Columns) {
		if c.SourcesConn == "" {
			config.ShowUsage(cmd, "Parameter sourcesConn must be set when requesting the `source` column. Use `sourceId` to get the unique source ID number.")
		}
		var err error
		sourcesDb, err = sql.Open("mysql", c.SourcesConn)
		l.FatalOnErr(err)
	}

	l.I("Querying", len(c.Databases), "databases:", strings.Join(c.Databases, ", "))
	l.V("Output format is tab-delimited as:")
	l.V(strings.Join(c.Columns, "\t"))

	var wg sync.WaitGroup
	for _, dbName := range c.Databases {
		wg.Add(1)
		go queryDatabase(dbName, &wg)
	}
	wg.Wait()
}

// preferUsingEmailRev replaces queries using the `email` column with queries using the `email_rev` column
func preferUsingEmailRev(stmt string) string {
	/** Prefer using email_rev column because it is indexed
	 * - replaces               | + with
	 * email like '%@gmail.com' | email_rev like REVERSE('%@gmail.com')
	 * email = "test@gmail.com" | email_rev = REVERSE("test@gmail.com")
	 * email >= 'abc@gmail.com' | email_rev >= REVERSE('abc@gmail.com')
	 * email LIKE 'DarkMatter%' | email_rev LIKE REVERSE('DarkMatter%') <- this will _not_ use the email_rev index
	 */
	return regexp.MustCompile("(?i)email\\s*(LIKE|[<>!=]{1,2})\\s*('[^']*'|\"[^\"]*\")").ReplaceAllString(stmt, "email_rev $1 REVERSE($2)")
}

func queryDatabase(dbName string, wg *sync.WaitGroup) {
	defer wg.Done()

	dbConn := c.Conn + dbName
	l.D("queryDatabase", "dbConn:", dbConn)
	db, err := sql.Open("mysql", dbConn)
	if err != nil {
		l.W(err)
		return
	}
	defer db.Close()

	q := "SELECT email, hash, password, sourceid, username FROM main WHERE " + c.Query
	l.D("queryDatabase", dbName, q)

	rows, err := db.Query(q)
	if err != nil {
		l.W(err)
		return
	}
	defer rows.Close()

	var (
		email    string
		hash     string
		password string
		sourceID int64
		username string
	)

	for rows.Next() {
		err := rows.Scan(&email, &hash, &password, &sourceID, &username)
		if err != nil {
			l.W(err)
			return
		}

		var arr []string
		for _, col := range c.Columns {
			switch col {
			case "email":
				arr = append(arr, email)
			case "hash":
				arr = append(arr, hash)
			case "password":
				arr = append(arr, password)
			case "source":
				s, err := sourceid.SourceName(sourceID, sourcesDb, sourcesTable)
				if err != nil {
					l.W(err)
					return
				}
				arr = append(arr, s)
			case "sourceid":
				arr = append(arr, strconv.FormatInt(sourceID, 10))
			case "username":
				arr = append(arr, username)
			}
		}
		// print result to stdout
		l.R(strings.Join(arr, "\t"))
	}

	err = rows.Err()
	if err != nil {
		l.W(err)
		return
	}
}
