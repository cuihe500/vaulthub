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
	CodeSuccess: "成功",

	CodeInvalidParam:     "参数无效",
	CodeMissingParam:     "缺少必需参数",
	CodeInvalidFormat:    "格式无效",
	CodeParamOutOfRange:  "参数超出范围",
	CodeInvalidJSON:      "JSON格式无效",
	CodeValidationFailed: "验证失败",

	CodeUnauthorized:        "未授权",
	CodeInvalidToken:        "令牌无效",
	CodeTokenExpired:        "令牌已过期",
	CodeInvalidCredentials:  "凭证无效",
	CodeAccountLocked:       "账户已锁定",
	CodeAccountNotActivated: "账户未激活",
	CodeAccountDisabled:     "账户已禁用",
	CodeWeakPassword:        "密码不符合安全要求",
	CodeUsernameExists:      "用户名已存在",

	CodeForbidden:              "禁止访问",
	CodeInsufficientPermission: "权限不足",
	CodeOperationNotAllowed:    "操作不被允许",
	CodeResourceLocked:         "资源已锁定",

	CodeResourceNotFound:      "资源未找到",
	CodeResourceAlreadyExists: "资源已存在",
	CodeResourceConflict:      "资源冲突",
	CodeResourceDeleted:       "资源已删除",

	CodeInternalError: "内部错误",
	CodeDatabaseError: "数据库错误",
	CodeCacheError:    "缓存错误",
	CodeConfigError:   "配置错误",
	CodeCryptoError:   "加密错误",
	CodeUnknownError:  "未知错误",

	CodeExternalServiceError:       "外部服务错误",
	CodeExternalServiceTimeout:     "外部服务超时",
	CodeExternalServiceUnavailable: "外部服务不可用",
}

func GetMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
