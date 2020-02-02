package parseline

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/darkmattermatt/dumpdb/internal/getsourceid"
)

// parseLineSplitExample parses a single line and returns a Record
// Rename this function to `parseLine` after modifying to suit your data
func parseLineSplitExample(line, source string, sourceDb *sql.DB, sourceTable string) (Record, error) {
	result := Record{}

	// try splitting by :
	r := strings.SplitN(line, ":", 2)

	// if it failed, return an error
	if len(r) == 1 {
		return result, errors.New("Incorrect number of columns")
	}

	// check for presence of an @ symbol in the email address
	if !strings.Contains(r[0], "@") {
		return result, errors.New("Email address is missing")
	}

	// first field is email, second field is password
	result.Email = r[0]
	result.Password = r[1]
	result.Source = source
	result.SourceID = getsourceid.GetSourceID(source, sourceDb, sourceTable)
	return result, nil
}
