package splitfilewriter

import (
	"bufio"
	"io"
	"os"
	"strconv"
)

const (
	defaultBufSize = 4096
)

// SplitFileWriter is a bufio writer which writes to a different file every `n` writes
type SplitFileWriter struct {
	NamePrefix string
	NameSuffix string
	MaxWrites  int
	FileFlag   int
	FilePerm   os.FileMode
	BufSize    int

	CurrentFile *os.File
	CurrentBuf  *bufio.Writer

	WriteCount int
	CurrentInc int
}

// OpenFileNewWriter calls os.OpenFile and then creates a new SplitFileWriter from it
func OpenFileNewWriter(namePrefix, nameSuffix string, maxWrites, fileFlag int, filePerm os.FileMode) (*SplitFileWriter, error) {
	s, err := OpenFileNewWriterSize(namePrefix, nameSuffix, maxWrites, fileFlag, filePerm, defaultBufSize)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// OpenFileNewWriterSize calls os.OpenFile and then creates a new SplitFileWriter from it
func OpenFileNewWriterSize(namePrefix, nameSuffix string, maxWrites, fileFlag int, filePerm os.FileMode, bufSize int) (*SplitFileWriter, error) {
	return OpenFileNewWriterSizeInc(namePrefix, nameSuffix, 0, maxWrites, fileFlag, filePerm, bufSize)
}

// OpenFileNewWriterSizeInc calls os.OpenFile and then creates a new SplitFileWriter from it
func OpenFileNewWriterSizeInc(namePrefix, nameSuffix string, currentInc, maxWrites, fileFlag int, filePerm os.FileMode, bufSize int) (*SplitFileWriter, error) {
	name := namePrefix + strconv.Itoa(currentInc) + nameSuffix
	f, err := os.OpenFile(name, fileFlag, filePerm)
	if err != nil {
		return nil, err
	}

	w := bufio.NewWriterSize(f, bufSize)
	return &SplitFileWriter{
		NamePrefix:  namePrefix,
		NameSuffix:  nameSuffix,
		MaxWrites:   maxWrites,
		FileFlag:    fileFlag,
		FilePerm:    filePerm,
		BufSize:     bufSize,
		CurrentFile: f,
		CurrentBuf:  w,
		CurrentInc:  currentInc,
	}, nil
}

// Buffered returns the number of bytes that have been written into the current buffer.
func (s *SplitFileWriter) Buffered() int {
	return s.CurrentBuf.Buffered()
}

// Flush writes any buffered data to the underlying io.Writer.
func (s *SplitFileWriter) Flush() error {
	return s.CurrentBuf.Flush()
}

// ReadFrom implements io.ReaderFrom.
func (s *SplitFileWriter) ReadFrom(r io.Reader) (int64, error) {
	return s.CurrentBuf.ReadFrom(r)
}

// Size returns the size of the underlying buffer in bytes.
func (s *SplitFileWriter) Size() int {
	return s.CurrentBuf.Size()
}

// Reset discards any unflushed buffered data, clears any error, and resets b to write its output to w.
func (s *SplitFileWriter) Write(p []byte) (int, error) {
	err := s.preWrite()
	if err != nil {
		return 0, err
	}
	return s.CurrentBuf.Write(p)
}

// WriteByte writes a single byte.
func (s *SplitFileWriter) WriteByte(c byte) error {
	err := s.preWrite()
	if err != nil {
		return err
	}
	return s.CurrentBuf.WriteByte(c)
}

// WriteRune writes a single Unicode code point, returning the number of bytes written and any error.
func (s *SplitFileWriter) WriteRune(r rune) (int, error) {
	err := s.preWrite()
	if err != nil {
		return 0, err
	}
	return s.CurrentBuf.WriteRune(r)
}

// WriteString writes a string. It returns the number of bytes written.
func (s *SplitFileWriter) WriteString(st string) (int, error) {
	err := s.preWrite()
	if err != nil {
		return 0, err
	}
	return s.CurrentBuf.WriteString(st)
}

// preWrite increments the writeCount and opens a new file if required
func (s *SplitFileWriter) preWrite() error {
	if s.WriteCount >= s.MaxWrites {
		err := s.Flush()
		if err != nil {
			return err
		}

		s.CurrentInc++
		n, err := OpenFileNewWriterSizeInc(s.NamePrefix, s.NameSuffix, s.CurrentInc, s.MaxWrites, s.FileFlag, s.FilePerm, s.BufSize)
		if err != nil {
			return err
		}

		s.copyToSelf(n)
	}
	s.WriteCount++
	return nil
}

func (s *SplitFileWriter) copyToSelf(n *SplitFileWriter) {
	s.NamePrefix = n.NamePrefix
	s.NameSuffix = n.NameSuffix
	s.MaxWrites = n.MaxWrites
	s.FileFlag = n.FileFlag
	s.FilePerm = n.FilePerm
	s.BufSize = n.BufSize
	s.CurrentFile = n.CurrentFile
	s.CurrentBuf = n.CurrentBuf
	s.WriteCount = n.WriteCount
	s.CurrentInc = n.CurrentInc
}
