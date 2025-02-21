package batch

import (
	"go-template/internal/infra/logger"
)

type Batch struct {
	logger *logger.Logger
}

func NewBatch(
	logger *logger.Logger,
) *Batch {
	//infra.NewApplication()
	return &Batch{logger}
}
