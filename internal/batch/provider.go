package batch

import (
	"go-template/internal/base/logger"
)

type Batch struct {
	logger *logger.Logger
}

func NewBatch(
	logger *logger.Logger,
) *Batch {
	//base.NewApplication()
	return &Batch{logger}
}
