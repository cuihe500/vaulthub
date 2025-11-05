package errors

const (
	CodeSuccess = 0
)

const (
	CodeInvalidParam     = 10001
	CodeMissingParam     = 10002
	CodeInvalidFormat    = 10003
	CodeParamOutOfRange  = 10004
	CodeInvalidJSON      = 10005
	CodeValidationFailed = 10006
)

const (
	CodeUnauthorized        = 20001
	CodeInvalidToken        = 20002
	CodeTokenExpired        = 20003
	CodeInvalidCredentials  = 20004
	CodeAccountLocked       = 20005
	CodeAccountNotActivated = 20006
	CodeAccountDisabled     = 20007
	CodeWeakPassword        = 20008
	CodeUsernameExists      = 20009
)

const (
	CodeForbidden              = 30001
	CodeInsufficientPermission = 30002
	CodeOperationNotAllowed    = 30003
	CodeResourceLocked         = 30004
)

const (
	CodeResourceNotFound      = 40001
	CodeResourceAlreadyExists = 40002
	CodeResourceConflict      = 40003
	CodeResourceDeleted       = 40004
)

const (
	CodeInternalError = 50001
	CodeDatabaseError = 50002
	CodeCacheError    = 50003
	CodeConfigError   = 50004
	CodeCryptoError   = 50005
	CodeUnknownError  = 50099
)

const (
	CodeExternalServiceError       = 60001
	CodeExternalServiceTimeout     = 60002
	CodeExternalServiceUnavailable = 60003
)

var codeMessages = map[int]string{
	CodeSuccess: "success",

	CodeInvalidParam:     "invalid parameter",
	CodeMissingParam:     "missing required parameter",
	CodeInvalidFormat:    "invalid format",
	CodeParamOutOfRange:  "parameter out of range",
	CodeInvalidJSON:      "invalid JSON format",
	CodeValidationFailed: "validation failed",

	CodeUnauthorized:        "unauthorized",
	CodeInvalidToken:        "invalid token",
	CodeTokenExpired:        "token expired",
	CodeInvalidCredentials:  "invalid credentials",
	CodeAccountLocked:       "account locked",
	CodeAccountNotActivated: "account not activated",
	CodeAccountDisabled:     "account disabled",
	CodeWeakPassword:        "password does not meet security requirements",
	CodeUsernameExists:      "username already exists",

	CodeForbidden:              "forbidden",
	CodeInsufficientPermission: "insufficient permission",
	CodeOperationNotAllowed:    "operation not allowed",
	CodeResourceLocked:         "resource locked",

	CodeResourceNotFound:      "resource not found",
	CodeResourceAlreadyExists: "resource already exists",
	CodeResourceConflict:      "resource conflict",
	CodeResourceDeleted:       "resource deleted",

	CodeInternalError: "internal error",
	CodeDatabaseError: "database error",
	CodeCacheError:    "cache error",
	CodeConfigError:   "configuration error",
	CodeCryptoError:   "cryptography error",
	CodeUnknownError:  "unknown error",

	CodeExternalServiceError:       "external service error",
	CodeExternalServiceTimeout:     "external service timeout",
	CodeExternalServiceUnavailable: "external service unavailable",
}

func GetMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "unknown error"
}
