package ingest

import (
	"encoding/json"
	"fmt"
	"os"
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

type ingest struct {
	cfg     config.Config
	log     *logrus.Logger
	encoder encoder
	decoder decoder
	enc     []encodes
	result  map[string]int
}

type offsetRange struct {
	start int64
	end   int64
}

func New(cfg config.Config, log *logrus.Logger) (*ingest, error) {
	i := ingest{
		cfg:     cfg,
		encoder: &EncodesData{},
		decoder: &DecodesData{},
	}
	if cfg.Debug {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
	i.log = log
	i.result = make(map[string]int)
	return &i, nil
}

func (in *ingest) Run() error {
	var err error
	in.enc, err = in.encoder.Read(in.cfg.EncodingFilePath)
	if err != nil {
		return err
	}

	offsetChan := make(chan offsetRange, in.cfg.Threads)
	resultChan := make(chan map[string]int)
	var wg sync.WaitGroup
	var agWg sync.WaitGroup

	agWg.Go(func(){
		in.agrigator(resultChan)
	})

	// start the worker threads
	for i := range in.cfg.Threads {
		d, err := NewDecodesData(in.cfg.DecodesFilePath)
		if err != nil {
			return err
		}
		defer d.Close()
		wg.Go(func() {
			in.runThread(i, d, offsetChan, resultChan)
		})
	}

	// send the range of
	s, err := os.Stat(in.cfg.DecodesFilePath)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	size := s.Size()
	var offset int64
	chunk := int64(in.cfg.ChunkSizeInKB * 1024)
	for offset = 0; offset < size; offset += chunk {
		offsetChan <- offsetRange{start: offset, end: offset + chunk}
	}

	close(offsetChan)
	wg.Wait()
	close(resultChan)
	agWg.Wait()

	return nil
}
func (in *ingest) PrintResults() {
	for url, count := range in.result {
		fmt.Printf("\n\nurl: %s count: %d\n", url, count)
	}
}
func (in *ingest) agrigator(recv <-chan map[string]int) {
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
func (in *ingest) runThread(threadId int, decode decoder, recv <-chan offsetRange, send chan<- map[string]int) {
	log := in.log.WithField("thread_id", threadId)
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
			if t.Before(in.cfg.StartTime) || t.After(in.cfg.EndTime) {
				log.Debugf("disregarding %v", t.String())
				continue
			}

			// read the hash from the end of the bitlink
			tokens := strings.Split(df.Bitlink, "/")
			hash := tokens[len(tokens)-1]
			url, ok := hash2url[hash]
			if !ok {
				log.Debugf("ignoring hash %s", hash)
				continue
			}
			result[url] += 1

		}
		send <- result

	}

}
