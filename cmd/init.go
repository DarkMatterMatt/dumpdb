package cmd

/**
 * Author: Matt Moran
 */

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/darkmattermatt/dumpdb/internal/config"
	l "github.com/darkmattermatt/dumpdb/pkg/simplelog"
	"github.com/darkmattermatt/dumpdb/pkg/stringinslice"
	"github.com/spf13/cobra"
)

const schemaVersion = "0.0.3"

// the `init` command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise a database to prepare for importing.",
	Long:  "",
	Run:   runInit,
	PreRun: func(cmd *cobra.Command, args []string) {
		v.BindPFlags(cmd.Flags())
	},
	Args: func(cmd *cobra.Command, args []string) error {
		databases, _ := cmd.Flags().GetStringSlice("databases")
		sourcesDatabase, _ := cmd.Flags().GetString("sourcesDatabase")
		if len(args) < 1 && len(databases) == 0 && sourcesDatabase == "" {
			return errors.New("Missing database names to process")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Positional args: databases: the names of databases to initialise. Also support using -d flag
	initCmd.Flags().StringSliceP("databases", "d", []string{}, "comma separated list of databases to initialise")
	initCmd.Flags().StringP("conn", "c", "", "connection string for the MySQL. Like user:pass@tcp(127.0.0.1:3306)")
	initCmd.Flags().StringP("sourcesDatabase", "s", "", "initialise the sources database")
	initCmd.Flags().String("engine", "aria", "the database engine. Aria is recommended (requires MariaDB), MyISAM is supported for MySQL")

	initCmd.MarkFlagRequired("conn")
}

func loadInitConfig(cmd *cobra.Command, databases []string) {
	c.Databases = append(v.GetStringSlice("databases"), databases...)
	c.SourcesDatabase = v.GetString("sourcesDatabase")
	c.Engine = strings.ToLower(v.GetString("engine"))

	validEngines := []string{"aria", "myisam"}
	if !stringinslice.StringInSlice(c.Engine, validEngines) {
		showUsage(cmd, "Error: unknown database engine: "+c.Engine+". Valid options are: "+strings.Join(validEngines, ", ")+"\n")
	}

	c.Conn = v.GetString("conn")
	if !config.ValidDSNConn(c.Conn) {
		showUsage(cmd, "Invalid MySQL connection string "+c.Conn+". It must look like user:pass@tcp(127.0.0.1:3306)")
	}
	c.Conn += "/"
}

func runInit(cmd *cobra.Command, databases []string) {
	loadInitConfig(cmd, databases)

	var err error
	l.D("Using MySQL connection string: " + c.Conn)
	db, err = sql.Open("mysql", c.Conn)
	l.FatalOnErr(err)

	if c.SourcesDatabase != "" {
		err = createDatabase(c.SourcesDatabase)
		l.FatalOnErr(err)

		err = createSourcesTable(c.SourcesDatabase, c.Engine)
		l.FatalOnErr(err)
	}

	metadata := map[string]string{
		"schema_version": schemaVersion,
		"created":        time.Now().Format("2006-01-02 15:04"),
	}

	for _, dbName := range c.Databases {
		err = createDatabase(dbName)
		l.FatalOnErr(err)

		err = createMainTable(dbName, c.Engine)
		l.FatalOnErr(err)

		err = createMetadataTable(dbName, c.Engine)
		l.FatalOnErr(err)

		err = addMetadata(dbName, metadata)
		l.FatalOnErr(err)
	}
}

func createDatabase(dbName string) error {
	l.I("createDatabase: " + dbName)
	_, err := db.Exec(`
		CREATE DATABASE ` + dbName + `
	`)
	return err
}

func createMainTable(dbName, engine string) error {
	l.I("createMainTable: " + dbName + "/" + mainTable)
	_, err := db.Exec(`USE ` + dbName)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE ` + mainTable + ` (
			id              INT UNSIGNED        AUTO_INCREMENT,
			hash            VARCHAR(256),
			password        VARCHAR(128),
			sourceid        INT UNSIGNED,
			email           VARCHAR(320)        GENERATED ALWAYS AS (REVERSE(email_rev)) VIRTUAL,
			email_rev       VARCHAR(320),       /* max length 320 https://stackoverflow.com/a/574698/6595777 */
			username        VARCHAR(128),
			extra        	VARCHAR(1024),		/* extra data that does not fit in an existing column, e.g. password hints */

			PRIMARY KEY     (id)
		)
		CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_unicode_ci' ENGINE '` + engine + `'
	`)
	return err
}

func createSourcesTable(dbName, engine string) error {
	l.I("createSourcesTable: " + dbName + "/" + sourcesTable)
	_, err := db.Exec(`USE ` + dbName)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE ` + sourcesTable + ` (
			id              INT UNSIGNED        AUTO_INCREMENT,
			name            VARCHAR(250),       /* 250 is the max length that can be indexed */

			UNIQUE          (name),
			PRIMARY KEY     (id)
		)
		CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_unicode_ci' ENGINE '` + engine + `'
	`)
	return err
}

func createMetadataTable(dbName, engine string) error {
	l.I("createMetadataTable: " + dbName + "/" + metadataTable)
	_, err := db.Exec(`USE ` + dbName)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE ` + metadataTable + ` (
			k	VARCHAR(128),
			v	VARCHAR(8192),

			PRIMARY KEY		(k)
		)
		CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_unicode_ci' ENGINE '` + engine + `'
	`)
	return err
}

func addMetadata(dbName string, data map[string]string) error {
	l.V("addMetadata: " + dbName)
	_, err := db.Exec(`USE ` + dbName)
	if err != nil {
		return err
	}

	var (
		placeholders []string
		args         []interface{}
	)
	for k, v := range data {
		placeholders = append(placeholders, "(?, ?)")
		args = append(args, k, v)
	}

	_, err = db.Exec(`
		INSERT INTO `+metadataTable+` (k, v)
		VALUES `+strings.Join(placeholders, ", ")+`
		ON DUPLICATE KEY UPDATE v=v
	`, args...)
	return err
}
