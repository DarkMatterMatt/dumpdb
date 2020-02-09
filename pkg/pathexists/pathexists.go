package pathexists

import (
	"errors"
	"os"
)

// PathExists checks that a path is a file or directory
func PathExists(path string) bool {
	return AssertPathExists(path) != nil
}

// AssertPathExists validates that a path is a file or directory
func AssertPathExists(path string) error {
	_, err := os.Stat(path)
	return err
}

// AssertPathsAllExist validates that all paths are files or directories
func AssertPathsAllExist(paths []string) error {
	for _, path := range paths {
		if err := AssertPathExists(path); err != nil {
			return err
		}
	}
	return nil
}

// IsFile checks that a path is a file
func IsFile(path string) bool {
	return !IsDir(path)
}

// AssertPathIsFile validates that the path is a file
func AssertPathIsFile(path string) error {
	if !IsFile(path) {
		return errors.New(path + " is not a file!")
	}
	return nil
}

// AssertPathsAreFiles validates that all paths are files
func AssertPathsAreFiles(paths []string) error {
	for _, path := range paths {
		if err := AssertPathIsFile(path); err != nil {
			return err
		}
	}
	return nil
}

// IsDir checks that a path is a directory
func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// AssertPathIsDir validates that the path is a directory
func AssertPathIsDir(path string) error {
	if !IsFile(path) {
		return errors.New(path + " is not a directory!")
	}
	return nil
}

// AssertPathsAreDirectories validates that all paths are directories
func AssertPathsAreDirectories(paths []string) error {
	for _, path := range paths {
		if err := AssertPathIsDir(path); err != nil {
			return err
		}
	}
	return nil
}
