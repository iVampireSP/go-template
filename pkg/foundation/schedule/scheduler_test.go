package schedule

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSchedulerStartReturnsErrorWhenCronIsNil(t *testing.T) {
	scheduler := New(nil, nil, nil, &cobra.Command{Use: "app"})
	scheduler.Define().Call("test-event", func(ctx context.Context) error { return nil }).EveryMinute()

	err := scheduler.Start(context.Background())
	if err == nil {
		t.Fatalf("expected nil cron error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "cron instance is nil") {
		t.Fatalf("expected cron nil error, got %v", err)
	}
}
