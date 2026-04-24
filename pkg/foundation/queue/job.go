package queue

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/iVampireSP/go-template/pkg/cerr"
	"github.com/iVampireSP/go-template/pkg/logger"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
)

var (
	ErrEncodeMessage  = cerr.Internal("failed to encode queue message").WithCode("JOB_ENCODE_MESSAGE")
	ErrEnqueueJob     = cerr.ServiceUnavailable("failed to enqueue queue").WithCode("JOB_ENQUEUE")
	ErrDecodeJob      = cerr.Internal("failed to decode queue payload").WithCode("JOB_DECODE")
	ErrProcessJobTask = cerr.Internal("failed to process queue task").WithCode("JOB_PROCESS")
)

type Queue struct {
	router   *Router
	cfg      Config
	redisCfg RedisConfig
	tracker  *tracker
	client   *asynq.Client
	server   *asynq.Server
	handlers map[string]Handler
	mw       []Middleware
	mu       sync.RWMutex
	running  bool
}

// NewQueue 创建任务队列。redisCfg 用于 asynq 连接，cfg 控制重试策略，redisClient 用于 tracker。
func NewQueue(redisCfg RedisConfig, cfg Config, redisClient redis.UniversalClient) *Queue {
	router := newRouter()
	client := asynq.NewClient(buildAsynqRedisConnOpt(redisCfg))

	var t *tracker
	if redisClient != nil {
		t = newTracker(redisClient)
	}

	return &Queue{
		router:   router,
		cfg:      cfg,
		redisCfg: redisCfg,
		tracker:  t,
		client:   client,
		handlers: make(map[string]Handler),
		mw:       []Middleware{recovery(), logging(), retryInfo()},
	}
}

func (q *Queue) Dispatch(ctx context.Context, job Job, dispatchOptions ...DispatchOption) (string, error) {
	maxRetry, retryDelays := q.resolveRetryConfig(job)

	env, err := newJobEnvelope(job.Name(), job.Key(), job, maxRetry, retryDelays)
	if err != nil {
		return "", ErrEncodeMessage.WithCause(err)
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(env.Metadata))

	queueName := q.resolveQueue(job.Name(), job)
	env.Queue = queueName
	payload, err := env.Encode()
	if err != nil {
		return "", ErrEncodeMessage.WithCause(err)
	}

	retryCount := maxRetry - 1
	if retryCount < 0 {
		retryCount = 0
	}

	cfg := DispatchConfig{}
	for _, opt := range dispatchOptions {
		if opt == nil {
			continue
		}
		opt(&cfg)
	}

	taskOptions := []asynq.Option{
		asynq.Queue(queueName),
		asynq.MaxRetry(retryCount),
	}
	if key := job.Key(); key != "" && key != job.Name() {
		taskOptions = append(taskOptions, asynq.TaskID(key))
	}
	if cfg.ProcessAt != nil {
		taskOptions = append(taskOptions, asynq.ProcessAt(*cfg.ProcessAt))
	} else if cfg.ProcessIn != nil {
		taskOptions = append(taskOptions, asynq.ProcessIn(*cfg.ProcessIn))
	}

	task := asynq.NewTask(taskType(job.Name()), payload, taskOptions...)

	if _, err := q.client.EnqueueContext(ctx, task); err != nil {
		if errors.Is(err, asynq.ErrTaskIDConflict) {
			return env.ID, nil
		}
		return "", ErrEnqueueJob.WithCause(err)
	}

	if q.tracker != nil {
		q.tracker.recordPending(ctx, env, queueName)
	}

	return env.ID, nil
}

// Wait 等待指定 Job 到达终态（succeeded/dlq），返回最终 JobInfo。
func (q *Queue) Wait(ctx context.Context, jobID string) (*JobInfo, error) {
	if q.tracker == nil {
		return nil, errors.New("tracker not available")
	}

	pubsub := q.tracker.redis.Subscribe(ctx, doneChannel(jobID))
	defer pubsub.Close()
	if _, err := pubsub.Receive(ctx); err != nil {
		return nil, fmt.Errorf("subscribe job done channel: %w", err)
	}

	info, err := q.tracker.Job(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if info != nil && isTerminal(info.Status) {
		return info, nil
	}

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case _, ok := <-ch:
			if !ok {
				return nil, errors.New("subscription closed")
			}
			info, err := q.tracker.Job(ctx, jobID)
			if err != nil {
				return nil, err
			}
			if info != nil && isTerminal(info.Status) {
				return info, nil
			}
		}
	}
}

