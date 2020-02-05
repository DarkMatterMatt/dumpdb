package pathexists

import (
	"os"
)

// PathExists validates that a path is a file or directory
func PathExists(path string) error {
	_, err := os.Stat(path)
	return err
}

// PathsAllExist validates that all paths are files or directories
func PathsAllExist(paths []string) error {
	for _, path := range paths {
		if err := PathExists(path); err != nil {
			return err
		}
	}
	return nil
}
