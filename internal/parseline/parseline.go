package parseline

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/darkmattermatt/dumpdb/internal/getsourceid"
)

// ParseLine parses a single line and returns a Record
// Modify this function to match your data
func ParseLine(line, source string, sourceDb *sql.DB, sourceTable string) (Record, error) {
	result := Record{}

	// try splitting by ;
	r := strings.SplitN(line, ";", 2)

	// if that doesn't work, try splitting by :
	if len(r) == 1 {
		r = strings.SplitN(line, ":", 2)
	}

	// if that doesn't work, try splitting by tabs
	if len(r) == 1 {
		r = strings.SplitN(line, "\t", 2)
	}

	// if that also failed, return an error
	if len(r) == 1 {
		return result, errors.New("Incorrect number of columns")
	}

	// check for presence of an @ symbol in the email address
	if !strings.Contains(r[0], "@") {
		return result, errors.New("Email address is missing")
	}

	// first field is email, second field is password
	// set either emailRev or email
	result.Email = r[0]
	// result.EmailRev = reverse(r[0])
	result.Password = r[1]
	result.Source = source
	result.SourceID = getsourceid.GetSourceID(source, sourceDb, sourceTable)
	return result, nil
}