func isTerminal(status string) bool {
	return status == JobStatusSucceeded || status == JobStatusDLQ
}

func (q *Queue) Process(name string, handler Handler) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[name] = handler
}

func (q *Queue) Use(middleware ...Middleware) {
	if len(middleware) == 0 {
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()
	q.mw = append(q.mw, middleware...)
}

func (q *Queue) Run(ctx context.Context) error {
	q.mu.Lock()
	if q.running {
		q.mu.Unlock()
		return nil
	}

	handlers := make(map[string]Handler, len(q.handlers))
	queueSet := map[string]struct{}{}
	for name, handler := range q.handlers {
		handlers[name] = handler
		queueSet[q.router.JobQueue(name)] = struct{}{}
	}
	middleware := append([]Middleware(nil), q.mw...)
	if len(queueSet) == 0 {
		queueSet[q.router.JobQueue("")] = struct{}{}
	}

	queues := make(map[string]int, len(queueSet))
	for queueName := range queueSet {
		queues[queueName] = 1
	}

	q.server = asynq.NewServer(buildAsynqRedisConnOpt(q.redisCfg), asynq.Config{
		Concurrency:    16,
		Queues:         queues,
		RetryDelayFunc: q.retryDelay,
		IsFailure: func(err error) bool {
			return !errors.Is(err, asynq.SkipRetry)
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(errCtx context.Context, task *asynq.Task, err error) {
			q.onTaskError(errCtx, task, err)
		}),
	})
	q.running = true
	server := q.server
	q.mu.Unlock()

	mux := asynq.NewServeMux()
	for name, handler := range handlers {
		jobName := name
		jobHandler := handler
		mux.HandleFunc(taskType(jobName), func(taskCtx context.Context, task *asynq.Task) error {
			return q.handleTask(taskCtx, jobName, jobHandler, task, middleware)
		})
	}

	if err := server.Start(mux); err != nil {
		q.mu.Lock()
		q.running = false
		q.server = nil
		q.mu.Unlock()
		return err
	}

	if q.tracker != nil {
		q.tracker.startHeartbeat(ctx)
	}

	<-ctx.Done()
	server.Shutdown()

	q.mu.Lock()
	q.running = false
	q.server = nil
	q.mu.Unlock()

	return ctx.Err()
}

func (q *Queue) RetryEnvelope(ctx context.Context, env *Envelope) error {
	env.Attempt = 0
	queueName := env.Queue
	if queueName == "" {
		queueName = q.router.JobQueue(env.Name)
		env.Queue = queueName
	}
	taskPayload, err := env.Encode()
	if err != nil {
		return ErrEncodeMessage.WithCause(err)
	}

	retryCount := env.MaxRetry - 1
	if retryCount < 0 {
		retryCount = 0
	}

	task := asynq.NewTask(taskType(env.Name), taskPayload,
		asynq.Queue(queueName),
		asynq.MaxRetry(retryCount),
	)
	if _, err := q.client.EnqueueContext(ctx, task); err != nil {
		return ErrEnqueueJob.WithCause(err)
	}
	if q.tracker != nil {
		q.tracker.recordPending(ctx, env, queueName)
	}
	return nil
}

func (q *Queue) Tracker() *tracker {
	return q.tracker
}

func (q *Queue) handleTask(ctx context.Context, name string, handler Handler, task *asynq.Task, middleware []Middleware) error {
	env, err := decodeEnvelope(task.Payload())
	if err != nil {
		return ErrDecodeJob.WithCause(err)
	}

	if retryCount, ok := asynq.GetRetryCount(ctx); ok {
		env.Attempt = retryCount
	}

	if len(env.Metadata) > 0 {
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(env.Metadata))
	}

	ctx = ContextWithEnvelope(ctx, env)
	if q.tracker != nil {
		q.tracker.recordProcessing(ctx, env)
		q.tracker.setCurrentJob(ctx, env.ID)
		defer q.tracker.clearCurrentJob(ctx)
	}

	run := handler
	for idx := len(middleware) - 1; idx >= 0; idx-- {
		run = middleware[idx](run)
	}

	err = run(ctx, env.Payload)
	if err == nil {
		if q.tracker != nil {
			q.tracker.recordSucceeded(ctx, env)
		}
		return nil
	}

	skipRetry := errors.Is(err, asynq.SkipRetry)
	retriesExhausted := env.MaxRetry > 0 && env.Attempt+1 >= env.MaxRetry

	if q.tracker != nil {
		if skipRetry || retriesExhausted {
			q.tracker.recordDLQ(ctx, env, err)
		} else {
			q.tracker.recordRetrying(ctx, env)
		}
	}

	return ErrProcessJobTask.WithCause(err)
}

