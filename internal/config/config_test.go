package config

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	defaultThreads, _ := strconv.Atoi(DefaultThreads)
	defaultChunk, _ := strconv.Atoi(DefaultChunkSizeInKB)
	start, _ := time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")
	end, _ := time.Parse(time.RFC3339, "2021-12-31T23:59:59Z")
	cases := map[string]struct {
		env      map[string]string
		expected *Config
		errMsg   string
	}{
		"Sunny day": {

			env: map[string]string{
				"THREADS":            "8",
				"CHUNK_SIZE_IN_KB":   "500",
				"ENCODING_FILE_PATH": "data/my_encodes.txt",
				"DECODES_FILE_PATH":  "data/my_decodes.json",
				"DEBUG":              "true",
				"START_TIME":         "2021-01-01T00:00:00Z",
				"END_TIME":           "2021-12-31T23:59:59Z",
			},
			expected: &Config{
				Threads:          8,
				ChunkSizeInKB:    500,
				EncodingFilePath: "data/my_encodes.txt",
				DecodesFilePath:  "data/my_decodes.json",
				Debug:            true,
				EndTime:          end,
				StartTime:        start,
			},
		},
		"Defaults": {

			env: map[string]string{},
			expected: &Config{
				Threads:          defaultThreads,
				ChunkSizeInKB:    defaultChunk,
				EncodingFilePath: DefaultEncodingFilePath,
				DecodesFilePath:  DefaultDecodesFilePath,
				Debug:            DefaultDebug,
				EndTime:          end,
				StartTime:        start,
			},
		},
		"threads err": {

			env: map[string]string{
				"THREADS":            "foo",
				"CHUNK_SIZE_IN_KB":   "500",
				"ENCODING_FILE_PATH": "data/my_encodes.txt",
				"DECODES_FILE_PATH":  "data/my_decodes.json",
				"DEBUG":              "true",
			},
			errMsg: `strconv.Atoi: parsing "foo": invalid syntax`,
		},
		"chunk err": {

			env: map[string]string{
				"CHUNK_SIZE_IN_KB":   "bar",
				"ENCODING_FILE_PATH": "data/my_encodes.txt",
				"DECODES_FILE_PATH":  "data/my_decodes.json",
				"DEBUG":              "true",
			},
			errMsg: `strconv.Atoi: parsing "bar": invalid syntax`,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			cfg, err := New()
			if tc.errMsg != "" {
				assert.Equal(t, tc.errMsg, err.Error())
				assert.Nil(t, cfg)
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, cfg)
			assert.Equal(t, tc.expected, cfg)

		})
	}
}
