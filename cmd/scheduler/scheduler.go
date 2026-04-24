package scheduler

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/iVampireSP/go-template/internal/infra/cache"
	"github.com/iVampireSP/go-template/internal/infra/config"
	jobqueue "github.com/iVampireSP/go-template/internal/infra/queue"
	"github.com/iVampireSP/go-template/internal/infra/tracing"
	"github.com/iVampireSP/go-template/pkg/httpserver"
	"github.com/iVampireSP/go-template/pkg/logger"
	"github.com/iVampireSP/go-template/pkg/schedule"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

// Scheduler holds Scheduler service dependencies.
type Scheduler struct {
	cron     *cron.Cron
	locker   *cache.Locker
	mq       *jobqueue.Queue
	cronjobs []schedule.CronJob
}

// NewScheduler declares Scheduler command dependencies.
func NewScheduler(cron *cron.Cron, locker *cache.Locker, mq *jobqueue.Queue, cronjobs []schedule.CronJob) *Scheduler {
	return &Scheduler{
		cron:     cron,
		locker:   locker,
		mq:       mq,
		cronjobs: cronjobs,
	}
}

// Command constructs the scheduler cobra command.
func (s *Scheduler) Command() *cobra.Command {
	cmd := &cobra.Command{Use: "scheduler", Short: "Start cron queue scheduler"}
	cmd.Flags().Bool("once", false, "Run all jobs once and exit")
	cmd.Flags().StringP("queue", "j", "", "Run a specific job by name and exit")
	cmd.Flags().BoolP("list", "l", false, "List all registered cronjobs")
	return cmd
}

// Handle runs the scheduler service.
func (s *Scheduler) Handle(cmd *cobra.Command) error {
	runOnce, _ := cmd.Flags().GetBool("once")
	runJob, _ := cmd.Flags().GetString("queue")
	listJobs, _ := cmd.Flags().GetBool("list")

	ctx := cmd.Context()

	tp, err := tracing.GetService("app-scheduler")
	if err != nil {
		return err
	}
	defer tracing.ShutdownWithTimeout(tp)

	if s.cron == nil {
		return errors.New("scheduler bootstrap failed: cron instance is nil")
	}

	mutex := schedule.NewRedisMutex(s.locker)
	sched := schedule.New(s.cron, mutex, s.mq, cmd.Root())
	sched.RegisterAll(s.cronjobs)

	if listJobs {
		listAllEvents(sched)
		return nil
	}

	if runJob != "" {
		logger.Info("scheduler: running single", "job", runJob)
		if err := sched.RunEvent(ctx, runJob); err != nil {
			logger.Error("scheduler", "job", runJob, "failed", err)
			return err
		}
		logger.Info("scheduler completed", "job", runJob)
		return nil
	}

	if runOnce {
		logger.Info("scheduler: running jobs once", "all", len(s.cronjobs))
		if err := sched.RunAllEvents(ctx); err != nil {
			logger.Error("scheduler: run-all", "failed", err)
			return err
		}
		logger.Info("scheduler: all jobs completed")
		return nil
	}

	server := httpserver.New("scheduler", "1.0.0", httpserver.WithMetrics(httpserver.MetricsConfig{
		Enabled:         config.Bool("metrics.enabled", true),
		Host:            config.String("metrics.host", "0.0.0.0"),
		Port:            config.Int("metrics.port", 9090),
		ShutdownTimeout: 30 * time.Second,
	}))
	if err := server.Start(); err != nil {
		return err
	}
	defer server.ShutdownWithTimeout()

	if err := sched.Start(ctx); err != nil {
		return err
	}
	defer sched.Stop()

	s.cron.Start()
	defer s.cron.Stop()

	logger.Info("scheduler: started cronjobs, waiting for signals", "with", len(s.cronjobs))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info("scheduler: received signal, stopping", "signal", sig)
	return nil
}

func listAllEvents(sched *schedule.Scheduler) {
	events := sched.ListEvents()
	if len(events) == 0 {
		fmt.Println("No cronjobs registered.")
		return
	}

	nameW, exprW, descW := len("NAME"), len("SCHEDULE"), len("DESCRIPTION")
	for _, e := range events {
		if len(e.Name) > nameW {
			nameW = len(e.Name)
		}
		if len(e.Expression) > exprW {
			exprW = len(e.Expression)
		}
		if len(e.Description) > descW {
			descW = len(e.Description)
		}
	}

	rowFmt := fmt.Sprintf("  %%-%ds  %%-%ds  %%-%ds  %%s\n", nameW, exprW, descW)
	fmt.Printf(rowFmt, "NAME", "SCHEDULE", "DESCRIPTION", "FLAGS")
	fmt.Printf("  %s  %s  %s  %s\n",
		strings.Repeat("-", nameW),
		strings.Repeat("-", exprW),
		strings.Repeat("-", descW),
		strings.Repeat("-", 5),
	)
	for _, e := range events {
		var flags []string
		if e.OnOneServer {
			flags = append(flags, "one-server")
		}
		if e.WithoutOverlapping {
			flags = append(flags, "no-overlap")
		}
		fmt.Printf(rowFmt, e.Name, e.Expression, e.Description, strings.Join(flags, ", "))
	}
	fmt.Printf("\n  %d cronjob(s) total\n", len(events))
}
