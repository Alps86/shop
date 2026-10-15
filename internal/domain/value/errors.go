package value

import "errors"

type Error struct {
	Code  string
	Field string
}

func (err Error) Error() string {
	return err.Code + ": " + err.Field
}

func NewError(code, field string) error {
	return Error{Code: code, Field: field}
}

func Code(err error) string {
	if err == nil {
		return "OK"
	}

	var domain Error
	if errors.As(err, &domain) {
		return domain.Code
	}

	return "InternalError"
}
