package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stewyb314/bitly/internal/app/ingest"
	"github.com/stewyb314/bitly/internal/config"
	"golang.org/x/sync/errgroup"
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

	in.Log.Debugf("config: %+v", cfg)
	if cfg.RunDaemon {
		return runDaemon(in)
	} else {
		return runStandalone(in)
	}
}

func runStandalone(in *ingest.Ingest) error {
	err := in.Run()
	if err != nil {
		return err
	}
	in.Log.Info(in.PrintResults())
	return nil
}
func runDaemon(in *ingest.Ingest) error {
	eg, ctx := errgroup.WithContext(context.Background())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})
	server := &http.Server{Addr: fmt.Sprintf(":%d", in.Cfg.Port)}

	eg.Go(func() error {
		return server.ListenAndServe()
	})

	eg.Go(func() error {
		ticker := time.NewTicker(30 * time.Second)
		for {
			select {
			case <-ctx.Done():
				{
					return nil
				}
			case <-ticker.C:
				{
					err := in.Run()
					if err != nil {
						server.Shutdown(ctx)
						return err
					}
					in.Log.Info(in.PrintResults())
				}
			}

		}
	})

	err := eg.Wait()

	return err
}
