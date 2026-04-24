package queue

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/iVampireSP/go-template/pkg/json"
	"github.com/iVampireSP/go-template/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Job 状态常量
const (
	JobStatusPending    = "pending"
	JobStatusProcessing = "processing"
	JobStatusSucceeded  = "succeeded"
	JobStatusFailed     = "failed"
	JobStatusRetrying   = "retrying"
	JobStatusDLQ        = "dlq"
)

// Redis key TTL
const (
	succeededTTL = 1 * time.Hour
	failedTTL    = 24 * time.Hour
	retryingTTL  = 24 * time.Hour
	dlqTTL       = 7 * 24 * time.Hour
	workerTTL    = 30 * time.Second
	heartbeatInt = 10 * time.Second
)

// tracker 基于 Redis 追踪 Job 生命周期
type tracker struct {
	redis    redis.UniversalClient
	workerID string
}

// newTracker 创建 tracker
func newTracker(r redis.UniversalClient) *tracker {
	hostname, _ := os.Hostname()
	wid := fmt.Sprintf("%s-%d-%s", hostname, os.Getpid(), uuid.New().String()[:8])
	return &tracker{
		redis:    r,
		workerID: wid,
	}
}

// recordPending 记录 Job 进入 pending 状态
func (t *tracker) recordPending(ctx context.Context, env *Envelope, topic string) {
	now := time.Now()
	key := jobKey(env.ID)

	payloadStr := string(env.Payload)

	pipe := t.redis.Pipeline()
	pipe.HSet(ctx, key, map[string]any{
		"id":         env.ID,
		"name":       env.Name,
		"key":        env.Key,
		"status":     JobStatusPending,
		"topic":      topic,
		"payload":    payloadStr,
		"attempt":    env.Attempt,
		"max_retry":  env.MaxRetry,
		"created_at": now.Unix(),
	})
	pipe.Expire(ctx, key, failedTTL)

	pipe.ZAdd(ctx, statusKey(JobStatusPending), redis.Z{Score: float64(now.Unix()), Member: env.ID})
	pipe.ZAdd(ctx, allJobsKey(), redis.Z{Score: float64(now.Unix()), Member: env.ID})

	if _, err := pipe.Exec(ctx); err != nil {
		logger.Debug("tracker: failed to record pending", zap.Error(err))
	}
}

// recordProcessing 记录 Job 开始处理
func (t *tracker) recordProcessing(ctx context.Context, env *Envelope) {
	now := time.Now()
	key := jobKey(env.ID)

	pipe := t.redis.Pipeline()
	pipe.HSet(ctx, key, map[string]any{
		"status":     JobStatusProcessing,
		"worker_id":  t.workerID,
		"started_at": now.Unix(),
		"attempt":    env.Attempt,
	})
	pipe.ZRem(ctx, statusKey(JobStatusPending), env.ID)
	pipe.ZAdd(ctx, statusKey(JobStatusProcessing), redis.Z{Score: float64(now.Unix()), Member: env.ID})
	pipe.ZAdd(ctx, allJobsKey(), redis.Z{Score: float64(now.Unix()), Member: env.ID})

	if _, err := pipe.Exec(ctx); err != nil {
		logger.Debug("tracker: failed to record processing", zap.Error(err))
	}
}

// recordSucceeded 记录 Job 成功
func (t *tracker) recordSucceeded(ctx context.Context, env *Envelope) {
	now := time.Now()
	key := jobKey(env.ID)

	pipe := t.redis.Pipeline()
	pipe.HSet(ctx, key, map[string]any{
		"status":       JobStatusSucceeded,
		"completed_at": now.Unix(),
	})
	pipe.Expire(ctx, key, succeededTTL)

	pipe.ZRem(ctx, statusKey(JobStatusProcessing), env.ID)
	pipe.ZAdd(ctx, statusKey(JobStatusSucceeded), redis.Z{Score: float64(now.Unix()), Member: env.ID})
	pipe.ZAdd(ctx, allJobsKey(), redis.Z{Score: float64(now.Unix()), Member: env.ID})
	pipe.Publish(ctx, doneChannel(env.ID), JobStatusSucceeded)

	if _, err := pipe.Exec(ctx); err != nil {
		logger.Debug("tracker: failed to record succeeded", zap.Error(err))
	}
}

// recordRetrying 记录 Job 重试中
func (t *tracker) recordRetrying(ctx context.Context, env *Envelope) {
	now := time.Now()
	key := jobKey(env.ID)

	pipe := t.redis.Pipeline()
	pipe.HSet(ctx, key, map[string]any{
		"status":  JobStatusRetrying,
		"attempt": env.Attempt,
	})
	pipe.Expire(ctx, key, retryingTTL)

	pipe.ZRem(ctx, statusKey(JobStatusProcessing), env.ID)
	pipe.ZAdd(ctx, statusKey(JobStatusRetrying), redis.Z{Score: float64(now.Unix()), Member: env.ID})
	pipe.ZAdd(ctx, allJobsKey(), redis.Z{Score: float64(now.Unix()), Member: env.ID})

	if _, err := pipe.Exec(ctx); err != nil {
		logger.Debug("tracker: failed to record retrying", zap.Error(err))
	}
}

