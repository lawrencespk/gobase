package session

import (
	"context"
	"os"
	"sync"
	"time"

	"gobase/pkg/auth/jwt/security"
	"gobase/pkg/errors"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metric"
	"gobase/pkg/trace/jaeger"

	"github.com/google/uuid"
)

var (
	// 会话指标
	sessionCreated = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "session",
		Name:      "created_total",
		Help:      "Total number of sessions created",
	})

	sessionDeleted = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "session",
		Name:      "deleted_total",
		Help:      "Total number of sessions deleted",
	})

	sessionUpdated = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "session",
		Name:      "updated_total",
		Help:      "Total number of sessions updated",
	})

	sessionExpired = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "session",
		Name:      "expired_total",
		Help:      "Total number of sessions expired",
	})

	activeSessions = metric.NewGauge(metric.GaugeOpts{
		Namespace: "gobase",
		Subsystem: "session",
		Name:      "active_total",
		Help:      "Current number of active sessions",
	})
)

func init() {
	// 注册所有指标
	for _, m := range []interface{ Register() error }{
		sessionCreated,
		sessionDeleted,
		sessionUpdated,
		sessionExpired,
		activeSessions,
	} {
		if err := m.Register(); err != nil {
			// 忽略已注册错误
			if err.Error() != "duplicate metrics collector registration attempted" {
				panic(err)
			}
		}
	}
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	UserID     string    // 用户ID
	TokenInfo  TokenInfo // 令牌信息
	DeviceInfo Device    // 设备信息
}

// TokenInfo 令牌信息
type TokenInfo struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Device 设备信息
type Device struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Platform string `json:"platform"`
}

// SessionManager 会话管理器
type SessionManager struct {
	store  Store
	policy *security.Policy
	logger types.Logger
	mutex  sync.RWMutex
}

// NewManager 创建会话管理器
func NewManager(store Store, policy *security.Policy) (*SessionManager, error) {
	log, err := logger.NewLogger(
		logger.WithOutput(os.Stdout),
		logger.WithLevel(types.InfoLevel),
	)
	if err != nil {
		return nil, err
	}

	m := &SessionManager{
		store:  store,
		policy: policy,
		logger: log,
	}

	return m, nil
}

// CreateSession 创建会话
func (m *SessionManager) CreateSession(ctx context.Context, req CreateSessionRequest) (*Session, error) {
	span, ctx := jaeger.StartSpanFromContext(ctx, "session.CreateSession")
	defer span.Finish()

	// 检查活跃会话数量
	if m.policy.MaxActiveSessions > 0 {
		count, err := m.store.Count(ctx, req.UserID)
		if err != nil {
			return nil, err
		}
		if count >= m.policy.MaxActiveSessions {
			return nil, errors.NewPolicyViolationError("max active sessions exceeded", nil)
		}
	}

	// 创建会话
	now := time.Now()
	session := &Session{
		UserID:    req.UserID,
		TokenID:   uuid.New().String(),
		ExpiresAt: req.TokenInfo.ExpiresAt,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata: map[string]interface{}{
			"device_id":       req.DeviceInfo.ID,
			"device_type":     req.DeviceInfo.Type,
			"device_name":     req.DeviceInfo.Name,
			"device_platform": req.DeviceInfo.Platform,
			"access_token":    req.TokenInfo.AccessToken,
			"refresh_token":   req.TokenInfo.RefreshToken,
		},
	}

	// 保存会话
	if err := m.store.Save(ctx, session); err != nil {
		return nil, err
	}

	// 更新指标
	sessionCreated.Inc()
	activeSessions.Inc()

	return session, nil
}
