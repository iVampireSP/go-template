package cmd

import (
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/spf13/cobra"
	v1 "go-template/proto/gen/proto/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"sync"
)

func init() {
	RootCmd.AddCommand(rpcServerCommand)
}

var rpcServerCommand = &cobra.Command{
	Use:   "serve",
	Short: "Start gRPC",
	Run: func(cmd *cobra.Command, args []string) {
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", app.Config.Listen.Host, app.Config.Listen.Port))
		if err != nil {
			app.Logger.Sugar.Fatal(err)
		}
		var opts = []grpc.ServerOption{
			// 不再支持认证
			grpc.ChainUnaryInterceptor(
				logging.UnaryServerInterceptor(app.Api.GRPC.Interceptor.Logger.ZapLogInterceptor()),
			),
			grpc.ChainStreamInterceptor(
				logging.StreamServerInterceptor(app.Api.GRPC.Interceptor.Logger.ZapLogInterceptor()),
			),
		}
		grpcServer := grpc.NewServer(opts...)

		// 注册服务
		v1.RegisterDocumentServiceServer(grpcServer, app.Api.GRPC.DocumentApi)

		// 注册反射
		reflection.Register(grpcServer)

		app.Logger.Sugar.Infof("gRPC listening on %s:%d",
			app.Config.Listen.Host, app.Config.Listen.Port)

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			if err := grpcServer.Serve(lis); err != nil {
				app.Logger.Sugar.Fatal(err)
			}
		}()

		wg.Wait()

	},
}
