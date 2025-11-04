package errors

type ErrorType string

const (
	TypeValidation     ErrorType = "ValidationError"
	TypeAuthentication ErrorType = "AuthenticationError"
	TypeAuthorization  ErrorType = "AuthorizationError"
	TypeResource       ErrorType = "ResourceError"
	TypeSystem         ErrorType = "SystemError"
	TypeExternal       ErrorType = "ExternalError"
	TypeUnknown        ErrorType = "UnknownError"
)

func GetErrorType(code int) ErrorType {
	switch {
	case code >= 10001 && code < 20000:
		return TypeValidation
	case code >= 20001 && code < 30000:
		return TypeAuthentication
	case code >= 30001 && code < 40000:
		return TypeAuthorization
	case code >= 40001 && code < 50000:
		return TypeResource
	case code >= 50001 && code < 60000:
		return TypeSystem
	case code >= 60001 && code < 70000:
		return TypeExternal
	default:
		return TypeUnknown
	}
}
