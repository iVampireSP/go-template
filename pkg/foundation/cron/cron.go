package cron

import (
	"github.com/robfig/cron/v3"
)

func New() *cron.Cron {
	return cron.New(cron.WithSeconds())
}
