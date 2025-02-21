package tasks

import (
	"context"
	"github.com/hibiken/asynq"
)

func (h *Handler) HandleSyncStorageCredentialsTask(ctx context.Context, t *asynq.Task) error {
	h.app.Logger.Sugar.Error("示例任务：同步存储信息")

	return nil
}
