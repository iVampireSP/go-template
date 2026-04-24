package cron

import (
	foundationcron "github.com/iVampireSP/go-template/pkg/foundation/cron"
	"github.com/robfig/cron/v3"
)

func NewCron() *cron.Cron {
	return foundationcron.New()
}
