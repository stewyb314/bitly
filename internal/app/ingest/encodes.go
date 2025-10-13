package ingest

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	LongUrlHeader = "long_url"
	DomainHeader  = "domain"
	HashHeader    = "hash"
)

type encodes struct {
	url    string
	domain string
	hash   string
}

type EncodesData struct {
}

func (e *EncodesData) Read(path string) ([]encodes, error) {

	file, err := os.Open(path)
	if err != nil {
		return []encodes{}, fmt.Errorf("opening encoding file: %w", err)
	}
	defer file.Close()
	entries, err := readEncodes(csv.NewReader(file))
	if err != nil {
		return []encodes{}, err
	}
	return entries, err
}

func readEncodes(r *csv.Reader) ([]encodes, error) {
	entries := make([]encodes, 0)
	// verify expected CSV headers
	records, err := r.ReadAll()
	if err != nil {
		return entries, fmt.Errorf("reading csv headers: %w", err)
	}
	headers := records[0]
	if len(headers) != 3 || headers[0] != LongUrlHeader ||
		headers[1] != DomainHeader || headers[2] != HashHeader {
		return entries, errors.New("csv headers: invalid format")
	}

	for _, record := range records[1:] {
		entries = append(entries, encodes{
			url:    strings.TrimSpace(record[0]),
			domain: strings.TrimSpace(record[1]),
			hash:   strings.TrimSpace(record[2]),
		})
	}
	return entries, nil
}
