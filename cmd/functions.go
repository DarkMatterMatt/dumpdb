package cmd

/**
 * Author: Matt Moran
 */

import (
	"os"
	"os/exec"
	"path/filepath"

	l "github.com/darkmattermatt/dumpdb/pkg/simplelog"
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
	l.FatalOnErr(err)
	l.D("dataDir: " + dataDir)
	return
}

func disableDatabaseIndexes(dataDir string) {
	l.V("Disabling database indexes")
	out, err := exec.Command("aria_chk", "-rq", "--keys-used", "0", dataDir+c.Database+"/"+mainTable).CombinedOutput()
	l.D(string(out))
	l.FatalOnErr(err)
}

func restoreDatabaseIndexes(dataDir, tmpDir string) {
	l.V("Indexing database")
	out, err := exec.Command("aria_pack", "--tmpdir", tmpDir, dataDir+c.Database+"/"+mainTable).CombinedOutput()
	l.D(string(out))
	l.FatalOnErr(err)
}

func compressDatabase(dataDir, tmpDir string) {
	l.V("Compressing database")
	out, err := exec.Command("aria_chk", "-rq", "--tmpdir", tmpDir, dataDir+c.Database+"/"+mainTable).CombinedOutput()
	l.D(string(out))
	l.FatalOnErr(err)
}

func importToDatabase(filename string, mysqlDone chan bool) {
	filename, err := filepath.Abs(filename)
	l.FatalOnErr(err)

	l.I("Importing " + filename + " to the database")
	_, err = db.Exec(`
		LOAD DATA INFILE '` + filename + `'
		IGNORE INTO TABLE ` + mainTable + `
		FIELDS TERMINATED BY '\t' ESCAPED BY ''
		LINES TERMINATED BY '\n'
		(sourceid, username, email_rev, hash, password, extra)
	`)
	l.FatalOnErr(err)
	mysqlDone <- true

	// delete file we just loaded
	err = os.Remove(filename)
	l.WarnOnErr(err)
}

func waitForImport(mysqlDone chan bool) {
	l.D("Waiting for a database load to finish")
	<-mysqlDone
}
