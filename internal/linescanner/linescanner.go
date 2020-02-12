package linescanner

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"strings"
)

// LineScanner creates a bufio.Scanner from a file, decompressing the file if necessary.
func LineScanner(path string, callback func(string, *bufio.Scanner) error) error {
	if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tgz") {
		return TarGzLineScanner(path, callback)
	}
	return TextLineScanner(path, callback)
}

// TarGzLineScanner creates a bufio.Scanner from a .tar.gz file.
func TarGzLineScanner(path string, callback func(string, *bufio.Scanner) error) error {
	// open tar.gz
	tarGz, err := os.Open(path)
	if err != nil {
		return err
	}
	defer tarGz.Close()

	// decompress
	gzf, err := gzip.NewReader(tarGz)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(gzf)

	// loop through lines in tar.gz files
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// skip non-regular files
		if header.Typeflag != tar.TypeReg {
			continue
		}

		// iterate through the lines in the file
		lineScanner := bufio.NewScanner(tarReader)
		err = callback(header.Name, lineScanner)
		if err != nil {
			return err
		}
	}
	return nil
}

// TextLineScanner creates a bufio.Scanner from a plain text file.
func TextLineScanner(path string, callback func(string, *bufio.Scanner) error) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	lineScanner := bufio.NewScanner(file)
	return callback(path, lineScanner)
}
