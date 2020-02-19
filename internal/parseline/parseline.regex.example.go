package parseline

import (
	"errors"
	"regexp"
	"strings"
)

/* Modify this line to match your data. Note that this is ~40% slower than using parseLine.split.example.go */
// named group 'email' with pattern     .+@.+\..+
// delimiter is                         [:;]
// named group 'password' with pattern  .*
var lineFormat = regexp.MustCompile(`(?P<email>.+@.+\..+)[:;](?P<password>.*)`)

func init() {
	lineParsers["regex"] = func(line, source string) (Record, error) {
		result := Record{}

		// match by regex
		match := lineFormat.FindStringSubmatch(line)
		if match == nil {
			// return an empty map if no match
			return result, errors.New("Regex match failed")
		}

		// map the regex groups into the result
		for i, name := range lineFormat.SubexpNames() {
			switch strings.ToLower(name) {
			case "username":
				result.Username = match[i]
			case "email":
				result.Email = match[i]
			case "emailrev":
				result.EmailRev = match[i]
			case "hash":
				result.Hash = match[i]
			case "password":
				result.Password = match[i]
			case "extra":
				result.Extra = match[i]
			}
		}
		result.Source = source
		return result, nil
	}
}
