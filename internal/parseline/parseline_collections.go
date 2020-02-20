package parseline

import (
	"errors"
	"strings"
)

func init() {
	lineParsers["collections"] = func(line, source string) (Record, error) {
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

		result.Email = r[0]
		result.Password = r[1]
		return result, nil
	}
}
