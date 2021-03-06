package parseline

import (
	"errors"
	"strings"
)

func init() {
	lineParsers["example_split_simple"] = func(line, source string) (Record, error) {
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
		return result, nil
	}
}
