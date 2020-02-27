package parseline

import (
	"errors"
	"strings"
)

func init() {
	lineParsers["adobe"] = func(line, source string) (Record, error) {
		result := Record{}

		r := strings.Split(line, "-|-")
		if len(r) < 5 {
			return result, errors.New("Incorrect number of columns")
		}

		// check for presence of an @ symbol in the email address
		if strings.Contains(r[2], "@") {
			result.Email = r[2]
		} else {
			result.Username = r[2]
		}

		result.Hash = r[3]
		result.Extra = strings.TrimRight(r[4], "-|")
		result.Source = "adobe"
		return result, nil
	}
}
