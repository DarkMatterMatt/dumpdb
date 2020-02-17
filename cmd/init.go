package cmd

/**
 * Author: Matt Moran
 */

import (
	"database/sql"
	"errors"
	"os"
	"strings"
	"time"

	l "github.com/darkmattermatt/dumpdb/pkg/simplelog"
	"github.com/darkmattermatt/dumpdb/pkg/stringinslice"
	"github.com/spf13/cobra"
)

const schemaVersion = "0.0.2"

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
	initCmd.Flags().StringP("sources", "s", "", "initialise the sources directory")
	initCmd.Flags().String("engine", "Aria", "the database engine. Aria is recommended (requires MariaDB), MyISAM is supported for MySQL")

	initCmd.MarkFlagRequired("conn")
}

func loadInitConfig(cmd *cobra.Command) {
	c.Sources = v.GetString("sources")
	c.Engine = strings.ToLower(v.GetString("engine"))

	validEngines := []string{"aria", "myisam"}
	if !stringinslice.StringInSlice(c.Engine, validEngines) {
		l.R("Error: unknown database engine: " + c.Engine + ". Valid options are: " + strings.Join(validEngines, ", ") + "\n")
		cmd.Usage()
		os.Exit(1)
	}

	c.Conn = v.GetString("conn")
	if !strings.HasSuffix(c.Conn, "/") {
		c.Conn += "/"
	}
}

func runInit(cmd *cobra.Command, dbNames []string) {
	loadInitConfig(cmd)

	var err error
	l.D("Using MySQL connection string: " + c.Conn)
	db, err = sql.Open("mysql", c.Conn)
	l.FatalOnErr(err)

	if c.Sources != "" {
		err = createDatabase(c.Sources)
		l.FatalOnErr(err)

		err = createSourcesTable(c.Sources, c.Engine)
		l.FatalOnErr(err)
	}

	metadata := map[string]string{
		"schema_version": schemaVersion,
		"created":        time.Now().Format("2006-01-02 15:04"),
	}

	for _, dbName := range dbNames {
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
	l.D("createDatabase: " + dbName)
	_, err := db.Exec(`
		CREATE DATABASE ` + dbName + `
	`)
	return err
}

func createMainTable(dbName, engine string) error {
	l.D("createMainTable: " + dbName + "/" + mainTable)
	_, err := db.Exec(`USE ` + dbName)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE ` + mainTable + ` (
			id              INT UNSIGNED        AUTO_INCREMENT,
			hash            VARCHAR(256),
			password        VARCHAR(128),
			source          INT UNSIGNED,
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
	l.D("createSourcesTable: " + dbName + "/" + sourcesTable)
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
	l.D("createMetadataTable: " + dbName + "/" + metadataTable)
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
	l.D("addMetadata:" + dbName)
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