func (q *Queue) retryDelay(retried int, _ error, task *asynq.Task) time.Duration {
	delays := q.cfg.RetryDelays
	if task != nil {
		if env, err := decodeEnvelope(task.Payload()); err == nil && len(env.RetryDelays) > 0 {
			delays = env.RetryDelays
		}
	}
	if len(delays) == 0 {
		return time.Second
	}
	if retried <= 0 {
		retried = 1
	}
	idx := retried - 1
	if idx >= len(delays) {
		idx = len(delays) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return delays[idx]
}

func (q *Queue) resolveRetryConfig(job Job) (int, []time.Duration) {
	if jrc, ok := job.(JobWithRetryConfig); ok {
		rc := jrc.RetryConfig()
		maxRetry := rc.MaxRetry
		if maxRetry <= 0 {
			maxRetry = q.cfg.MaxRetry
		}
		delays := rc.RetryDelays
		if len(delays) == 0 {
			delays = q.cfg.RetryDelays
		}
		return maxRetry, delays
	}
	if jm, ok := job.(JobWithMaxRetry); ok {
		if mr := jm.MaxRetry(); mr > 0 {
			return mr, q.cfg.RetryDelays
		}
	}
	return q.cfg.MaxRetry, q.cfg.RetryDelays
}

func (q *Queue) onTaskError(ctx context.Context, task *asynq.Task, err error) {
	retried, retriedOK := asynq.GetRetryCount(ctx)
	maxRetry, maxRetryOK := asynq.GetMaxRetry(ctx)
	if !retriedOK || !maxRetryOK {
		logger.Warn("asynq task failed", zap.String("type", task.Type()), zap.Error(err))
		return
	}
	logger.Warn("asynq task failed",
		zap.String("type", task.Type()),
		zap.Int("retried", retried),
		zap.Int("max_retry", maxRetry),
		zap.Bool("retry_exhausted", retried >= maxRetry),
		zap.Error(err),
	)
}

func taskType(name string) string {
	return "queue:" + name
}

func (q *Queue) resolveQueue(name string, j Job) string {
	if jq, ok := j.(JobWithQueue); ok {
		if queue := jq.Queue(); queue != "" {
			return queue
		}
	}
	return q.router.JobQueue(name)
}

func buildAsynqRedisConnOpt(cfg RedisConfig) asynq.RedisConnOpt {
	if cfg.ClusterAddrs != "" {
		return asynq.RedisClusterClientOpt{
			Addrs:    parseRedisClusterAddresses(cfg.ClusterAddrs),
			Password: cfg.Password,
		}
	}

	host := cfg.Host
	if host == "" {
		host = "localhost"
	}
	port := cfg.Port
	if port == 0 {
		port = 6379
	}

	return asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: cfg.Password,
		DB:       cfg.DB,
	}
}

func parseRedisClusterAddresses(addresses string) []string {
	parts := strings.Split(addresses, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		addr := strings.TrimSpace(part)
		if addr == "" {
			continue
		}
		if !strings.Contains(addr, ":") {
			addr += ":6379"
		}
		result = append(result, addr)
	}
	return result
}
