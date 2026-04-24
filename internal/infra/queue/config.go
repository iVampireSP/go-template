package queue

import (
	"time"

	infraconfig "github.com/iVampireSP/go-template/internal/infra/config"
)

type config struct {
	MaxRetry    int
	RetryDelays []time.Duration
}

func loadConfig() *config {
	cfg := &config{
		MaxRetry: 5,
		RetryDelays: []time.Duration{
			1 * time.Second,
			2 * time.Second,
			4 * time.Second,
			8 * time.Second,
			16 * time.Second,
		},
	}

	if maxRetry := infraconfig.Int("job.defaults.max_retry"); maxRetry > 0 {
		cfg.MaxRetry = maxRetry
	}
	if delays := infraconfig.Strings("job.defaults.retry_delays"); len(delays) > 0 {
		cfg.RetryDelays = make([]time.Duration, 0, len(delays))
		for _, d := range delays {
			if duration, err := time.ParseDuration(d); err == nil {
				cfg.RetryDelays = append(cfg.RetryDelays, duration)
			}
		}
	}

	return cfg
}
