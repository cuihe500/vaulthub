package middleware

// Casbin资源和动作常量定义
// 用于统一管理权限验证中的资源标识和操作类型，避免硬编码字符串导致的拼写错误

// 资源类型常量
// 定义系统中所有需要权限控制的资源类型
const (
	// ResourceUser 用户管理资源
	// 用于用户列表查询、用户信息获取、用户状态/角色管理等接口
	ResourceUser = "user"

	// ResourceProfile 用户档案资源
	// 用于用户档案的CRUD操作，包括管理员对其他用户档案的管理
	ResourceProfile = "profile"

	// ResourceConfig 系统配置资源
	// 用于系统配置的查询、更新、批量更新、重新加载等接口
	ResourceConfig = "config"

	// ResourceVault 密钥仓库资源
	// 用于密钥仓库的基本访问权限控制
	ResourceVault = "vault"

	// ResourceKey 加密密钥资源
	// 用于用户加密密钥的创建、验证、轮换、状态查询等操作
	ResourceKey = "key"

	// ResourceSecret 秘密资源
	// 用于加密秘密的创建、查询、解密、删除等操作
	ResourceSecret = "secret"

	// ResourceScope 数据作用域资源
	// 用于控制用户的数据访问范围（全局/个人）
	ResourceScope = "scope"

	// ResourceCasbin Casbin权限系统资源
	// 用于权限策略的重新加载等管理操作
	ResourceCasbin = "casbin"
)

// 操作类型常量
// 定义系统中所有权限验证的操作类型
const (
	// ActionRead 读操作
	// 用于查询、列表、获取详情等只读操作
	ActionRead = "read"

	// ActionWrite 写操作
	// 用于创建、更新、删除等修改数据的操作
	ActionWrite = "write"

	// ActionGlobal 全局访问权限
	// 特殊操作类型，用于控制是否可以访问全局数据（所有用户的数据）
	// 通常只授予admin角色，用于审计日志、统计数据等场景
	ActionGlobal = "global"

	// ActionReload 重新加载权限
	// 用于重新加载配置、权限策略等热更新操作
	ActionReload = "reload"
)

// 通配符常量
const (
	// WildcardResource 资源通配符
	// 表示所有资源，用于admin角色的超级权限
	WildcardResource = "*"

	// WildcardAction 操作通配符
	// 表示所有操作，用于admin角色的超级权限
	WildcardAction = "*"
)

// 预定义的权限组合（可选，用于提高代码可读性）
var (
	// PermUserRead 用户读权限
	PermUserRead = Permission{Resource: ResourceUser, Action: ActionRead}

	// PermUserWrite 用户写权限
	PermUserWrite = Permission{Resource: ResourceUser, Action: ActionWrite}

	// PermProfileRead 档案读权限
	PermProfileRead = Permission{Resource: ResourceProfile, Action: ActionRead}

	// PermProfileWrite 档案写权限
	PermProfileWrite = Permission{Resource: ResourceProfile, Action: ActionWrite}

	// PermConfigRead 配置读权限
	PermConfigRead = Permission{Resource: ResourceConfig, Action: ActionRead}

	// PermConfigWrite 配置写权限
	PermConfigWrite = Permission{Resource: ResourceConfig, Action: ActionWrite}

	// PermKeyRead 密钥读权限
	PermKeyRead = Permission{Resource: ResourceKey, Action: ActionRead}

	// PermKeyWrite 密钥写权限
	PermKeyWrite = Permission{Resource: ResourceKey, Action: ActionWrite}

	// PermSecretRead 秘密读权限
	PermSecretRead = Permission{Resource: ResourceSecret, Action: ActionRead}

	// PermSecretWrite 秘密写权限
	PermSecretWrite = Permission{Resource: ResourceSecret, Action: ActionWrite}

	// PermScopeGlobal 全局作用域权限
	PermScopeGlobal = Permission{Resource: ResourceScope, Action: ActionGlobal}

	// PermCasbinReload Casbin重新加载权限
	PermCasbinReload = Permission{Resource: ResourceCasbin, Action: ActionReload}
)

// Permission 权限结构体
// 用于表示一个完整的权限定义（资源+操作）
type Permission struct {
	Resource string // 资源类型
	Action   string // 操作类型
}
