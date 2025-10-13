package ingest

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetScanRange(t *testing.T) {
	cases := map[string]struct {
		start  int64
		end    int64
		errStr string
	}{
		"Happy path": {
			start: 0,
			end:   10,
		},
		"Seek error": {
			start: 10000000,
			end:   10,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := os.WriteFile("decodes.csv", []byte("short_url\nbit.ly/1234\nbit.ly/5678\nbit.ly/91011\n"), 0644)
			assert.Nil(t, err)
			defer os.Remove("decodes.csv")
			d, err := NewDecodesData("decodes.csv")
			assert.Nil(t, err)
			err = d.SetScanRange(tc.start, tc.end)
			if tc.errStr != "" {
				assert.True(t, strings.Contains(err.Error(), tc.errStr), fmt.Sprintf("expected string: %s got: %s", tc.errStr, err.Error()))
				return
			}
			assert.Nil(t, err)
		})
	}
}

func TestScan(t *testing.T) {

	cases := map[string]struct {
		start    int64
		end      int64
		expected string
		errStr   string
		contents string
	}{
		"From Beginning": {
			start:    0,
			end:      10,
			expected: "short_url",
			contents: "short_url\nbit.ly/1234\nbit.ly/5678\nbit.ly/91011\n",
		},
		"Mid token": {
			start:    5,
			end:      15,
			expected: "_url",
			contents: "short_url\nbit.ly/1234\nbit.ly/5678\nbit.ly/91011\n",
		},
		"second token": {
			start:    10,
			end:      50,
			expected: "bit.ly/1234",
			contents: "short_url\nbit.ly/1234\nbit.ly/5678\nbit.ly/91011\n",
		},
		"EOF": {
			start:    100,
			end:      500,
			contents: "short_url\nbit.ly/1234\nbit.ly/5678\nbit.ly/91011\n",
			errStr:   "EOF",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := os.WriteFile("/tmp/decodes", []byte(tc.contents), 0644)
			assert.Nil(t, err)
			defer os.Remove("/tmp/decodes")
			d, err := NewDecodesData("/tmp/decodes")
			assert.Nil(t, err)
			err = d.SetScanRange(tc.start, tc.end)
			actual, err := d.Scan()
			if tc.errStr != "" {
				assert.NotNil(t, err)
				if err != nil {
					assert.Equal(t, tc.errStr, err.Error())
				}
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, actual)

		})
	}
}

func TestMultiScan(t *testing.T) {

}
