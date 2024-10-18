package cmd

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/spf13/cobra"
	"go-template/pkg/protos/documentService"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

func init() {
	RootCmd.AddCommand(documentServiceCommand)
}

var documentServiceCommand = &cobra.Command{
	Use:   "document",
	Short: "Start document service",
	Run: func(cmd *cobra.Command, args []string) {
		app, err := CreateApp()
		if err != nil {
			panic(err)
			return
		}

		app.Logger.Sugar.Info("Start document service")

		lis, err := net.Listen("tcp", app.Config.Grpc.Address)
		if err != nil {
			app.Logger.Sugar.Fatal(err)
		}
		var opts = []grpc.ServerOption{
			grpc.ChainUnaryInterceptor(
				logging.UnaryServerInterceptor(app.Handler.GRPC.Interceptor.Logger.ZapLogInterceptor()),

				auth.UnaryServerInterceptor(app.Handler.GRPC.Interceptor.Auth.JwtAuth),
			),
			grpc.ChainStreamInterceptor(
				logging.StreamServerInterceptor(app.Handler.GRPC.Interceptor.Logger.ZapLogInterceptor()),
				auth.StreamServerInterceptor(app.Handler.GRPC.Interceptor.Auth.JwtAuth),
			),
		}
		grpcServer := grpc.NewServer(opts...)

		documentService.RegisterDocumentServiceServer(grpcServer, app.Handler.GRPC.DocumentService)
		reflection.Register(grpcServer)

		app.Logger.Sugar.Info("Document Service listing on " + app.Config.Grpc.Address)

		if err := grpcServer.Serve(lis); err != nil {
			app.Logger.Sugar.Fatal(err)
		}
	},
}
