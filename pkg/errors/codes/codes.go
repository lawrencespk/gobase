package codes

const (
	// 系统级错误码 (1000-1099)
	SystemError        = "1000" // 系统内部错误
	ConfigError        = "1001" // 配置错误
	NetworkError       = "1002" // 网络错误
	DatabaseError      = "1003" // 数据库错误
	CacheError         = "1004" // 缓存错误
	TimeoutError       = "1005" // 超时错误
	ValidationError    = "1006" // 验证错误
	SerializationError = "1007" // 序列化错误
	ThirdPartyError    = "1008" // 第三方服务错误
	InitializeError    = "1009" // 初始化错误
	ShutdownError      = "1010" // 关闭错误
	MemoryError        = "1011" // 内存错误
	DiskError          = "1012" // 磁盘错误
	ResourceExhausted  = "1013" // 资源耗尽

	// 中间件错误码 (1100-1199)
	MiddlewareError = "1100" // 中间件错误
	AuthError       = "1101" // 认证中间件错误
	RateLimitError  = "1102" // 限流中间件错误
	TimeoutMWError  = "1103" // 超时中间件错误
	CORSError       = "1104" // 跨域中间件错误
	TracingError    = "1105" // 追踪中间件错误
	MetricsError    = "1106" // 指标中间件错误
	LoggingError    = "1107" // 日志中间件错误

	// 数据操作错误码 (1200-1299)
	DataAccessError   = "1200" // 数据访问错误
	DataCreateError   = "1201" // 数据创建错误
	DataUpdateError   = "1202" // 数据更新错误
	DataDeleteError   = "1203" // 数据删除错误
	DataQueryError    = "1204" // 数据查询错误
	DataConvertError  = "1205" // 数据转换错误
	DataValidateError = "1206" // 数据验证错误
	DataCorruptError  = "1207" // 数据损坏错误

	// 业务通用错误码 (2000-2099)
	InvalidParams      = "2000" // 无效参数
	Unauthorized       = "2001" // 未授权
	Forbidden          = "2002" // 禁止访问
	NotFound           = "2003" // 资源不存在
	AlreadyExists      = "2004" // 资源已存在
	BadRequest         = "2005" // 错误的请求
	TooManyRequests    = "2006" // 请求过多
	OperationFailed    = "2007" // 操作失败
	DataConflict       = "2008" // 数据冲突
	ServiceUnavailable = "2009" // 服务不可用
	RequestTimeout     = "2010" // 请求超时
	InvalidToken       = "2011" // 无效令牌
	TokenExpired       = "2012" // 令牌过期
	InvalidSignature   = "2013" // 无效签名

	// 用户相关错误码 (2100-2199)
	UserNotFound      = "2100" // 用户不存在
	UserDisabled      = "2101" // 用户已禁用
	UserLocked        = "2102" // 用户已锁定
	InvalidPassword   = "2103" // 密码错误
	PasswordExpired   = "2104" // 密码过期
	InvalidUsername   = "2105" // 用户名无效
	DuplicateUsername = "2106" // 用户名重复

	// 权限相关错误码 (2200-2299)
	NoPermission     = "2200" // 无权限
	RoleNotFound     = "2201" // 角色不存在
	InvalidRole      = "2202" // 角色无效
	PermissionDenied = "2203" // 权限被拒绝

	// 文件操作错误码 (2300-2399)
	FileNotFound       = "2300" // 文件不存在
	FileUploadError    = "2301" // 文件上传错误
	FileDownloadError  = "2302" // 文件下载错误
	InvalidFileType    = "2303" // 文件类型无效
	FileTooLarge       = "2304" // 文件太大
	FileDeleteError    = "2305" // 文件删除错误
	FileOperationError = "2306" // 文件操作错误
	FileOpenError      = "2307" // 文件打开错误
	FileWriteError     = "2308" // 文件写入错误
	FileReadError      = "2309" // 文件读取错误
	FileCloseError     = "2310" // 文件关闭错误
	FileFlushError     = "2311" // 文件刷新错误

	// 通信相关错误码 (2400-2499)
	MessageError    = "2400" // 消息错误
	InvalidFormat   = "2401" // 格式无效
	InvalidProtocol = "2402" // 协议无效
	EncryptionError = "2403" // 加密错误
	DecryptionError = "2404" // 解密错误

	// 第三方服务错误码 (2500-2599)
	APIError         = "2500" // API错误
	ServiceError     = "2501" // 服务错误
	IntegrationError = "2502" // 集成错误
	DependencyError  = "2503" // 依赖错误

	// ELK相关错误码 (2510-2519)
	ELKConnectionError = "2510" // ELK连接错误
	ELKIndexError      = "2511" // ELK索引错误
	ELKQueryError      = "2512" // ELK查询错误
	ELKBulkError       = "2513" // ELK批量操作错误
	ELKConfigError     = "2514" // ELK配置错误
	ELKTimeoutError    = "2515" // ELK超时错误

	// 任务处理错误码 (2600-2699)
	TaskError      = "2600" // 任务错误
	JobError       = "2601" // 作业错误
	ScheduleError  = "2602" // 调度错误
	ExecutionError = "2603" // 执行错误

	// 缓存相关错误码 (2700-2799)
	CacheMissError    = "2700" // 缓存未命中
	CacheExpiredError = "2701" // 缓存已过期
	CacheFullError    = "2702" // 缓存已满

	// 数据库相关错误码 (2800-2899)
	DBConnError        = "2800" // 数据库连接错误
	DBQueryError       = "2801" // 数据库查询错误
	DBTransactionError = "2802" // 数据库事务错误
	DBDeadlockError    = "2803" // 数据库死锁错误

	// 配置相关错误码 (2900-2999)
	ConfigNotFound    = "2900" // 配置不存在
	ConfigInvalid     = "2901" // 配置无效
	ConfigUpdateError = "2902" // 配置更新错误
)
