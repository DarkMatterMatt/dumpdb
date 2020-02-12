package cmd

/**
 * Author: Matt Moran
 */

import (
	"database/sql"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/darkmattermatt/DumpDB/internal/sourceid"
	"github.com/darkmattermatt/DumpDB/pkg/stringinslice"
	l "github.com/darkmattermatt/dumpdb/pkg/simplelog"
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
	searchCmd.Flags().StringSliceP("databases", "d", []string{}, "comma separated list of databases to search")
	searchCmd.Flags().StringP("connPrefix", "c", "", "connection string prefix to connect to MySQL databases. Like user:pass@tcp(127.0.0.1:3306)")
	searchCmd.Flags().StringP("dbTable", "t", "main", "database table name to search. Must be the same for all databases")
	searchCmd.Flags().StringP("sourcesConn", "C", "", "connection string for the sources database. Like user:pass@tcp(127.0.0.1:3306)/sources")
	searchCmd.Flags().StringP("sourcesTable", "T", "sources", "SQL connection string for the sources database. Like user:pass@tcp(127.0.0.1:3306)/sources")

	searchCmd.Flags().StringP("query", "Q", "", "the WHERE clause of a SQL query. Yes it's injected, so try not to break your own database")
	searchCmd.Flags().StringSlice("columns", []string{}, "columns to retrieve from the database")

	searchCmd.MarkFlagRequired("databases")
	searchCmd.MarkFlagRequired("connPrefix")
	searchCmd.MarkFlagRequired("query")

	v.BindPFlags(searchCmd.Flags())
}

func loadSearchConfig() error {
	c.Databases = v.GetStringSlice("databases")
	c.ConnPrefix = v.GetString("connPrefix")
	c.DbTable = v.GetString("dbTable")
	c.SourcesConn = v.GetString("sourcesConn")
	c.SourcesTable = v.GetString("sourcesTable")
	c.Query = v.GetString("query")
	c.Columns = v.GetStringSlice("columns")

	validCols := []string{"email", "hash", "password", "source", "sourceid", "username"}
	if len(c.Columns) == 0 {
		c.Columns = validCols
	} else {
		for _, col := range c.Columns {
			if !stringinslice.StringInSlice(col, validCols) {
				return errors.New("Invalid column name: " + col)
			}
		}
	}
	return nil
}

func runSearch(cmd *cobra.Command, args []string) {
	err := loadSearchConfig()
	l.FatalOnErr(err)

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

	dbConn := c.ConnPrefix + dbName
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

		var b strings.Builder
		for _, col := range c.Columns {
			switch col {
			case "email":
				b.WriteString(email)
			case "hash":
				b.WriteString(hash)
			case "password":
				b.WriteString(password)
			case "source":
				s, err := sourceid.SourceName(sourceID, sourcesDb, c.SourcesTable)
				if err != nil {
					l.W(err)
					return
				}
				b.WriteString(s)
			case "sourceID":
				b.WriteString(strconv.FormatInt(sourceID, 10))
			case "username":
				b.WriteString(username)
			}
			b.WriteString("\t")
		}
		// trim trailing tab, then print
		l.R(b.String()[:b.Len()-1])
	}

	err = rows.Err()
	if err != nil {
		l.W(err)
		return
	}
}
