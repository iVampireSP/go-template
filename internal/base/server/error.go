package server

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"go-template/internal/base/logger"
	"go-template/internal/pkg/validator"
	"go-template/internal/types/dto"
	"go-template/internal/types/errs"
	"gorm.io/gorm"
	"net/http"
)

func errorConverter(logger *logger.Logger, ctx *fiber.Ctx, err error) error {
	status := http.StatusInternalServerError

	if err == nil {
		return dto.Ctx(ctx).Error(errs.ErrInternalServerError).Status(status).Send()
	}

	var errorMsg dto.IError

	switch {
	// 404
	//case errors.Is(err, fiber.ErrNotFound):
	//	status = http.StatusNotFound
	//	errorMsg = errs.RouteNotFound

	case errors.Is(err, fiber.ErrUnprocessableEntity):
		status = http.StatusUnprocessableEntity
		errorMsg = err

	case errors.Is(err, errs.ErrPermissionDenied):
		status = http.StatusForbidden

	case errors.Is(err, validator.ErrValidationFailed):
		status = http.StatusBadRequest

	case errors.Is(err, gorm.ErrRecordNotFound):
		errorMsg = errs.ErrNotFound

	default:
		logger.Sugar.Errorf("fiber error: %s", err)
		errorMsg = errs.UnknownError
	}

	return dto.Ctx(ctx).Status(status).Error(errorMsg).Send()
}
