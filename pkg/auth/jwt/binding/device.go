package binding

import (
	"context"
	"time"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/collector"
	"gobase/pkg/trace/jaeger"
)

// DeviceValidator 设备绑定验证器
type DeviceValidator struct {
	store   Store
	logger  types.Logger
	metrics *collector.BusinessCollector
}

// NewDeviceValidator 创建设备绑定验证器
func NewDeviceValidator(store Store) (*DeviceValidator, error) {
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

	// 创建业务指标收集器
	metrics := collector.NewBusinessCollector("gobase_auth")
	if err := metrics.Register(); err != nil {
		return nil, errors.Wrap(err, "failed to register metrics collector")
	}

	return &DeviceValidator{
		store:   store,
		logger:  log,
		metrics: metrics,
	}, nil
}

// ValidateDevice 验证设备绑定
func (v *DeviceValidator) ValidateDevice(ctx context.Context, claims jwt.Claims, deviceInfo *DeviceInfo) error {
	span, spanCtx := jaeger.StartSpanFromContext(ctx, "binding.validate_device")
	if span != nil {
		defer span.Finish()
		ctx = spanCtx
	}

	// 开始计时
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		v.metrics.ObserveOperation("device_binding", duration, nil)
	}()

	// 验证设备信息
	if err := v.validateDeviceInfo(deviceInfo); err != nil {
		v.metrics.ObserveOperation("device_binding_validate", 0, err)
		return err
	}

	// 获取已绑定的设备
	boundDevice, err := v.store.GetDeviceBinding(ctx, claims.GetDeviceID())
	if err != nil {
		if errors.Is(err, errors.NewError(codes.NotFound, "device binding not found", nil)) {
			// 首次绑定
			if err := v.store.SaveDeviceBinding(ctx, claims.GetUserID(), claims.GetDeviceID(), deviceInfo); err != nil {
				v.metrics.ObserveOperation("device_binding_save", 0, err)
				return err
			}
			v.metrics.ObserveOperation("device_binding_first", 0, nil)
			return nil
		}
		v.metrics.ObserveOperation("device_binding_get", 0, err)
		return err
	}

	// 验证设备是否匹配
	if !v.isDeviceMatch(boundDevice, deviceInfo) {
		// 创建日志字段
		fields := []types.Field{
			{Key: "device_id", Value: claims.GetDeviceID()},
			{Key: "bound_device", Value: boundDevice.ID},
			{Key: "current_device", Value: deviceInfo.ID},
		}

		// 记录警告日志
		v.logger.Warn(ctx, "device binding mismatch", fields...)

		err := errors.NewBindingMismatchError("device mismatch", nil)
		v.metrics.ObserveOperation("device_binding_mismatch", 0, err)
		return err
	}

	v.metrics.ObserveOperation("device_binding_match", 0, nil)
	return nil
}

// validateDeviceInfo 验证设备信息
func (v *DeviceValidator) validateDeviceInfo(device *DeviceInfo) error {
	if device == nil {
		return errors.NewBindingInvalidError("device info is required", nil)
	}
	if device.ID == "" {
		return errors.NewBindingInvalidError("device ID is required", nil)
	}
	if device.Fingerprint == "" {
		return errors.NewBindingInvalidError("device fingerprint is required", nil)
	}
	return nil
}

// isDeviceMatch 检查设备是否匹配
func (v *DeviceValidator) isDeviceMatch(bound, current *DeviceInfo) bool {
	return bound.ID == current.ID &&
		bound.Fingerprint == current.Fingerprint
}
