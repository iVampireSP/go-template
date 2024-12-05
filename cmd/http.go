package cmd

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func init() {
	RootCmd.AddCommand(httpCmd)
}

var httpCmd = &cobra.Command{
	Use: "http",
	Run: func(cmd *cobra.Command, args []string) {
		initHttpServer()
	},
}

func initHttpServer() {
	app, err := CreateApp()
	if err != nil {
		panic(err)
		return
	}

	if app.Config.Http.Host == "" {
		app.Config.Http.Host = "0.0.0.0"
	}

	if app.Config.Http.Port == 0 {
		app.Config.Http.Port = 8000
	}

	var bizServer *http.Server
	var metricServer *http.Server

	bizServer = &http.Server{
		Addr: ":8080",
	}

	bizServer.Addr = app.Config.Http.Host + ":" + strconv.Itoa(app.Config.Http.Port)
	bizServer.Handler = adaptor.FiberApp(app.HttpServer.BizRouter())

	// 启动 http
	go func() {
		// refresh
		app.Service.Jwks.SetupAuthRefresh()

		app.Logger.Sugar.Info("Listening and serving HTTP on ", bizServer.Addr)
		err = bizServer.ListenAndServe()
		if err != nil && !errors.Is(http.ErrServerClosed, err) {
			panic(err)
			return
		}
	}()

	// 启动 metrics
	if app.Config.Metrics.Enabled {
		metricServer = &http.Server{
			Addr: ":8080",
		}
		metricServer.Addr = app.Config.Metrics.Host + ":" + strconv.Itoa(app.Config.Metrics.Port)
		go func() {
			app.Logger.Sugar.Info("Metrics and serving HTTP on ", metricServer.Addr)

			metricServer.Handler = adaptor.FiberApp(app.HttpServer.MetricRouter())

			err = metricServer.ListenAndServe()
			if err != nil && !errors.Is(http.ErrServerClosed, err) {
				panic(err)
				return
			}
		}()
	}

	quit := make(chan os.Signal)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	app.Logger.Sugar.Info("Shutdown http server")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	if err := bizServer.Shutdown(ctx); err != nil {
		app.Logger.Sugar.Fatalf("Biz Server Shutdown Error: %s", err)
	}

	if err := metricServer.Shutdown(ctx); err != nil {
		app.Logger.Sugar.Fatalf("Metric Server Shutdown Error: %s", err)
	}
}
