package validator

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/gookit/validate"
)

func init() {
	validate.AddValidator("myCheck0", func(val any) bool {
		return true
	})
}

// Struct validates a struct and returns Huma-compatible error details.
// Returns the list of errors and a huma.StatusError if validation failed.
func Struct(data any) ([]*huma.ErrorDetail, error) {
	v := validate.Struct(data)

	if v.Validate() {
		return nil, nil
	}

	var details []*huma.ErrorDetail
	for field, errs := range v.Errors {
		for _, errMsg := range errs {
			details = append(details, &huma.ErrorDetail{
				Location: "body." + field,
				Message:  errMsg,
			})
		}
	}

	return details, huma.Error422UnprocessableEntity("validation failed", toErrors(details)...)
}

// toErrors converts ErrorDetail slice to error slice for huma.NewError
func toErrors(details []*huma.ErrorDetail) []error {
	errs := make([]error, len(details))
	for i, d := range details {
		errs[i] = d
	}
	return errs
}
