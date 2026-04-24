package event

import "github.com/iVampireSP/go-template/pkg/foundation/bus"

// UserRegistered 用户注册事件
type UserRegistered struct {
	UserID      int    `json:"user_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

func (e UserRegistered) Name() string       { return "user.registered" }
func (e UserRegistered) Key() string        { return e.Email }
func (e UserRegistered) Topic() bus.TopicID { return bus.TopicDefault }

// UserDeleted 用户删除事件
type UserDeleted struct {
	UserID int `json:"user_id"`
}

func (e UserDeleted) Name() string       { return "user.deleted" }
func (e UserDeleted) Key() string        { return "user.deleted" }
func (e UserDeleted) Topic() bus.TopicID { return bus.TopicDefault }
