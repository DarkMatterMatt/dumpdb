package splitfilewriter

import (
	"math"
	"os"
	"strconv"
	"testing"
)

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

// TestSplitFileWriter tests the splitfilewriter
func TestSplitFileWriter(t *testing.T) {
	const numWritesPerFile = 10
	const writeCount = 888
	const testString = "Testing abcdefghijkmnopqrstuvwxyz: "
	dir := os.TempDir() + "/go/github.com/darkmattermatt/splitfilewriter"
	os.MkdirAll(dir, 0777)
	prefix := dir + "/test"

	// write files
	s, err := Create(prefix, ".txt", numWritesPerFile)
	checkErr(t, err)
	for i := 0; i < writeCount; i++ {
		_, err = s.WriteString(testString + strconv.Itoa(i) + "\n")
		checkErr(t, err)
	}
	err = s.Flush()
	checkErr(t, err)

	// read files
	for i := 0; i < writeCount; {
		f, err := os.Open(prefix + strconv.Itoa(i/numWritesPerFile) + ".txt")
		checkErr(t, err)

		b := make([]byte, (len(testString))*numWritesPerFile*2)
		n, err := f.Read(b)
		checkErr(t, err)

		var str string
		for j := 0; j < numWritesPerFile && i < writeCount; j++ {
			str += testString + strconv.Itoa(i) + "\n"
			i++
		}

		if str != string(b[:n]) {
			t.Errorf("Strings did not match. Expected %s, found %s", str, b)
		}

		err = f.Close()
		checkErr(t, err)
	}

	// delete files, fail silently
	for i := 0; i < int(math.Ceil(float64(writeCount)/float64(numWritesPerFile))); i++ {
		_ = os.Remove(prefix + strconv.Itoa(i) + ".txt")
	}
}
