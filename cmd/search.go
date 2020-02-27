package cmd

import (
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/darkmattermatt/dumpdb/internal/config"
	"github.com/darkmattermatt/dumpdb/internal/parseline"
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
	Args: func(cmd *cobra.Command, args []string) error {
		databases, _ := cmd.Flags().GetStringSlice("databases")
		if len(args) < 1 && len(databases) == 0 {
			return errors.New("Missing database names to search")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Positional args: databases: the names of databases to search. Also support using -d flag
	searchCmd.Flags().StringP("conn", "c", "", "connection string to connect to MySQL databases. Like user:pass@tcp(127.0.0.1:3306)")
	searchCmd.Flags().StringSliceP("databases", "d", []string{}, "comma separated list of databases to search")
	searchCmd.Flags().StringP("sourcesDatabase", "s", "", "database name to resolve sourceIDs to their names from")

	searchCmd.Flags().StringP("query", "Q", "", "the WHERE clause of a SQL query. Yes it's injected, so try not to break your own database")
	searchCmd.Flags().StringP("format", "f", "text", "the output format")
	searchCmd.Flags().StringSliceP("columns", "C", []string{}, "comma separated list of columns to retrieve")

	searchCmd.MarkFlagRequired("conn")
	searchCmd.MarkFlagRequired("query")
}

func loadSearchConfig(cmd *cobra.Command, databases []string) {
	c.Databases = append(v.GetStringSlice("databases"), databases...)
	c.SourcesDatabase = v.GetString("sourcesDatabase")
	c.Query = preferUsingEmailRev(v.GetString("query"))

	c.OutputFormat = strings.ToLower(v.GetString("format"))
	if _, ok := searchPerRecordCallbacks[c.OutputFormat]; !ok {
		showUsage(cmd, "Invalid output format: "+c.OutputFormat+". Options are: text, json")
	}

	requestedCols := v.GetStringSlice("columns")
	dbCols := []string{"email", "hash", "password", "sourceid", "username", "extra"}
	if len(requestedCols) == 0 {
		c.Columns = dbCols
		if c.SourcesDatabase != "" {
			c.Columns = append(c.Columns, "source")
		}
	} else {
		for _, col := range requestedCols {
			col = strings.ToLower(col)
			if col == "source" {
				if c.SourcesDatabase == "" {
					showUsage(cmd, "Parameter sourcesDatabase must be set when requesting the `source` column. Use `sourceId` to get the unique source ID number.")
				}
			} else if !stringinslice.StringInSlice(col, dbCols) {
				showUsage(cmd, "Invalid column name: "+col+". Options are: "+strings.Join(append(dbCols, "source"), ", "))
			}
			c.Columns = append(c.Columns, col)
		}
	}

	c.Conn = v.GetString("conn")
	if !config.ValidDSNConn(c.Conn) {
		showUsage(cmd, "Invalid MySQL connection string "+c.Conn+". It must look like user:pass@tcp(127.0.0.1:3306)")
	}
	c.Conn += "/"
}

func runSearch(cmd *cobra.Command, databases []string) {
	loadSearchConfig(cmd, databases)

	if c.SourcesDatabase != "" {
		var err error
		sourcesDb, err = sql.Open("mysql", c.Conn+c.SourcesDatabase)
		l.FatalOnErr("Opening sources database", err)
	}

	l.I("Querying", len(c.Databases), "databases:", strings.Join(c.Databases, ", "))
	l.V("Output format is tab-delimited as:\n    " + strings.Join(c.Columns, "\t"))

	var wg sync.WaitGroup
	for _, dbName := range c.Databases {
		wg.Add(1)
		go queryDatabase(dbName, &wg, searchPerRecordCallbacks[c.OutputFormat])
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

var searchPerRecordCallbacks = make(map[string]func(r *parseline.Record) error)

func init() {
	searchPerRecordCallbacks["text"] = func(r *parseline.Record) error {
		var arr []string
		for _, col := range c.Columns {
			switch col {
			case "email":
				arr = append(arr, r.Email)
			case "hash":
				arr = append(arr, r.Hash)
			case "password":
				arr = append(arr, r.Password)
			case "source":
				s, err := sourceid.SourceName(r.SourceID, sourcesDb, sourcesTable)
				if err != nil {
					return err
				}
				arr = append(arr, s)
			case "sourceid":
				arr = append(arr, strconv.FormatInt(r.SourceID, 10))
			case "username":
				arr = append(arr, r.Username)
			case "extra":
				arr = append(arr, r.Extra)
			}
		}
		// print result to stdout
		l.R(strings.Join(arr, "\t"))
		return nil
	}

	searchPerRecordCallbacks["jsonl"] = func(r *parseline.Record) error {
		var arr []string
		for _, col := range c.Columns {
			switch col {
			case "email":
				tmp, err := json.Marshal(r.Email)
				if err != nil {
					return err
				}
				arr = append(arr, "\"email\":"+string(tmp))
			case "hash":
				tmp, err := json.Marshal(r.Hash)
				if err != nil {
					return err
				}
				arr = append(arr, "\"hash\":"+string(tmp))
			case "password":
				tmp, err := json.Marshal(r.Password)
				if err != nil {
					return err
				}
				arr = append(arr, "\"password\":"+string(tmp))
			case "source":
				s, err := sourceid.SourceName(r.SourceID, sourcesDb, sourcesTable)
				if err != nil {
					return err
				}
				tmp, err := json.Marshal(s)
				if err != nil {
					return err
				}
				arr = append(arr, "\"source\":"+string(tmp))
			case "sourceid":
				arr = append(arr, "\"sourceid\":"+strconv.FormatInt(r.SourceID, 10))
			case "username":
				tmp, err := json.Marshal(r.Username)
				if err != nil {
					return err
				}
				arr = append(arr, "\"username\":"+string(tmp))
			case "extra":
				tmp, err := json.Marshal(r.Extra)
				if err != nil {
					return err
				}
				arr = append(arr, "\"extra\":"+string(tmp))
			}
		}
		// print result to stdout
		l.R("{" + strings.Join(arr, ",") + "}")
		return nil
	}
}
