package milvus

import (
	"context"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"go-template/internal/base/conf"
	"go-template/internal/base/logger"

	"strconv"
)

func NewService(config *conf.Config, logger *logger.Logger) client.Client {
	var address = config.Milvus.Host + ":" + strconv.Itoa(config.Milvus.Port)

	logger.Sugar.Infof("Connecting Milvus, address=%s, dbname=%s", address, config.Milvus.DBName)

	c, err := client.NewClient(context.Background(), client.Config{
		Address: address,
		DBName:  config.Milvus.DBName,
	})

	logger.Sugar.Infof("Milvus connected!")

	if err != nil {
		panic(err)
	}

	return c
}
