package mail

import (
	"context"
	"fmt"
	"strings"

	"github.com/iVampireSP/go-template/pkg/json"

	infraemail "github.com/iVampireSP/go-template/pkg/foundation/email"
	jobqueue "github.com/iVampireSP/go-template/pkg/foundation/queue"
)

const JobName = "mail.send"

// Job 异步邮件 Job payload，Subject 和 HTML 由调用方渲染后传入。
type Job struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (j Job) Name() string { return JobName }
func (j Job) Key() string  { return strings.Join(j.To, ",") }

var _ jobqueue.Job = Job{}

// Dispatch 渲染 mailable 并将邮件任务投入队列。
func Dispatch(ctx context.Context, q *jobqueue.Queue, to []string, m infraemail.Mailable) error {
	content, err := m.Content(ctx)
	if err != nil {
		return fmt.Errorf("mail: render: %w", err)
	}
	_, dispatchErr := q.Dispatch(ctx, Job{To: to, Subject: content.Subject, HTML: content.HTML})
	return dispatchErr
}

// Handler 实现 job.Handler，接收已渲染内容并通过 SMTP 发送。
type Handler struct {
	emailSender *infraemail.Email
}

// NewHandler 创建邮件异步发送处理器。
func NewHandler(emailSender *infraemail.Email) *Handler {
	return &Handler{
		emailSender: emailSender,
	}
}

// JobName 实现 job.Handler 接口。
func (h *Handler) JobName() string {
	return JobName
}

// Handle 发送已渲染的邮件内容。
func (h *Handler) Handle(ctx context.Context, payload []byte) error {
	var job Job
	if err := json.Unmarshal(payload, &job); err != nil {
		return fmt.Errorf("mail: unmarshal job: %w", err)
	}
	return h.emailSender.SendRaw(ctx, job.To, job.Subject, job.HTML)
}
