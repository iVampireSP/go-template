package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go-template/internal/infra"
)

// 定义任务类型常量
const (
	// TypeSyncWorkspace 同步工作区任务
	TypeSyncWorkspace = "workspace:sync"
	// TypeSyncStorageCredentials 同步 S3 服务器列表任务
	TypeSyncStorageCredentials = "storage:credentials:sync"
)

// Handler 任务处理器
type Handler struct {
	app *infra.Application
}

// NewHandler 创建任务处理器
func NewHandler(app *infra.Application) *Handler {
	return &Handler{app: app}
}

// DebugTasks 存储所有需要调试的任务
var DebugTasks = map[string]func(ctx context.Context, t *asynq.Task) error{}

// RegisterDebugTask 注册调试任务
func (h *Handler) RegisterDebugTask(taskType string, handler func(ctx context.Context, t *asynq.Task) error) {
	DebugTasks[taskType] = handler
}

// RunDebugTasks 在调试模式下立即执行所有计划任务
func (h *Handler) RunDebugTasks() {
	h.app.Logger.Sugar.Info("调试模式：立即执行所有计划任务")

	// 注册所有调试任务
	h.RegisterDebugTask(TypeSyncStorageCredentials, h.HandleSyncStorageCredentialsTask)

	// 执行所有调试任务
	for taskType, handler := range DebugTasks {
		h.app.Logger.Sugar.Infof("执行任务: %s", taskType)
		if err := handler(context.Background(), asynq.NewTask(taskType, nil)); err != nil {
			h.app.Logger.Sugar.Errorf("执行任务失败 [%s]: %v", taskType, err)
		} else {
			h.app.Logger.Sugar.Infof("任务执行成功: %s", taskType)
		}
	}
}

// NewScheduler 创建任务调度器
func NewScheduler(app *infra.Application) (*asynq.Scheduler, error) {
	redisOpt := asynq.RedisClientOpt{
		Addr:     app.Config.Redis.Host + ":" + fmt.Sprint(app.Config.Redis.Port),
		Password: app.Config.Redis.Password,
		DB:       app.Config.Redis.DB,
	}

	scheduler := asynq.NewScheduler(
		redisOpt,
		&asynq.SchedulerOpts{},
	)

	// 注册定时任务
	if _, err := scheduler.Register("*/1 * * * *", asynq.NewTask(TypeSyncWorkspace, nil)); err != nil {
		return nil, fmt.Errorf("注册同步工作区任务失败: %v", err)
	}

	return scheduler, nil
}

// NewServer 创建任务服务器
func NewServer(app *infra.Application) *asynq.Server {
	redisOpt := asynq.RedisClientOpt{
		Addr:     app.Config.Redis.Host + ":" + fmt.Sprint(app.Config.Redis.Port),
		Password: app.Config.Redis.Password,
		DB:       app.Config.Redis.DB,
	}

	return asynq.NewServer(
		redisOpt,
		asynq.Config{
			// 指定并发工作器数量
			Concurrency: 10,
			// 任务重试配置
			Queues: map[string]int{
				"default": 1,
			},
			RetryDelayFunc: func(n int, err error, t *asynq.Task) time.Duration {
				// 重试延迟策略：1min, 5min, 15min, 30min, 1h, 2h, 6h, 12h, 24h
				switch n {
				case 1:
					return 1 * time.Minute
				case 2:
					return 5 * time.Minute
				case 3:
					return 15 * time.Minute
				case 4:
					return 30 * time.Minute
				case 5:
					return 1 * time.Hour
				case 6:
					return 2 * time.Hour
				case 7:
					return 6 * time.Hour
				case 8:
					return 12 * time.Hour
				default:
					return 24 * time.Hour
				}
			},
		},
	)
}

// RegisterHandlers 注册所有任务处理器
func RegisterHandlers(mux *asynq.ServeMux, handler *Handler) {
	mux.HandleFunc(TypeSyncStorageCredentials, handler.HandleSyncStorageCredentialsTask)
}
