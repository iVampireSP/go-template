package schedule

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	jobqueue "github.com/iVampireSP/go-template/internal/infra/queue"
	"github.com/spf13/cobra"
)

// Action 定义可执行的动作接口
type Action interface {
	// Run 执行动作
	Run(ctx context.Context) error
	// Description 动作描述
	Description() string
}

// CommandAction 通过 Cobra 执行命令
type CommandAction struct {
	rootCmd *cobra.Command
	name    string
	args    []string
}

// NewCommandAction 创建命令动作
func NewCommandAction(rootCmd *cobra.Command, name string, args ...string) *CommandAction {
	return &CommandAction{
		rootCmd: rootCmd,
		name:    name,
		args:    args,
	}
}

// Run 执行命令（通过 Cobra 查找并执行子命令）
func (a *CommandAction) Run(ctx context.Context) error {
	parts, err := parseCommandActionName(a.name)
	if err != nil {
		return err
	}

	if a.rootCmd != nil {
		if _, _, findErr := a.rootCmd.Find(parts); findErr != nil {
			return fmt.Errorf("command '%s' not found: %w", a.name, findErr)
		}
	}

	commandArgs := make([]string, 0, len(parts)+len(a.args))
	commandArgs = append(commandArgs, parts...)
	commandArgs = append(commandArgs, a.args...)
	cmd := exec.CommandContext(ctx, os.Args[0], commandArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Description 返回描述
func (a *CommandAction) Description() string {
	if len(a.args) > 0 {
		return fmt.Sprintf("Command: %s %s", a.name, strings.Join(a.args, " "))
	}
	return fmt.Sprintf("Command: %s", a.name)
}

// JobAction 分发队列任务
type JobAction struct {
	job   jobqueue.Job
	queue *jobqueue.Queue
}

// NewJobAction 创建任务动作
func NewJobAction(job jobqueue.Job, q *jobqueue.Queue) *JobAction {
	return &JobAction{
		job:   job,
		queue: q,
	}
}

// Run 分发任务到队列
func (a *JobAction) Run(ctx context.Context) error {
	_, err := a.queue.Dispatch(ctx, a.job)
	return err
}

// Description 返回描述
func (a *JobAction) Description() string {
	return fmt.Sprintf("Job: %s", a.job.Name())
}

// CallbackAction 执行回调函数
type CallbackAction struct {
	fn          func(ctx context.Context) error
	description string
}

// NewCallbackAction 创建回调动作
func NewCallbackAction(fn func(ctx context.Context) error, description string) *CallbackAction {
	return &CallbackAction{
		fn:          fn,
		description: description,
	}
}

// Run 执行回调
func (a *CallbackAction) Run(ctx context.Context) error {
	return a.fn(ctx)
}

// Description 返回描述
func (a *CallbackAction) Description() string {
	if a.description != "" {
		return fmt.Sprintf("Callback: %s", a.description)
	}
	return "Callback"
}

func parseCommandActionName(name string) ([]string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("command name is required")
	}

	rawParts := strings.Split(name, ":")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("invalid command name: %s", name)
		}
		parts = append(parts, part)
	}

	return parts, nil
}
