package events

import "encoding/json"

type EventMessage interface {
	JSON() ([]byte, error)
}

type ProcessPostRequest struct {
	EventMessage
	PostId  string `json:"post_id"`
	Content string `json:"content"`
}

func (p *ProcessPostRequest) JSON() ([]byte, error) {
	return json.Marshal(p)
}

type ProcessPostResult struct {
	PostId   string   `json:"post_id"`
	Keywords []string `json:"keywords"`
}
