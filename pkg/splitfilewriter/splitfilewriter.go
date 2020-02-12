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

	CurrentFile *os.File
	CurrentBuf  *bufio.Writer

	WriteCount int
	CurrentInc int
}

// Create calls os.Create and then creates a new SplitFileWriter from it
func Create(namePrefix, nameSuffix string, maxWrites int) (*SplitFileWriter, error) {
	s, err := New(namePrefix, nameSuffix, 0, maxWrites, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666, defaultBufSize)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Open calls os.OpenFile and then creates a new SplitFileWriter from it
func Open(namePrefix, nameSuffix string, maxWrites, fileFlag int, filePerm os.FileMode) (*SplitFileWriter, error) {
	s, err := New(namePrefix, nameSuffix, 0, maxWrites, fileFlag, filePerm, defaultBufSize)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// New calls os.OpenFile and then creates a new SplitFileWriter from it, setting all struct members
func New(namePrefix, nameSuffix string, currentInc, maxWrites, fileFlag int, filePerm os.FileMode, bufSize int) (*SplitFileWriter, error) {
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

// Write writes the contents of p into the buffer.
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

// PrevFileName returns the path of the last file that was opened for writing.
func (s *SplitFileWriter) PrevFileName() string {
	if s.CurrentInc == 0 {
		return ""
	}
	return s.NamePrefix + strconv.Itoa(s.CurrentInc-1) + s.NameSuffix
}

// CurrentFileName returns the path of the current file open for writing.
func (s *SplitFileWriter) CurrentFileName() string {
	return s.NamePrefix + strconv.Itoa(s.CurrentInc) + s.NameSuffix
}

// NextFileName returns the path of the next file that will be opened for writing.
func (s *SplitFileWriter) NextFileName() string {
	return s.NamePrefix + strconv.Itoa(s.CurrentInc+1) + s.NameSuffix
}

// preWrite increments the writeCount and opens a new file if required
func (s *SplitFileWriter) preWrite() error {
	if s.WriteCount >= s.MaxWrites {
		err := s.Flush()
		if err != nil {
			return err
		}

		s.CurrentInc++
		n, err := New(s.NamePrefix, s.NameSuffix, s.CurrentInc, s.MaxWrites, s.FileFlag, s.FilePerm, s.CurrentBuf.Size())
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
	s.CurrentFile = n.CurrentFile
	s.CurrentBuf = n.CurrentBuf
	s.WriteCount = n.WriteCount
	s.CurrentInc = n.CurrentInc
}
