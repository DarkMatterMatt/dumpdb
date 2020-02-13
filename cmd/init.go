package cmd

/**
 * Author: Matt Moran
 */

import (
	"database/sql"
	"errors"
	"os"
	"strings"

	l "github.com/darkmattermatt/dumpdb/pkg/simplelog"
	"github.com/darkmattermatt/dumpdb/pkg/stringinslice"
	"github.com/spf13/cobra"
)

// the `init` command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise a database to prepare for importing.",
	Long:  "",
	Run:   runInit,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Missing database names to process")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Positional args: databaseNames: the names of databases to initialise
	initCmd.Flags().StringP("conn", "c", "", "connection string for the MySQL. Like user:pass@tcp(127.0.0.1:3306)")
	initCmd.Flags().StringP("table", "t", "main", "database table name to insert into")
	initCmd.Flags().StringP("sources", "s", "", "initialise the sources directory")
	initCmd.Flags().StringP("sourcesTable", "T", "sources", "database table name to store sources in")
	initCmd.Flags().String("engine", "Aria", "the database engine. Aria is recommended (requires MariaDB), MyISAM is supported for MySQL")

	initCmd.MarkFlagRequired("conn")

	v.BindPFlags(initCmd.Flags())
}

func loadInitConfig(cmd *cobra.Command) {
	c.Table = v.GetString("table")
	c.Sources = v.GetString("sources")
	c.SourcesTable = v.GetString("sourcesTable")
	c.Engine = strings.ToLower(v.GetString("engine"))

	validEngines := []string{"aria", "myisam", "innodb"}
	if !stringinslice.StringInSlice(c.Engine, validEngines) {
		l.R("Error: unknown database engine: " + c.Engine + ". Valid options are: " + strings.Join(validEngines, ", ") + "\n")
		cmd.Usage()
		os.Exit(1)
	}

	c.ConnPrefix = v.GetString("conn")
	if !strings.HasSuffix("/", c.ConnPrefix) {
		c.ConnPrefix += "/"
	}
}

func runInit(cmd *cobra.Command, dbNames []string) {
	loadInitConfig(cmd)

	var err error
	db, err = sql.Open("mysql", c.ConnPrefix)
	l.FatalOnErr(err)

	err = initDatabases(dbNames)
	l.FatalOnErr(err)

	if c.Sources != "" {
		err = initSourcesTable(c.Sources, c.SourcesTable, c.Engine)
		l.FatalOnErr(err)
	}

	for _, dbName := range dbNames {
		err = initTables(dbName, c.Table, c.Engine)
		l.FatalOnErr(err)
	}
}

func initDatabases(dbNames []string) error {
	for _, dbName := range dbNames {
		_, err := db.Exec(`
			CREATE DATABASE IF NOT EXISTS ` + dbName + `
		`)
		if err != nil {
			return err
		}
	}
	return nil
}

func initTables(dbName, tableName, engine string) error {
	_, err := db.Exec(`USE ` + dbName)
	if err != nil {
		return err
	}

	l.V("Creating " + dbName + "/" + tableName)
	_, err = db.Exec(`
		CREATE TABLE ` + tableName + ` (
			id              INT UNSIGNED        AUTO_INCREMENT,
			hash            VARCHAR(256),
			password        VARCHAR(128),
			source          INT UNSIGNED,
			email           VARCHAR(320)        GENERATED ALWAYS AS (REVERSE(email_rev)) VIRTUAL,
			email_rev       VARCHAR(320),       /* max length 320 https://stackoverflow.com/a/574698/6595777 */
			username        VARCHAR(128),

			PRIMARY KEY     (id)
		)
		CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_unicode_ci' ENGINE '` + engine + `'
	`)
	return err
}

func initSourcesTable(dbName, tableName, engine string) error {
	_, err := db.Exec(`USE ` + dbName)
	if err != nil {
		return err
	}

	l.V("Creating " + dbName + "/" + tableName)
	_, err = db.Exec(`
		CREATE TABLE ` + tableName + ` (
			id              INT UNSIGNED        AUTO_INCREMENT,
			name            VARCHAR(250),       /* 250 is the max length that can be indexed */
			last_updated    BIGINT              COMMENT 'Unix timestamp (seconds since 00:00:00 UTC 1 January 1970)',

			UNIQUE          (name),
			PRIMARY KEY     (id)
		)
		CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_unicode_ci' ENGINE '` + engine + `'
	`)
	return err
}
