package tracing

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/iVampireSP/go-template/pkg/json"
)

type Trace struct {
	TraceID string
	Spans   []Span
}

type Span struct {
	OperationName string
	ServiceName   string
	Tags          map[string]any
}

var traceIDPattern = regexp.MustCompile(`^[a-f0-9]{16,32}$`)

func GetTrace(ctx context.Context, traceID string) (*Trace, bool, error) {
	t, err := NewTracing()
	if err != nil {
		if errors.Is(err, ErrTracingDisabled) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return t.GetTrace(ctx, traceID)
}

func (t *Tracing) GetTrace(ctx context.Context, traceID string) (*Trace, bool, error) {
	traceID = strings.TrimSpace(strings.ToLower(traceID))
	if traceID == "" {
		return nil, false, nil
	}
	if !traceIDPattern.MatchString(traceID) {
		return nil, false, ErrInvalidTraceID
	}

	endpoint := *t.queryTraceURL
	endpoint.Path = endpoint.Path + traceID

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, false, fmt.Errorf("create request: %w", err)
	}
	if t.queryUsername != "" || t.queryPassword != "" {
		req.SetBasicAuth(t.queryUsername, t.queryPassword)
	}

	resp, err := t.queryClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("tracing query request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, false, nil
	}
	if resp.StatusCode >= 300 {
		return nil, false, fmt.Errorf("tracing query request failed with status %d", resp.StatusCode)
	}

	var payload struct {
		Data []struct {
			TraceID string `json:"traceID"`
			Spans   []struct {
				OperationName string `json:"operationName"`
				ProcessID     string `json:"processID"`
				Tags          []struct {
					Key   string `json:"key"`
					Value any    `json:"value"`
				} `json:"tags"`
			} `json:"spans"`
			Processes map[string]struct {
				ServiceName string `json:"serviceName"`
			} `json:"processes"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"msg"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, false, fmt.Errorf("decode tracing query response: %w", err)
	}
	if len(payload.Errors) > 0 {
		return nil, false, fmt.Errorf("tracing query error: %s", payload.Errors[0].Message)
	}
	if len(payload.Data) == 0 {
		return nil, false, nil
	}

	first := payload.Data[0]
	out := &Trace{TraceID: first.TraceID, Spans: make([]Span, 0, len(first.Spans))}
	for _, span := range first.Spans {
		serviceName := ""
		if proc, ok := first.Processes[span.ProcessID]; ok {
			serviceName = proc.ServiceName
		}
		tags := make(map[string]any, len(span.Tags))
		for _, tag := range span.Tags {
			tags[tag.Key] = tag.Value
		}
		out.Spans = append(out.Spans, Span{
			OperationName: span.OperationName,
			ServiceName:   serviceName,
			Tags:          tags,
		})
	}

	return out, true, nil
}
