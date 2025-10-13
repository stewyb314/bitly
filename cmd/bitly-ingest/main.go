package main

import (
	"github.com/sirupsen/logrus"
	"github.com/stewyb314/bitly/internal/app/ingest"
	"github.com/stewyb314/bitly/internal/config"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	// Although this is a commandline tool, we're simulating a long running procceess that streams the data rather
	// than reading it from a file.  Normally I would use command line options to tweak the default behavior, but
	// in this case I'm using environment varialbes, as this is more suitable for a long running process in a container.
	cfg, err := config.New()
	if err != nil {
		return err
	}
	in, err := ingest.New(*cfg, logrus.New())
	if err != nil {
		return err
	}
	if err := in.Run(); err != nil {
		return err
	}
	in.PrintResults()
	return nil
}
