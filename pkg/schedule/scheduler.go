package schedule

import (
	"context"
	"fmt"
	"sort"
	"sync"

	jobqueue "github.com/iVampireSP/go-template/internal/infra/queue"
	"github.com/iVampireSP/go-template/pkg/logger"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

// Scheduler 主调度器
type Scheduler struct {
	cron     *cron.Cron
	schedule *Schedule
	mutex    Mutex
	queue    *jobqueue.Queue
	rootCmd  *cobra.Command
	logger   *logger.Logger

	mu      sync.RWMutex
	running bool
}

// New 创建新的调度器
func New(c *cron.Cron, mutex Mutex, q *jobqueue.Queue, rootCmd *cobra.Command) *Scheduler {
	return &Scheduler{
		cron:    c,
		mutex:   mutex,
		queue:   q,
		rootCmd: rootCmd,
		logger:  logger.With("component", "Scheduler"),
	}
}

// Define 返回 Schedule 供定义任务
func (s *Scheduler) Define() *Schedule {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.schedule == nil {
		s.schedule = NewSchedule(s.queue, s.mutex, s.rootCmd)
	}
	return s.schedule
}

// Start 启动所有已注册的任务
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler already running")
	}

	if s.schedule == nil {
		s.logger.Warn("no events defined, scheduler will be idle")
		s.running = true
		return nil
	}
	if s.cron == nil {
		return fmt.Errorf("cron instance is nil")
	}

	s.logger.Info("starting scheduled tasks...")

	for _, event := range s.schedule.Events() {
		if err := s.scheduleEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to start task '%s': %w", event.Name(), err)
		}
	}

	s.printEvents()
	s.running = true

	return nil
}

// scheduleEvent 将事件添加到 cron 调度器
func (s *Scheduler) scheduleEvent(ctx context.Context, event *Event) error {
	if s.cron == nil {
		return fmt.Errorf("cron instance is nil")
	}
	if event.Expression() == "" {
		return fmt.Errorf("event '%s' has no schedule expression", event.Name())
	}

	_, err := s.cron.AddFunc(event.Expression(), func() {
		s.logger.Info("starting", "name", event.Name())
		if err := event.run(ctx); err != nil {
			s.logger.Error("failed", "name", event.Name(), "error", err)
		} else {
			s.logger.Info("completed", "name", event.Name())
		}
	})
	if err != nil {
		return err
	}

	return nil
}

// printEvents 打印所有已注册的事件
func (s *Scheduler) printEvents() {
	events := s.schedule.Events()
	if len(events) == 0 {
		s.logger.Info("no tasks registered")
		return
	}

	sorted := make([]*Event, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name() < sorted[j].Name()
	})

	s.logger.Info("scheduled tasks started:")
	for _, event := range sorted {
		flags := ""
		if event.onOneServer {
			flags += " [OnOneServer]"
		}
		if event.withoutOverlapping {
			flags += " [WithoutOverlapping]"
		}
		s.logger.Info("task", "name", event.Name(), "description", event.Description(), "flags", flags)
	}
}

// RunEvent 立即执行指定事件
func (s *Scheduler) RunEvent(ctx context.Context, name string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.schedule == nil {
		return fmt.Errorf("no events defined")
	}

	for _, event := range s.schedule.Events() {
		if event.Name() == name {
			s.logger.Info("manual run", "name", name)
			if err := event.run(ctx); err != nil {
				s.logger.Error("failed", "name", name, "error", err)
				return err
			}
			s.logger.Info("completed", "name", name)
			return nil
		}
	}

	return fmt.Errorf("event '%s' not found", name)
}

// RunAllEvents 立即执行所有事件
func (s *Scheduler) RunAllEvents(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.schedule == nil {
		return fmt.Errorf("no events defined")
	}

	events := s.schedule.Events()
	if len(events) == 0 {
		s.logger.Info("no tasks to run")
		return nil
	}

	s.logger.Info("running all tasks...")

	sorted := make([]*Event, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name() < sorted[j].Name()
	})

	var hasError bool
	for _, event := range sorted {
		s.logger.Info("starting", "name", event.Name())
		if err := event.run(ctx); err != nil {
			s.logger.Error("failed", "name", event.Name(), "error", err)
			hasError = true
		} else {
			s.logger.Info("completed", "name", event.Name())
		}
	}

	if hasError {
		return fmt.Errorf("some tasks failed to execute")
	}

	s.logger.Info("all tasks completed")
	return nil
}

// ListEvents 列出所有已注册的事件
func (s *Scheduler) ListEvents() []EventInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.schedule == nil {
		return make([]EventInfo, 0)
	}

	events := s.schedule.Events()
	result := make([]EventInfo, 0, len(events))
	for _, event := range events {
		result = append(result, event.Info())
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// GetEvent 获取指定的事件
func (s *Scheduler) GetEvent(name string) (*Event, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.schedule == nil {
		return nil, false
	}

	for _, event := range s.schedule.Events() {
		if event.Name() == name {
			return event, true
		}
	}
	return nil, false
}

// RegisterAll 注册所有 CronJob 到调度器。
func (s *Scheduler) RegisterAll(jobs []CronJob) {
	sched := s.Define()
	for _, job := range jobs {
		event := sched.Call(job.Name(), job.Run)
		job.Schedule(event)
	}
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.running = false
	s.logger.Info("scheduler stopped")
}
