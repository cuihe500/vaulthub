package errors

import "fmt"

type AppError struct {
	Code    int
	Message string
	Type    ErrorType
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code int, message string) *AppError {
	if message == "" {
		message = GetMessage(code)
	}
	return &AppError{
		Code:    code,
		Message: message,
		Type:    GetErrorType(code),
	}
}

func Wrap(code int, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: GetMessage(code),
		Type:    GetErrorType(code),
		Err:     err,
	}
}

func WithMessage(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Type:    GetErrorType(code),
		Err:     err,
	}
}

func NewValidationError(message string) *AppError {
	return New(CodeValidationFailed, message)
}

func NewAuthError(message string) *AppError {
	return New(CodeUnauthorized, message)
}

func NewPermissionError(message string) *AppError {
	return New(CodeForbidden, message)
}

func NewNotFoundError(message string) *AppError {
	return New(CodeResourceNotFound, message)
}

func NewInternalError(err error) *AppError {
	return Wrap(CodeInternalError, err)
}
