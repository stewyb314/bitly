package ingest

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type DecodesData struct {
	file    *os.File
	end     int64
	scanner *bufio.Scanner
}

func NewDecodesData(path string) (*DecodesData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open decode file: %w", err)
	}
	return &DecodesData{file: file}, nil
}

func (d *DecodesData) SetScanRange(start, end int64) error {
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

func (d *DecodesData) Scan() (string, error) {
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

func (d *DecodesData) Close() {
	if d.file != nil {
		d.file.Close()
	}
}
