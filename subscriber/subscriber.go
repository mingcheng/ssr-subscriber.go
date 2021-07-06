package subscriber

import (
	"context"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
)

type Subscriber struct {
	Fetcher *Fetcher
	Sources []string

	ticker *time.Ticker

	Interval       time.Duration
	FetchTimestamp time.Time
}

// Start to restore configs then fetch and update save
func (s *Subscriber) Start() error {
	if err := s.Fetcher.Restore(context.Background()); err != nil {
		log.Warn("restore config is failed, maybe redis is not configured")
	}

	if s.Interval > 0 {
		s.ticker = time.NewTicker(s.Interval)

		go func() {
			for {
				if err := s.Fetch(); err != nil {
					log.Error(err)
				}

				if err := s.Fetcher.Check(); err != nil {
					log.Error(err)
				}

				if err := s.Fetcher.Save(context.Background()); err != nil {
					log.Error(err)
				}

				<-s.ticker.C
			}
		}()
	}

	return nil
}

// Stop and save configs
func (s *Subscriber) Stop(ctx context.Context) error {
	if s.ticker != nil {
		s.ticker.Stop()
	}

	return s.Fetcher.Save(ctx)
}

// Fetch to get ssr configure from file or via url
func (s *Subscriber) Fetch() error {
	var err error
	for _, uri := range s.Sources {
		log.Infof("fetch from %s", uri)

		if path.IsAbs(uri) {
			err = s.Fetcher.FromFile(uri)
		} else {
			err = s.Fetcher.FromURL(uri)
		}

		log.Infof("last fetch timestamp is %v", s.FetchTimestamp)
		s.FetchTimestamp = time.Now()
	}

	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
