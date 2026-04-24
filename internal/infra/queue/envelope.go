package queue

import (
	"time"

	"github.com/iVampireSP/go-template/pkg/json"
	"github.com/google/uuid"
)

// Envelope 消息包装
type Envelope struct {
	ID        string            `json:"id"`         // UUID
	Name      string            `json:"name"`       // 消息名称
	Key       string            `json:"key"`        // 分区键
	Queue     string            `json:"queue"`      // 队列名称
	Payload   json.RawMessage   `json:"payload"`    // 数据
	Metadata  map[string]string `json:"metadata"`   // 元数据
	CreatedAt time.Time         `json:"created_at"` // 创建时间

	// 重试信息（仅 Job）
	Attempt     int             `json:"attempt,omitempty"`
	MaxRetry    int             `json:"max_retry,omitempty"`
	RetryDelays []time.Duration `json:"retry_delays,omitempty"`
}

func newEnvelope(name, key string, payload any) (*Envelope, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Envelope{
		ID:        uuid.New().String(),
		Name:      name,
		Key:       key,
		Payload:   data,
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
	}, nil
}

// newJobEnvelope 创建任务消息包装
func newJobEnvelope(name, key string, payload any, maxRetry int, retryDelays []time.Duration) (*Envelope, error) {
	env, err := newEnvelope(name, key, payload)
	if err != nil {
		return nil, err
	}
	env.MaxRetry = maxRetry
	env.RetryDelays = retryDelays
	return env, nil
}

// Encode 编码为 JSON
func (e *Envelope) Encode() ([]byte, error) {
	return json.Marshal(e)
}

// decodeEnvelope 解码消息包装
func decodeEnvelope(data []byte) (*Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	return &env, nil
}
