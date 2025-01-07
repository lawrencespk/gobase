package context

// contextKey 定义上下文键类型
type contextKey string

const (
	// ClaimsKey 存储JWT Claims的上下文键
	ClaimsKey contextKey = "jwt_claims"
	// TokenKey 存储JWT Token的上下文键
	TokenKey contextKey = "jwt_token"
	// TokenTypeKey 存储Token类型的上下文键
	TokenTypeKey contextKey = "jwt_token_type"
	// UserIDKey 存储用户ID的上下文键
	UserIDKey contextKey = "jwt_user_id"
	// UserNameKey 存储用户名的上下文键
	UserNameKey contextKey = "jwt_user_name"
	// RolesKey 存储角色列表的上下文键
	RolesKey contextKey = "jwt_roles"
	// PermissionsKey 存储权限列表的上下文键
	PermissionsKey contextKey = "jwt_permissions"
	// DeviceIDKey 存储设备ID的上下文键
	DeviceIDKey contextKey = "jwt_device_id"
	// IPAddressKey 存储IP地址的上下文键
	IPAddressKey contextKey = "jwt_ip_address"
)
