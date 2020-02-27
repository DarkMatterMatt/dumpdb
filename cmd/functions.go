package cmd

import (
	"bufio"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/darkmattermatt/dumpdb/internal/parseline"
	"github.com/darkmattermatt/dumpdb/internal/sourceid"
	"github.com/darkmattermatt/dumpdb/pkg/reverse"
	l "github.com/darkmattermatt/dumpdb/pkg/simplelog"
	"github.com/pbnjay/memory"
	"github.com/spf13/cobra"
)

// showUsage exits after printing an error message followed by the command's usage
func showUsage(cmd *cobra.Command, s string) {
	l.R(s)
	cmd.Usage()
	l.F(s)
}

func getDataDir() (dataDir string) {
	err := db.QueryRow("SELECT @@datadir").Scan(&dataDir)
	l.FatalOnErr("Querying location of MySQL databases", err)
	l.D("dataDir: " + dataDir)
	return
}

func formatCommandOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n\n", "\n")
	s = strings.ReplaceAll(s, "\n", "\n    ")
	s = strings.TrimSpace(s)
	return s
}

func queryDatabaseEngine() string {
	l.I("Querying database engine type")

	var engine string
	err := db.QueryRow(`
		SELECT engine
		FROM information_schema.tables
		WHERE table_name='` + mainTable + `' AND table_schema='` + c.Database + `'
	`).Scan(&engine)
	l.FatalOnErr("Querying database engine type", err)

	l.V("Found database engine: " + engine)
	return strings.ToLower(engine)
}

func disableDatabaseIndexes(dataDir string) {
	l.I("Disabling database indexes")

	packCmd := "aria_chk"
	if c.Engine == "myisam" {
		packCmd = "myisamchk"
	}

	out, err := exec.Command(packCmd, "-rq", "--keys-used", "0", dataDir+c.Database+"/"+mainTable).CombinedOutput()
	l.D(formatCommandOutput(string(out)))
	l.FatalOnErr("Disabling database indexes", err)
}

func restoreDatabaseIndexes(dataDir, tmpDir string) {
	l.I("Indexing database")

	mem := memory.TotalMemory()
	if mem != 0 {
		// TODO: Add configurable percentage
		l.V("Detected RAM: " + strconv.FormatUint(mem/1024/1024/1000, 10) + "GB. Using 25% as the sort buffer.")
		mem /= 4
	} else {
		l.V("Failed to detect the amount system RAM. Using 512MB as the sort buffer.")
		mem = 512 * 1024 * 1024
	}

	packCmd := "aria_chk"
	bufferParam := "--sort_buffer_size"
	if c.Engine == "myisam" {
		packCmd = "myisamchk"
		bufferParam = "--myisam_sort_buffer_size"
	}

	out, err := exec.Command(packCmd, "-rq", bufferParam, strconv.FormatUint(mem, 10), "--tmpdir", tmpDir, dataDir+c.Database+"/"+mainTable).CombinedOutput()
	l.D(formatCommandOutput(string(out)))
	l.FatalOnErr("Indexing database", err)
}

func compressDatabase(dataDir, tmpDir string) {
	l.I("Compressing database")

	packCmd := "aria_pack"
	if c.Engine == "myisam" {
		packCmd = "myisampack"
	}

	out, err := exec.Command(packCmd, "--tmpdir", tmpDir, dataDir+c.Database+"/"+mainTable).CombinedOutput()
	l.D(formatCommandOutput(string(out)))
	l.FatalOnErr("Compressing database", err)
}

func importToDatabase(filename string, mysqlDone chan bool) {
	filename, err := filepath.Abs(filename)
	l.FatalOnErr("Determining the absolute filepath of "+filename, err)

	l.I("Importing " + filename + " to the database")
	filename = strings.ReplaceAll(filename, "\\", "\\\\")

	_, err = db.Exec(`
		LOAD DATA INFILE '` + filename + `'
		IGNORE INTO TABLE ` + mainTable + `
		FIELDS TERMINATED BY '\t' ESCAPED BY ''
		LINES TERMINATED BY '\n'
		(sourceid, username, email_rev, hash, password, extra)
	`)
	l.FatalOnErr("Loading tmp file into database", err)
	mysqlDone <- true

	// delete file we just loaded
	err = os.Remove(filename)
	l.WarnOnErr("Removing tmp file "+filename, err)
}

func waitForImport(mysqlDone chan bool) {
	l.D("Waiting for a database load to finish")
	<-mysqlDone
}

func flushAndLockTables() {
	l.V("Flushing and locking the `" + mainTable + "` table")
	_, err := db.Exec(`
		FLUSH TABLES ` + mainTable + `
		FOR EXPORT
	`)
	l.FatalOnErr("Flushing and locking the `"+mainTable+"` table", err)
}

func unlockTables() {
	l.V("Unlocking the `" + mainTable + "` table")
	_, err := db.Exec(`
		UNLOCK TABLES
	`)
	l.FatalOnErr("Unlocking the `"+mainTable+"` table", err)
}

func processTextFileScanner(path string, lineScanner *bufio.Scanner, toImport bool) error {
	if !strings.HasSuffix(path, ".txt") && !strings.HasSuffix(path, ".csv") {
		l.V("Skipping: " + path)
		_, err := skipFile.WriteString(path + "\n")
		l.FatalOnErr("Writing to skip log", err)
		return nil
	}

	l.V("Processing: " + path)

	for lineScanner.Scan() {
		// CTRL+C means stop
		if signalInterrupt {
			return errSignalInterrupt
		}

		line := lineScanner.Text()
		// skip blank lines
		if line == "" {
			continue
		}

		// parse & reformat line
		r, err := parseline.ParseLine(c.LineParser, line, path)
		if err != nil {
			errFile.WriteString(line + "\n")
			continue
		}

		if r.EmailRev == "" && r.Email != "" {
			r.EmailRev = reverse.Reverse(r.Email)
		} else if r.Email == "" && r.EmailRev != "" {
			r.Email = reverse.Reverse(r.EmailRev)
		}

		var arr []string
		if toImport {
			r.SourceID, err = sourceid.SourceID(r.Source, sourcesDb, sourcesTable)
			l.FatalOnErr("Loading SourceID", err)
			arr = []string{strconv.FormatInt(r.SourceID, 10), r.Username, r.EmailRev, r.Hash, r.Password, r.Extra}
		} else {
			arr = []string{r.Source, r.Username, r.Email, r.Hash, r.Password, r.Extra}
		}

		// write string to output file
		_, err = outputFile.WriteString(strings.Join(arr, "\t") + "\n")
		l.FatalOnErr("Writing processed string to output file", err)
	}
	doneFile.WriteString(path + "\n")
	return nil
}

func queryDatabase(dbName string, wg *sync.WaitGroup, perRecordCallback func(*parseline.Record) error) error {
	defer wg.Done()

	dbConn := c.Conn + dbName
	l.D("queryDatabase", "dbConn:", dbConn)
	db, err := sql.Open("mysql", dbConn)
	if err != nil {
		return err
	}
	defer db.Close()

	q := "SELECT email, hash, password, sourceid, username, extra FROM main WHERE " + c.Query
	l.D("queryDatabase", dbName, q)

	rows, err := db.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		r := parseline.Record{}
		err := rows.Scan(&r.Email, &r.Hash, &r.Password, &r.SourceID, &r.Username, &r.Extra)
		if err != nil {
			return err
		}
		err = perRecordCallback(&r)
		if err != nil {
			return err
		}
	}
	return rows.Err()
}
