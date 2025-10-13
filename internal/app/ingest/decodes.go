package ingest

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type decodesData struct {
	file    *os.File
	end     int64
	scanner *bufio.Scanner
}

// NewDecodesData creates a new decodesData struct with defaults set.
// This struct is used to read a range of offsets from the decode json file
func NewDecodesData(path string) (*decodesData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open decode file: %w", err)
	}
	return &decodesData{file: file}, nil
}

// SetScanRange sets the start and end range.  Scan() will use these
// values to deterime where to start reading and when to return EOF
func (d *decodesData) SetScanRange(start, end int64) error {
	d.end = end
	if d.file == nil {
		return fmt.Errorf("file not opened")
	}
	ret, err := d.file.Seek(start, io.SeekStart)
	if err != nil {
		return fmt.Errorf("seek file: %w", err)
	}

	if ret != start {
		return fmt.Errorf("seek file: expected %d got %d", start, ret)
	}

	d.scanner = bufio.NewScanner(d.file)
	return nil
}

// Scan reads one line at a time from the file.
// EOF is returned either when the current offsset exceeds
// the end offset or when the end of the file is reached
func (d *decodesData) Scan() (string, error) {
	if d.scanner == nil {
		return "", fmt.Errorf("scanner not initialized")
	}
	cur, _ := d.file.Seek(0, io.SeekCurrent)
	if cur >= d.end {
		return "", io.EOF
	}
	if !d.scanner.Scan() {
		return "", io.EOF
	}
	return d.scanner.Text(), d.scanner.Err()
}

// Close closes the unerlying file object
func (d *decodesData) Close() {
	if d.file != nil {
		d.file.Close()
	}
}
