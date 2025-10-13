package config

import (
	"strconv"
	"time"
	"fmt"

	configv2 "github.com/gookit/config/v2"
)

const (
	DefaultThreads          = "4"
	DefaultChunkSizeInKB    = "200"
	DefaultEncodingFilePath = "data/encodes.csv"
	DefaultDecodesFilePath  = "data/decodes.json"
	DefaultDebug            = false
	DefaultStartTime        = "2021-01-01T00:00:00Z"
	DefaultEndTime          = "2021-12-31T23:59:59Z"
)

type Config struct {
	Threads          int
	ChunkSizeInKB    int
	EncodingFilePath string
	DecodesFilePath  string
	Debug            bool
	StartTime        time.Time
	EndTime          time.Time
}

func New() (*Config, error) {
	c := Config{}
	threads := configv2.GetEnv("THREADS", DefaultThreads)
	var err error
	if c.Threads, err = strconv.Atoi(threads); err != nil {
		return nil, err
	}

	chunkSize := configv2.GetEnv("CHUNK_SIZE_IN_KB", DefaultChunkSizeInKB)
	if c.ChunkSizeInKB, err = strconv.Atoi(chunkSize); err != nil {
		return nil, err
	}

	c.EncodingFilePath = configv2.GetEnv("ENCODING_FILE_PATH", DefaultEncodingFilePath)
	c.DecodesFilePath = configv2.GetEnv("DECODES_FILE_PATH", DefaultDecodesFilePath)
	c.StartTime, err = time.Parse(time.RFC3339, configv2.GetEnv("START_TIME", DefaultStartTime) )

	if err != nil {
		return nil, fmt.Errorf("parse start time: %w", err)
	}

	c.EndTime, err = time.Parse(time.RFC3339, configv2.GetEnv("END_TIME", DefaultEndTime))
	if err != nil {
		return nil, fmt.Errorf("parse end time: %w", err)
	}

	debug := configv2.GetEnv("DEBUG", "false")
	if debug == "true" || debug == "1" {
		c.Debug = true
	} else {
		c.Debug = DefaultDebug
	}
	return &c, nil
}
