package ingest

import (
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stewyb314/bitly/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestAgrigator(t *testing.T) {
	cases := map[string]struct {
		input    []map[string]int
		expected map[string]int
	}{
		"Happy path": {
			input: []map[string]int{
				{
					"foo": 2,
					"Bar": 3,
					"baz": 1,
				},
				{
					"fOo": 1,
					"baz": 1,
				},
				{
					"new": 1,
					"baz": 1,
				},
			},
			expected: map[string]int{
				"foo": 3,
				"bar": 3,
				"new": 1,
				"baz": 3,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			log, _ := test.NewNullLogger()
			c, _ := config.New()
			in, err := New(*c, log)
			assert.Nil(t, err)
			resultChan := make(chan map[string]int)
			go in.agrigator(resultChan)
			for _, input := range tc.input {
				resultChan <- input
			}
			close(resultChan)
			assert.EqualValues(t, tc.expected, in.result)
		})
	}
}

func TestRunThread(t *testing.T) {
	fileContent :=
		`{"bitlink": "http://bit.ly/1234",  "timestamp": "2021-01-03T00:00:00Z"}
		{"bitlink": "http://bit.ly/abcd",  "timestamp": "2021-01-03T00:00:00Z"}
		{"bitlink": "http://bit.ly/def", "timestamp": "2021-01-03T00:00:00Z"}
		{"bitlink": "http://es.pn/1234", "timestamp": "2022-01-03T00:00:00Z"}
		{"bitlink": "http://es.pn/def", "timestamp": "2021-01-03T00:00:00Z"}
		{"bitlink": "http://bit.ly/foo", "timestamp": "2021-01-03T00:00:00Z"}
		{"bitlink": "http://bit.ly/1234","timestamp": "2021-01-03T00:00:00Z"} 
		{"bitlink": "http://bit.ly/bar", "timestamp": "2021-01-03T00:00:00Z"}`
	encContents := `long_url,domain,hash
					1234.com,bit.ly,1234
					abcd.com,bit.ly,abcd
					not.in.decode.com,bit.ly,notindecode
					def.com,bit.ly,def`

	decodesFile := "/tmp/decodes.json"
	endcodesFile := "/tmp/encodes.csv"
	cases := map[string]struct {
		expected map[string]int
		env map[string]string

	}{
		"date limited": {
		env: map[string]string{
			"THREADS":"1",
			"DEBUG": "true",
			"ENCODING_FILE_PATH": endcodesFile,
			"DECODES_FILE_PATH": decodesFile,

		},
		expected: map[string]int {
			"1234.com": 2,
			"abcd.com": 1,
			"def.com":2,
		},

		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T){
			err := os.WriteFile(decodesFile, []byte(fileContent), 0644)
			assert.Nil(t, err)
			defer os.Remove(decodesFile)
			err = os.WriteFile(endcodesFile, []byte(encContents), 0644)
			assert.Nil(t, err)
			defer os.Remove(encContents)
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			//log, _ := test.NewNullLogger()
			log := logrus.New()
			c, err := config.New()
			assert.Nil(t, err)
			in, err := New(*c, log)
			err = in.Run()
			fmt.Printf("enc: %+v\n", in.enc)
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, in.result)




		})
	}
	/*
	t.Run("RunThread", func(t *testing.T) {
		err := os.WriteFile(decodesFile, []byte(fileContent), 0644)
		assert.Nil(t, err)
		defer os.Remove(decodesFile)
		d, err := NewDecodesData(decodesFile)
		log, _ := test.NewNullLogger()
		in, err := New(*c, log)
		in.enc = []encodes{
			{url: "1234.com", hash: "1234", domain: "bit.ly"},
			{url: "abcd.com", hash: "abcd", domain: "bit.ly"},
			{url: "not.in.decode.com", hash: "notindecode", domain: "bit.ly"},
			{url: "def.com", hash: "def", domain: "bit.ly"},
		}
		offsetChan := make(chan offsetRange)
		resultChan := make(chan map[string]int)

		go in.runThread(1, d, offsetChan, resultChan)
		offsetChan <- offsetRange{start: 0, end: 5000}
		close(offsetChan)
		result := <-resultChan
		close(resultChan)
		fmt.Printf("%+v", result)
		assert.Nil(t, 1)

	})
		*/

}