// recordDLQ 记录 Job 进入死信队列
func (t *tracker) recordDLQ(ctx context.Context, env *Envelope, jobErr error) {
	now := time.Now()
	key := jobKey(env.ID)

	errMsg := ""
	if jobErr != nil {
		errMsg = jobErr.Error()
	}

	pipe := t.redis.Pipeline()
	pipe.HSet(ctx, key, map[string]any{
		"status":       JobStatusDLQ,
		"error":        errMsg,
		"completed_at": now.Unix(),
	})
	pipe.Expire(ctx, key, dlqTTL)

	pipe.ZRem(ctx, statusKey(JobStatusProcessing), env.ID)
	pipe.ZRem(ctx, statusKey(JobStatusRetrying), env.ID)
	pipe.ZAdd(ctx, statusKey(JobStatusDLQ), redis.Z{Score: float64(now.Unix()), Member: env.ID})
	pipe.ZAdd(ctx, allJobsKey(), redis.Z{Score: float64(now.Unix()), Member: env.ID})
	pipe.Publish(ctx, doneChannel(env.ID), JobStatusDLQ)

	if _, err := pipe.Exec(ctx); err != nil {
		logger.Debug("tracker: failed to record dlq", zap.Error(err))
	}
}

// startHeartbeat 启动 Worker 心跳
func (t *tracker) startHeartbeat(ctx context.Context) {
	hostname, _ := os.Hostname()
	workerKey := workerInfoKey(t.workerID)

	pipe := t.redis.Pipeline()
	pipe.HSet(ctx, workerKey, map[string]any{
		"id":             t.workerID,
		"hostname":       hostname,
		"pid":            os.Getpid(),
		"started_at":     time.Now().Unix(),
		"last_heartbeat": time.Now().Unix(),
	})
	pipe.Expire(ctx, workerKey, workerTTL)
	pipe.ZAdd(ctx, workersKey(), redis.Z{Score: float64(time.Now().Unix()), Member: t.workerID})
	if _, err := pipe.Exec(ctx); err != nil {
		logger.Warn("tracker: failed to register worker", zap.Error(err))
	}

	go func() {
		ticker := time.NewTicker(heartbeatInt)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				bg := context.Background()
				t.redis.Del(bg, workerKey)
				t.redis.ZRem(bg, workersKey(), t.workerID)
				return
			case <-ticker.C:
				now := time.Now().Unix()
				pipe := t.redis.Pipeline()
				pipe.HSet(ctx, workerKey, "last_heartbeat", now)
				pipe.Expire(ctx, workerKey, workerTTL)
				pipe.ZAdd(ctx, workersKey(), redis.Z{Score: float64(now), Member: t.workerID})
				if _, err := pipe.Exec(ctx); err != nil {
					logger.Debug("tracker: heartbeat failed", zap.Error(err))
				}
			}
		}
	}()
}

// setCurrentJob 设置当前正在处理的 Job
func (t *tracker) setCurrentJob(ctx context.Context, envID string) {
	key := workerInfoKey(t.workerID)
	t.redis.HSet(ctx, key, "current_job", envID)
}

// clearCurrentJob 清除当前 Job
func (t *tracker) clearCurrentJob(ctx context.Context) {
	key := workerInfoKey(t.workerID)
	t.redis.HSet(ctx, key, "current_job", "")
}

// ---- 查询方法（供 Queue service 使用） ----

// JobInfo 查询返回的 Job 信息
type JobInfo struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Key         string          `json:"key"`
	Status      string          `json:"status"`
	Topic       string          `json:"topic"`
	Payload     json.RawMessage `json:"payload"`
	Attempt     int             `json:"attempt"`
	MaxRetry    int             `json:"max_retry"`
	Error       string          `json:"error,omitempty"`
	WorkerID    string          `json:"worker_id,omitempty"`
	CreatedAt   *time.Time      `json:"created_at,omitempty"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

// WorkerInfo 查询返回的 Worker 信息
type WorkerInfo struct {
	ID            string    `json:"id"`
	Hostname      string    `json:"hostname"`
	PID           int       `json:"pid"`
	StartedAt     time.Time `json:"started_at"`
	CurrentJob    string    `json:"current_job,omitempty"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

// Stats 队列统计
type Stats struct {
	Pending    int64 `json:"pending"`
	Processing int64 `json:"processing"`
	Succeeded  int64 `json:"succeeded"`
	Failed     int64 `json:"failed"`
	Retrying   int64 `json:"retrying"`
	DLQ        int64 `json:"dlq"`
}

