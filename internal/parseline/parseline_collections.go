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

		result.Password = r[1]

		// roughly check if it is an email address or a username
		idxAt := strings.Index(r[0], "@")
		idxDot := strings.LastIndex(r[0], ".")

		if idxAt > 0 && idxDot > idxAt+1 {
			// is email
			result.Email = r[0]
		} else {
			// not email
			result.Username = r[0]
		}

		return result, nil
	}
}
