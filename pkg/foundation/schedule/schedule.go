package schedule

import (
	"context"
	"strings"

	jobqueue "github.com/iVampireSP/go-template/pkg/foundation/queue"
	"github.com/spf13/cobra"
)

// Schedule 用于定义调度任务
type Schedule struct {
	events  []*Event
	queue   *jobqueue.Queue // 用于 Job 动作
	mutex   Mutex           // 用于分布式锁
	rootCmd *cobra.Command  // 用于命令调用
}

// NewSchedule 创建调度定义器
func NewSchedule(q *jobqueue.Queue, mutex Mutex, rootCmd *cobra.Command) *Schedule {
	return &Schedule{
		events:  make([]*Event, 0),
		queue:   q,
		mutex:   mutex,
		rootCmd: rootCmd,
	}
}

// Command 创建命令任务
func (s *Schedule) Command(name string, args ...string) *Event {
	action := NewCommandAction(s.rootCmd, name, args...)
	eventName := name
	if !strings.Contains(name, ":") {
		if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
			eventName = name + ":" + args[0]
			args = args[1:]
			action = NewCommandAction(s.rootCmd, eventName, args...)
		}
	}
	event := newEvent(eventName, action, s.mutex)
	s.events = append(s.events, event)
	return event
}

// Job 创建队列任务
func (s *Schedule) Job(job jobqueue.Job) *Event {
	action := NewJobAction(job, s.queue)
	event := newEvent(job.Name(), action, s.mutex)
	s.events = append(s.events, event)
	return event
}

// Call 创建回调任务
func (s *Schedule) Call(name string, fn func(ctx context.Context) error) *Event {
	action := NewCallbackAction(fn, name)
	event := newEvent(name, action, s.mutex)
	s.events = append(s.events, event)
	return event
}

// Events 返回所有已定义的事件
func (s *Schedule) Events() []*Event {
	return s.events
}
