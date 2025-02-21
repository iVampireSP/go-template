package events

import (
	"time"
)

// FileEvent 文件事件基础结构
type FileEvent struct {
	WorkspaceName string    `json:"workspace_name"`
	Path          string    `json:"path"`
	Size          int64     `json:"size"`
	EventTime     time.Time `json:"event_time"`
}

func (fe *FileEvent) Name() string {
	return "file_event"
}
