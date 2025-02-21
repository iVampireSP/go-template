package cmd

import (
	"github.com/hibiken/asynq"
	"github.com/spf13/cobra"
	"go-template/internal/infra"
	"go-template/internal/tasks"
	"sync"
)

func init() {
	RootCmd.AddCommand(scheduleCmd)
}

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Schedule commands",
	Long:  `Schedule commands`,
	Run: func(cmd *cobra.Command, args []string) {
		app, err := CreateApp()
		if err != nil {
			panic(err)
		}

		var wg = &sync.WaitGroup{}

		defer wg.Wait()
		wg.Add(1)

		// 启动任务调度器
		go func() {
			app.Logger.Sugar.Info("任务调度器启动")
			runSchedule(app)
			app.Logger.Sugar.Info("任务调度器停止")
			wg.Done()
		}()

		//wg.Add(1)
		//// 启动 Kafka 事件推 MQTT
		//go func() {
		//	app.Logger.Sugar.Info("Kafka 事件推 MQTT 启动")
		//
		//	reader := app.Service.Stream.Consumer(app.Config.Kafka.Topics.InternalEvents, "internal-kafka-mqtt")
		//	//message, err := reader.ReadMessage(context.TODO())
		//	//if err != nil {
		//	//	app.Logger.Sugar.Error("读取 Kafka 消息失败:", err)
		//	//	return
		//	//}
		//
		//	for {
		//		msg, err := reader.ReadMessage(context.TODO())
		//		if err != nil {
		//			break
		//		}
		//		app.Logger.Sugar.Info("接收到 Kafka 消息:", string(msg.Value))
		//
		//		t := app.Stream.MQTT.Publish(app.Config.Stream.MQTTPublishTopic, 0, false, msg.Value)
		//		if t.Error() != nil {
		//			app.Logger.Sugar.Error("发布 MQTT 消息失败:", t.Error())
		//			return
		//		}
		//	}
		//
		//	wg.Done()
		//}()

	},
}

func runSchedule(app *infra.Application) {
	// 创建任务处理器
	handler := tasks.NewHandler(app)

	// 如果是调试模式，立即执行所有任务
	if app.Config.Debug.Enabled {
		app.Logger.Sugar.Info("调试模式已启用，将立即执行所有计划任务")
		handler.RunDebugTasks()
		return
	}

	// 创建任务调度器
	scheduler, err := tasks.NewScheduler(app)
	if err != nil {
		app.Logger.Sugar.Fatal("创建调度器失败:", err)
	}

	// 创建任务服务器
	srv := tasks.NewServer(app)

	// 注册任务处理器
	mux := asynq.NewServeMux()
	tasks.RegisterHandlers(mux, handler)

	// 启动调度器
	go func() {
		if err := scheduler.Run(); err != nil {
			app.Logger.Sugar.Fatal("调度器启动失败:", err)
		}
	}()

	// 启动任务处理服务器
	if err := srv.Run(mux); err != nil {
		app.Logger.Sugar.Fatal("任务处理服务器启动失败:", err)
	}
}
