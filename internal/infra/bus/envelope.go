package bus

import (
	"time"

	"github.com/iVampireSP/go-template/pkg/json"
	"github.com/google/uuid"
)

type Envelope struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Key       string            `json:"key"`
	Payload   json.RawMessage   `json:"payload"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
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
		Metadata:  map[string]string{},
		CreatedAt: time.Now(),
	}, nil
}

func (e *Envelope) Encode() ([]byte, error) {
	return json.Marshal(e)
}

func decodeEnvelope(data []byte) (*Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

type DLQEnvelope struct {
	Original Envelope  `json:"original"`
	Error    string    `json:"error"`
	FailedAt time.Time `json:"failed_at"`
}

func newDLQEnvelope(original *Envelope, err error) *DLQEnvelope {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return &DLQEnvelope{
		Original: *original,
		Error:    msg,
		FailedAt: time.Now(),
	}
}
