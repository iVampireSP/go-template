package example

import (
	"context"
	"fmt"

	"github.com/iVampireSP/go-template/internal/event"
	"github.com/iVampireSP/go-template/pkg/foundation/bus"
	"github.com/iVampireSP/go-template/pkg/json"
	"github.com/iVampireSP/go-template/pkg/logger"
)

// Listener 示例事件监听器，演示如何订阅和处理事件。
type Listener struct {
	log *logger.Logger
}

// NewListener 创建示例事件监听器。
func NewListener() *Listener {
	return &Listener{
		log: logger.With("component", "ExampleListener"),
	}
}

// Handlers 实现 bus.Listener 接口，注册事件处理函数。
// key 是事件名称模式，支持通配符：
//   - "user.registered" — 精确匹配
//   - "user.*"          — 前缀匹配（匹配所有 user.xxx 事件）
//   - "*"               — 匹配所有事件
func (l *Listener) Handlers() map[string]bus.Handler {
	return map[string]bus.Handler{
		"user.registered": l.handleUserRegistered,
		"user.deleted":    l.handleUserDeleted,
	}
}

func (l *Listener) handleUserRegistered(ctx context.Context, payload []byte) error {
	var evt event.UserRegistered
	if err := json.Unmarshal(payload, &evt); err != nil {
		return fmt.Errorf("unmarshal UserRegistered: %w", err)
	}

	l.log.Info("new user registered",
		"user_id", evt.UserID,
		"email", evt.Email,
		"name", evt.DisplayName,
	)

	// 在这里可以：
	// - 发送欢迎邮件（通过 job queue 异步发送）
	// - 初始化用户资源
	// - 发送通知
	// - 触发后续流程

	return nil
}

func (l *Listener) handleUserDeleted(ctx context.Context, payload []byte) error {
	var evt event.UserDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return fmt.Errorf("unmarshal UserDeleted: %w", err)
	}

	l.log.Info("user deleted, cleaning up resources", "user_id", evt.UserID)

	// 在这里可以：
	// - 清理用户关联资源
	// - 发送告别邮件
	// - 归档用户数据

	return nil
}
