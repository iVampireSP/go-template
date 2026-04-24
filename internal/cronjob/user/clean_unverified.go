package user

import (
	"context"
	"time"

	"github.com/iVampireSP/go-template/internal/infra/config"
	usersvc "github.com/iVampireSP/go-template/internal/service/identity/user"
	"github.com/iVampireSP/go-template/pkg/foundation/schedule"
	"github.com/iVampireSP/go-template/pkg/logger"
)

// CleanUnverified 清理注册超过指定小时数未验证邮箱的用户。
type CleanUnverified struct {
	svc *usersvc.User
}

// NewCleanUnverified 创建定时任务。
func NewCleanUnverified(svc *usersvc.User) *CleanUnverified {
	return &CleanUnverified{svc: svc}
}

func (c *CleanUnverified) Name() string {
	return "user:clean-unverified"
}

func (c *CleanUnverified) Schedule(e *schedule.Event) *schedule.Event {
	return e.Hourly().
		OnOneServer().
		WithDescription("Clean up users who registered over 24 hours ago without verifying email")
}

func (c *CleanUnverified) Run(ctx context.Context) error {
	hours := config.Int("user.unverified_cleanup_hours", 24)
	olderThan := time.Duration(hours) * time.Hour

	affected, err := c.svc.CleanupUnverified(ctx, olderThan)
	if err != nil {
		return err
	}

	logger.Info("cleaned unverified users", "affected", affected, "older_than_hours", hours)
	return nil
}
