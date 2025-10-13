package ingest

import (
	"encoding/csv"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadEncodes(t *testing.T) {
	cases := map[string]struct {
		csv      string
		expected []encodes
		errStr   string
	}{
		"Sunny Day": {
			csv: `long_url,domain,hash
				  https://foo.bar.com,bit.ly,1234
				  https://hire.stew.now,bit.ly,34565
				  https://google.com,bit.ly,987`,
			expected: []encodes{
				{
					url:    "https://foo.bar.com",
					domain: "bit.ly",
					hash:   "1234",
				},
				{
					url:    "https://hire.stew.now",
					domain: "bit.ly",
					hash:   "34565",
				},
				{
					url:    "https://google.com",
					domain: "bit.ly",
					hash:   "987",
				},
			},
		},
		"Extra Field": {
			csv: `long_url,domain,hash
				  https://foo.bar.com,bit.ly,1234
				  https://bing.com,bit.ly,34565,extra
				  https://google.com,bit.ly,987`,
			errStr: "wrong number of fields",
		},
		"Missing Field": {
			csv: `long_url,domain,hash
				  https://foo.bar.com,bit.ly,1234
				  https://bing.com,bit.ly
				  https://google.com,bit.ly,987`,
			errStr: "wrong number of fields",
		},
		"Bad Header": {
			csv: `long_url,foo,hash
				  https://foo.bar.com,bit.ly,1234
				  https://bing.com,bit.ly,34565
				  https://google.com,bit.ly,987`,
			errStr: "csv headers: invalid format",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := csv.NewReader(strings.NewReader(tc.csv))
			actual, err := readEncodes(r)
			if tc.errStr != "" {
				assert.True(t, strings.Contains(err.Error(), tc.errStr), fmt.Sprintf("expected string: %s got: %s", tc.errStr, err.Error()))
				return
			}
			assert.Nil(t, err)
			assert.ElementsMatch(t, tc.expected, actual)
		})
	}
}
