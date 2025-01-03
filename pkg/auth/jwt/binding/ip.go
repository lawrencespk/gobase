package binding

import (
	"context"
	"net"
	"time"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/collector"
	"gobase/pkg/trace/jaeger"
)

// IPValidator IP绑定验证器
type IPValidator struct {
	store   Store
	logger  types.Logger
	metrics *collector.BusinessCollector
}

// NewIPValidator 创建IP绑定验证器
func NewIPValidator(store Store) (*IPValidator, error) {
	// 创建logger选项
	opts := logrus.DefaultOptions()
	logrus.WithLevel(types.InfoLevel)(opts)
	logrus.WithOutputPaths([]string{"stdout"})(opts)
	logrus.WithDevelopment(true)(opts)

	// 创建文件管理器选项
	fileOpts := logrus.FileOptions{
		BufferSize:    32 * 1024,   // 32KB 缓冲区
		FlushInterval: time.Second, // 1秒刷新间隔
		MaxOpenFiles:  100,         // 最大打开文件数
		DefaultPath:   "app.log",   // 默认日志文件路径
	}
	fileManager := logrus.NewFileManager(fileOpts)

	// 创建队列配置
	queueConfig := logrus.QueueConfig{
		MaxSize:         1000,             // 队列最大大小
		BatchSize:       100,              // 批处理大小
		Workers:         1,                // 工作协程数量
		FlushInterval:   time.Second,      // 刷新间隔
		RetryCount:      3,                // 重试次数
		RetryInterval:   time.Second,      // 重试间隔
		MaxBatchWait:    time.Second * 5,  // 最大批处理等待时间
		ShutdownTimeout: time.Second * 10, // 关闭超时时间
	}

	log, err := logrus.NewLogger(fileManager, queueConfig, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create logger")
	}

	// 使用已注册的 collector
	metrics := GetCollector()
	if metrics == nil {
		return nil, errors.NewError(codes.SystemError, "metrics collector not initialized", nil)
	}

	return &IPValidator{
		store:   store,
		logger:  log,
		metrics: metrics,
	}, nil
}

// ValidateIP 验证IP绑定
func (v *IPValidator) ValidateIP(ctx context.Context, claims jwt.Claims, currentIP string) error {
	span, spanCtx := jaeger.StartSpanFromContext(ctx, "binding.validate_ip")
	if span != nil {
		defer span.Finish()
		ctx = spanCtx
	}

	// 开始计时
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		v.metrics.ObserveOperation("ip_binding", duration, nil)
	}()

	// 验证IP格式
	if net.ParseIP(currentIP) == nil {
		err := errors.NewBindingInvalidError("invalid IP address", nil)
		v.metrics.ObserveOperation("ip_binding_validate", 0, err)
		return err
	}

	// 获取已绑定的IP
	boundIP, err := v.store.GetIPBinding(ctx, claims.GetDeviceID())
	if err != nil {
		if errors.Is(err, errors.NewError(codes.NotFound, "ip binding not found", nil)) {
			// 首次绑定
			if err := v.store.SaveIPBinding(ctx, claims.GetUserID(), claims.GetDeviceID(), currentIP); err != nil {
				v.metrics.ObserveOperation("ip_binding_save", 0, err)
				return err
			}
			v.metrics.ObserveOperation("ip_binding_first", 0, nil)
			return nil
		}
		v.metrics.ObserveOperation("ip_binding_get", 0, err)
		return err
	}

	// 验证IP是否匹配
	if boundIP != currentIP {
		// 创建日志字段
		fields := []types.Field{
			{Key: "device_id", Value: claims.GetDeviceID()},
			{Key: "bound_ip", Value: boundIP},
			{Key: "current_ip", Value: currentIP},
		}

		// 记录警告日志
		v.logger.Warn(ctx, "IP binding mismatch", fields...)

		err := errors.NewBindingMismatchError("IP address mismatch", nil)
		v.metrics.ObserveOperation("ip_binding_mismatch", 0, err)
		return err
	}

	v.metrics.ObserveOperation("ip_binding_match", 0, nil)
	return nil
}