// Jobs 查询 Job 列表，status 为空时查询所有状态
func (t *tracker) Jobs(ctx context.Context, status, name string, offset, limit int) ([]JobInfo, int64, error) {
	sKey := statusKey(status)
	if status == "" {
		sKey = allJobsKey()
	}

	total, err := t.redis.ZCard(ctx, sKey).Result()
	if err != nil {
		return nil, 0, err
	}

	ids, err := t.redis.ZRevRange(ctx, sKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, 0, err
	}

	jobs := make([]JobInfo, 0, len(ids))
	for _, id := range ids {
		info, err := t.jobInfo(ctx, id)
		if err != nil || info == nil {
			if status == "" {
				t.redis.ZRem(ctx, allJobsKey(), id)
			}
			continue
		}
		if name != "" && info.Name != name {
			continue
		}
		jobs = append(jobs, *info)
	}

	return jobs, total, nil
}

// Job 查询单个 Job
func (t *tracker) Job(ctx context.Context, id string) (*JobInfo, error) {
	return t.jobInfo(ctx, id)
}

// Workers 查询活跃 Workers
func (t *tracker) Workers(ctx context.Context) ([]WorkerInfo, error) {
	ids, err := t.redis.ZRevRange(ctx, workersKey(), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	workers := make([]WorkerInfo, 0, len(ids))
	for _, wid := range ids {
		result, err := t.redis.HGetAll(ctx, workerInfoKey(wid)).Result()
		if err != nil || len(result) == 0 {
			continue
		}
		w := WorkerInfo{
			ID:       result["id"],
			Hostname: result["hostname"],
		}
		if pid, err := strconv.Atoi(result["pid"]); err == nil {
			w.PID = pid
		}
		if ts, err := strconv.ParseInt(result["started_at"], 10, 64); err == nil {
			w.StartedAt = time.Unix(ts, 0)
		}
		w.CurrentJob = result["current_job"]
		if ts, err := strconv.ParseInt(result["last_heartbeat"], 10, 64); err == nil {
			w.LastHeartbeat = time.Unix(ts, 0)
		}
		workers = append(workers, w)
	}

	return workers, nil
}

// StatsAll 获取各状态计数
func (t *tracker) StatsAll(ctx context.Context) (*Stats, error) {
	statuses := []string{JobStatusPending, JobStatusProcessing, JobStatusSucceeded, JobStatusFailed, JobStatusRetrying, JobStatusDLQ}
	pipe := t.redis.Pipeline()
	cmds := make([]*redis.IntCmd, len(statuses))
	for i, s := range statuses {
		cmds[i] = pipe.ZCard(ctx, statusKey(s))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, err
	}

	return &Stats{
		Pending:    cmds[0].Val(),
		Processing: cmds[1].Val(),
		Succeeded:  cmds[2].Val(),
		Failed:     cmds[3].Val(),
		Retrying:   cmds[4].Val(),
		DLQ:        cmds[5].Val(),
	}, nil
}

// jobInfo 从 Redis 读取单个 Job 信息
func (t *tracker) jobInfo(ctx context.Context, id string) (*JobInfo, error) {
	result, err := t.redis.HGetAll(ctx, jobKey(id)).Result()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}

	info := &JobInfo{
		ID:       result["id"],
		Name:     result["name"],
		Key:      result["key"],
		Status:   result["status"],
		Topic:    result["topic"],
		Error:    result["error"],
		WorkerID: result["worker_id"],
	}

	if result["payload"] != "" {
		info.Payload = json.RawMessage(result["payload"])
	}
	if v, err := strconv.Atoi(result["attempt"]); err == nil {
		info.Attempt = v
	}
	if v, err := strconv.Atoi(result["max_retry"]); err == nil {
		info.MaxRetry = v
	}
	if ts, err := strconv.ParseInt(result["created_at"], 10, 64); err == nil {
		t := time.Unix(ts, 0)
		info.CreatedAt = &t
	}
	if ts, err := strconv.ParseInt(result["started_at"], 10, 64); err == nil {
		t := time.Unix(ts, 0)
		info.StartedAt = &t
	}
	if ts, err := strconv.ParseInt(result["completed_at"], 10, 64); err == nil {
		t := time.Unix(ts, 0)
		info.CompletedAt = &t
	}

	return info, nil
}

// ---- Redis key 格式 ----

func jobKey(id string) string {
	return fmt.Sprintf("queue:queue:%s", id)
}
func statusKey(status string) string {
	return fmt.Sprintf("queue:status:%s", status)
}
func allJobsKey() string {
	return "queue:jobs"
}
func workersKey() string {
	return "queue:workers"
}
func workerInfoKey(wid string) string {
	return fmt.Sprintf("queue:worker:%s", wid)
}
func doneChannel(id string) string {
	return fmt.Sprintf("queue:done:%s", id)
}
