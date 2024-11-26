package cmd

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	"go-template/pkg/protos/documentService"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
	"sync"
)

func init() {
	RootCmd.AddCommand(documentServiceCommand)
}

var documentServiceCommand = &cobra.Command{
	Use:   "grpc",
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

				app.Handler.GRPC.Interceptor.Auth.UnaryJWTAuth(),
			),
			grpc.ChainStreamInterceptor(
				logging.StreamServerInterceptor(app.Handler.GRPC.Interceptor.Logger.ZapLogInterceptor()),
				app.Handler.GRPC.Interceptor.Auth.StreamJWTAuth(),
			),
		}
		grpcServer := grpc.NewServer(opts...)

		documentService.RegisterDocumentServiceServer(grpcServer, app.Handler.GRPC.DocumentService)
		reflection.Register(grpcServer)

		app.Logger.Sugar.Infof("Document Service listening on %s, http gateway listening on %s",
			app.Config.Grpc.Address, app.Config.Grpc.AddressGateway)

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			if err := grpcServer.Serve(lis); err != nil {
				app.Logger.Sugar.Fatal(err)
			}
		}()

		wg.Add(1)
		go func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mux := runtime.NewServeMux()
			err := documentService.RegisterDocumentServiceHandlerFromEndpoint(ctx, mux, app.Config.Grpc.Address,
				[]grpc.DialOption{
					grpc.WithTransportCredentials(insecure.NewCredentials()),
				})

			if err != nil {
				app.Logger.Sugar.Fatal(err)
			}

			if err := http.ListenAndServe(app.Config.Grpc.AddressGateway, mux); err != nil {
				app.Logger.Sugar.Fatal(err)
			}

		}()

		wg.Wait()

	},
}
