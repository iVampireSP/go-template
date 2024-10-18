package cmd

import (
	"github.com/spf13/cobra"
	"go-template/internal/handler/grpc/documents"
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
		s := grpc.NewServer()

		documentService.RegisterDocumentServiceServer(s, documents.NewDocumentService())
		reflection.Register(s)

		app.Logger.Sugar.Info("Document Service listing on " + app.Config.Grpc.Address)

		if err := s.Serve(lis); err != nil {
			app.Logger.Sugar.Fatal(err)
		}
	},
}
