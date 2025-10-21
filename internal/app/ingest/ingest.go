package ingest

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stewyb314/bitly/internal/config"
)

type encoder interface {
	Read(string) ([]encodes, error)
}

type decoder interface {
	SetScanRange(int64, int64) error
	Scan() (string, error)
	Close()
}

type decodeForamt struct {
	Bitlink   string `json:"bitlink"`
	Timestamp string `json:"timestamp"`
}

type Ingest struct {
	Cfg     config.Config
	Log     *logrus.Logger
	encoder encoder
	decoder decoder
	enc     []encodes
	result  map[string]int
}

type offsetRange struct {
	start int64
	end   int64
}

// New creates a new ingest struct with private members initilized
func New(cfg config.Config, log *logrus.Logger) (*Ingest, error) {
	i := Ingest{
		Cfg:     cfg,
		encoder: &EncodesData{},
		decoder: &decodesData{},
	}
	if cfg.Debug {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
	i.Log = log
	i.result = make(map[string]int)
	return &i, nil
}

// Run starts and manages go routines that do the work of parsing the decodes file
func (in *Ingest) Run() error {
	var err error
	in.enc, err = in.encoder.Read(in.Cfg.EncodingFilePath)
	if err != nil {
		return err
	}

	// Run() will use this channel to send each go routine the next range of file offsets to read
	offsetChan := make(chan offsetRange, in.Cfg.Threads)
	// as each go routine getnerates a map of url->count, it passes the map to resultChan.
	// agrigator() reads from resultChan and agrigates the results into a single map
	resultChan := make(chan map[string]int, in.Cfg.Threads)
	var wg sync.WaitGroup
	var agWg sync.WaitGroup

	agWg.Go(func() {
		in.agrigator(resultChan)
	})

	// start each thread
	for i := range in.Cfg.Threads {
		d, err := NewDecodesData(in.Cfg.DecodesFilePath)
		if err != nil {
			return err
		}
		defer d.Close()
		wg.Go(func() {
			in.runThread(i, d, offsetChan, resultChan)
		})
	}

	// get the size of the file and start sending start, end offsets for each go routine to read
	s, err := os.Stat(in.Cfg.DecodesFilePath)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	size := s.Size()
	var offset int64
	chunk := int64(in.Cfg.ChunkSizeInKB * 1024)
	for offset = 0; offset < size; offset += chunk {
		offsetChan <- offsetRange{start: offset, end: offset + chunk}
	}

	// we've sent all if the offsets, so close the channel and wait for the go routines to finish
	close(offsetChan)
	wg.Wait()
	// done processing the data, close the resultChan to indicate to agrigator() that we're done then wait for agriagtor to finish
	close(resultChan)
	agWg.Wait()

	return nil
}

// PrintResults sorts the in.results by count and prints the results
func (in *Ingest) PrintResults() string {
	type keyValuePair struct {
		key   string
		value int
	}
	values := make([]keyValuePair, 0)
	for k, v := range in.result {
		values = append(values, keyValuePair{key: k, value: v})
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].value > values[j].value
	})
	return fmt.Sprintf("%+v", values)
}

// agrigator recieves map of url->count when each runThread() finishes reading a chunk of the file,
// it then merges each of the maps into the final result
func (in *Ingest) agrigator(recv <-chan map[string]int) {
	for data := range recv {
		for url, count := range data {
			url = strings.ToLower(url)
			if _, ok := in.result[url]; !ok {
				in.result[url] = 0
			}
			in.result[url] += count
		}
	}
}

// runThread:
// 1) receives a range of offsets from the recv channel
// 2) reads data from the file one line at a time until it reaches the end offset
// 3) unmarshals each line
// 4) if the line is in the correct date range and a hash from the encoding.csv file, add it to the results map
// 5) when EOF is reached (indicating the end offset has been reached), send the result map to agrigator() via the send channel
// 6) loop and read the next offset
func (in *Ingest) runThread(threadId int, decode decoder, recv <-chan offsetRange, send chan<- map[string]int) {
	log := in.Log.WithField("thread_id", threadId)
	log.Debug("starting...")
	hash2url := make(map[string]string)
	// build a quick lookup table to match hashes to the url
	for _, e := range in.enc {
		hash2url[e.hash] = e.url
	}

	// read offsets from the recv channel until the channel closes or we hit an EOF
	for offset := range recv {
		log.Debugf("scanning range %+v", offset)
		result := make(map[string]int)
		decode.SetScanRange(offset.start, offset.end)
		for {
			line, err := decode.Scan()
			if err != nil {
				if err.Error() == "EOF" {
					log.Debug("reached EOF")
				} else {
					log.WithError(err).Error("scanning line")
				}
				break
			}
			var df decodeForamt
			if err := json.Unmarshal([]byte(line), &df); err != nil {
				log.WithError(err).Debugf("unmarshal failed: %s", line)
				continue
			}

			t, err := time.Parse(time.RFC3339, df.Timestamp)
			if err != nil {
				log.WithError(err).Error("parsing timestamp")
				continue
			}
			if t.Before(in.Cfg.StartTime) || t.After(in.Cfg.EndTime) {
				continue
			}

			// read the hash from the end of the bitlink
			tokens := strings.Split(df.Bitlink, "/")
			hash := tokens[len(tokens)-1]
			url, ok := hash2url[hash]
			if !ok {
				continue
			}
			result[url] += 1

		}
		send <- result

	}

}
