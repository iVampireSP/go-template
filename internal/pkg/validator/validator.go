package validator

import (
	"errors"
	"github.com/gookit/validate"
	"go-template/internal/types/dto"
)

var (
	ErrValidationFailed = errors.New("validation field failed")
)

func init() {
	// 可以自定义验证
	// Custom struct validation tag format
	//if err := validate.Struct("teener", func(fl validator.FieldLevel) bool {
	//	// User.Age needs to fit our needs, 12-18 years old.
	//	return fl.Field().Int() >= 12 && fl.Field().Int() <= 18
	//}); err != nil {
	//	panic(err)
	//}

	validate.AddValidator("myCheck0", func(val any) bool {
		// do validate val ...
		return true
	})
}

func Struct(data interface{}) (validationErrors *[]dto.ValidateError, err error) {
	v := validate.Struct(data)

	var e error
	var ves []dto.ValidateError

	if v.Validate() {
		return &ves, e // 返回指针
	} else {
		e = ErrValidationFailed

		for _, err := range v.Errors {
			ves = append(ves, dto.ValidateError{
				Message: err.String(),
			})
		}
	}

	//if errs != nil {
	//	for
	//	//for _, err := range errs.(validator.ValidationErrors) {
	//	//	// In this case data object is actually holding the User struct
	//	//	var elem ErrorResponse
	//	//
	//	//	elem.FailedField = err.Field() // Export struct field name
	//	//	elem.Tag = err.Tag()           // Export struct tag
	//	//	elem.Value = err.Value()       // Export field value
	//	//	elem.Error = true
	//	//
	//	//	validationErrors = append(validationErrors, elem)
	//	//}
	//}

	return &ves, e
}
