package schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/iVampireSP/go-template/pkg/logger"
)

// Event 代表一个调度事件，支持链式配置
type Event struct {
	name        string         // 事件名称/ID
	description string         // 描述
	expression  string         // Cron 表达式
	action      Action         // 执行动作
	timezone    *time.Location // 时区

	// 约束选项
	onOneServer        bool          // 分布式锁
	withoutOverlapping bool          // 防止重叠
	overlapTimeout     time.Duration // 重叠锁超时
	when               func() bool   // 条件执行
	skip               func() bool   // 跳过条件

	// 内部状态
	mutex   Mutex     // 锁实现
	lastRun time.Time // 上次执行时间
	logger  *logger.Logger
}

// newEvent 创建新的事件
func newEvent(name string, action Action, mutex Mutex) *Event {
	return &Event{
		name:           name,
		action:         action,
		mutex:          mutex,
		overlapTimeout: 30 * time.Minute,
		logger:         logger.With("component", "Scheduler", "event", name),
	}
}

// Name 返回事件名称
func (e *Event) Name() string {
	return e.name
}

// Description 返回事件描述
func (e *Event) Description() string {
	if e.description != "" {
		return e.description
	}
	return e.action.Description()
}

// Expression 返回 Cron 表达式
func (e *Event) Expression() string {
	return e.expression
}

// WithDescription 设置描述
func (e *Event) WithDescription(desc string) *Event {
	e.description = desc
	return e
}

// OnOneServer 启用分布式锁，确保只有一个实例执行
func (e *Event) OnOneServer() *Event {
	e.onOneServer = true
	return e
}

// WithoutOverlapping 防止任务重叠执行
func (e *Event) WithoutOverlapping(timeout ...time.Duration) *Event {
	e.withoutOverlapping = true
	if len(timeout) > 0 {
		e.overlapTimeout = timeout[0]
	}
	return e
}

// When 设置条件执行函数，返回 true 时执行
func (e *Event) When(fn func() bool) *Event {
	e.when = fn
	return e
}

// Skip 设置跳过条件函数，返回 true 时跳过
func (e *Event) Skip(fn func() bool) *Event {
	e.skip = fn
	return e
}

// Timezone 设置时区
func (e *Event) Timezone(loc *time.Location) *Event {
	e.timezone = loc
	return e
}

// run 执行事件
func (e *Event) run(ctx context.Context) error {
	if e.onOneServer && e.mutex != nil {
		lockKey := fmt.Sprintf("schedule:%s:one-server", e.name)
		ttl := e.estimatedDuration()
		acquired, err := e.mutex.Acquire(ctx, lockKey, ttl)
		if err != nil {
			return fmt.Errorf("failed to acquire one-server lock: %w", err)
		}
		if !acquired {
			e.logger.Debug("skipping event, another instance is running")
			return nil
		}
		defer func() {
			if err := e.mutex.Release(ctx, lockKey); err != nil {
				e.logger.Warn("failed to release one-server", "lock", err)
			}
		}()
	}

	if e.withoutOverlapping && e.mutex != nil {
		lockKey := fmt.Sprintf("schedule:%s:overlapping", e.name)
		acquired, err := e.mutex.Acquire(ctx, lockKey, e.overlapTimeout)
		if err != nil {
			return fmt.Errorf("failed to acquire overlap lock: %w", err)
		}
		if !acquired {
			e.logger.Debug("skipping event due to overlapping")
			return nil
		}
		defer func() {
			if err := e.mutex.Release(ctx, lockKey); err != nil {
				e.logger.Warn("failed to release overlap", "lock", err)
			}
		}()
	}

	if e.when != nil && !e.when() {
		e.logger.Debug("skipping event, condition not met")
		return nil
	}
	if e.skip != nil && e.skip() {
		e.logger.Debug("skipping event, skip condition met")
		return nil
	}

	e.lastRun = time.Now()
	return e.action.Run(ctx)
}

// estimatedDuration 估计任务执行时长，用于设置锁 TTL
func (e *Event) estimatedDuration() time.Duration {
	return time.Hour
}

// EventInfo 事件信息（用于列表展示）
type EventInfo struct {
	Name               string `json:"name"`
	Description        string `json:"description"`
	Expression         string `json:"expression"`
	OnOneServer        bool   `json:"on_one_server"`
	WithoutOverlapping bool   `json:"without_overlapping"`
	LastRun            string `json:"last_run,omitempty"`
}

// Info 返回事件信息
func (e *Event) Info() EventInfo {
	info := EventInfo{
		Name:               e.name,
		Description:        e.Description(),
		Expression:         e.expression,
		OnOneServer:        e.onOneServer,
		WithoutOverlapping: e.withoutOverlapping,
	}
	if !e.lastRun.IsZero() {
		info.LastRun = e.lastRun.Format(time.RFC3339)
	}
	return info
}
