package apperrors

import "errors"

type ServiceError struct {
	Code    string
	Message string
	Err     error
}

func (e *ServiceError) Error() string {
	return e.Message
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

func Wrap(err error) error {
	if err == nil {
		return nil
	}

	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		return err
	}

	return &ServiceError{
		Code:    "INTERNAL_ERROR",
		Message: "service operation failed",
		Err:     err,
	}
}

func WrapDatabase(err error) error {
	if err == nil {
		return nil
	}

	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		return err
	}

	return &ServiceError{
		Code:    "DATABASE_ERROR",
		Message: "database operation failed",
		Err:     err,
	}
}
