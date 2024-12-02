package unit

import (
	"testing"
	"time"

	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
)

func TestValidateOptions(t *testing.T) {
	t.Run("Valid Basic Options", func(t *testing.T) {
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{"stdout", "/var/log/app.log"},
		}
		if err := logrus.ValidateOptions(opts); err != nil {
			t.Errorf("Expected valid options, got error: %v", err)
		}
	})

	t.Run("Invalid Log Level", func(t *testing.T) {
		opts := &logrus.Options{
			Level:       types.Level(999),
			OutputPaths: []string{"stdout"},
		}
		if err := logrus.ValidateOptions(opts); err == nil {
			t.Error("Expected error for invalid log level, got nil")
		}
	})

	t.Run("Empty Output Paths", func(t *testing.T) {
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{},
		}
		if err := logrus.ValidateOptions(opts); err == nil {
			t.Error("Expected error for empty output paths, got nil")
		}
	})

	t.Run("Valid Compress Config", func(t *testing.T) {
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{"stdout"},
			CompressConfig: logrus.CompressConfig{
				Enable:       true,
				Algorithm:    "gzip",
				Level:        6,
				DeleteSource: true,
			},
		}
		if err := logrus.ValidateOptions(opts); err != nil {
			t.Errorf("Expected valid compress config, got error: %v", err)
		}
	})

	t.Run("Invalid Compress Algorithm", func(t *testing.T) {
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{"stdout"},
			CompressConfig: logrus.CompressConfig{
				Enable:    true,
				Algorithm: "invalid_algo",
			},
		}
		if err := logrus.ValidateOptions(opts); err == nil {
			t.Error("Expected error for invalid compress algorithm, got nil")
		}
	})

	t.Run("Valid Cleanup Config", func(t *testing.T) {
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{"stdout"},
			CleanupConfig: logrus.CleanupConfig{
				Enable:     true,
				MaxAge:     7,
				MaxBackups: 5,
				Interval:   time.Hour * 24,
			},
		}
		if err := logrus.ValidateOptions(opts); err != nil {
			t.Errorf("Expected valid cleanup config, got error: %v", err)
		}
	})

	t.Run("Invalid Cleanup Values", func(t *testing.T) {
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{"stdout"},
			CleanupConfig: logrus.CleanupConfig{
				Enable:     true,
				MaxAge:     -1,
				MaxBackups: -1,
				Interval:   -time.Hour,
			},
		}
		if err := logrus.ValidateOptions(opts); err == nil {
			t.Error("Expected error for invalid cleanup values, got nil")
		}
	})

	t.Run("Valid Async Config", func(t *testing.T) {
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{"stdout"},
			AsyncConfig: logrus.AsyncConfig{
				Enable:        true,
				BufferSize:    1024,
				FlushInterval: time.Second,
				BlockOnFull:   true,
			},
		}
		if err := logrus.ValidateOptions(opts); err != nil {
			t.Errorf("Expected valid async config, got error: %v", err)
		}
	})

	t.Run("Invalid Buffer Size", func(t *testing.T) {
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{"stdout"},
			AsyncConfig: logrus.AsyncConfig{
				Enable:     true,
				BufferSize: -1,
			},
		}
		if err := logrus.ValidateOptions(opts); err == nil {
			t.Error("Expected error for invalid buffer size, got nil")
		}
	})

	t.Run("Valid Recovery Config", func(t *testing.T) {
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{"stdout"},
			RecoveryConfig: logrus.RecoveryConfig{
				Enable:           true,
				MaxRetries:       3,
				RetryInterval:    time.Second,
				EnableStackTrace: true,
				MaxStackSize:     4096,
			},
		}
		if err := logrus.ValidateOptions(opts); err != nil {
			t.Errorf("Expected valid recovery config, got error: %v", err)
		}
	})

	t.Run("Invalid Recovery Values", func(t *testing.T) {
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{"stdout"},
			RecoveryConfig: logrus.RecoveryConfig{
				Enable:           true,
				MaxRetries:       -1,
				RetryInterval:    -time.Second,
				EnableStackTrace: true,
				MaxStackSize:     -1,
			},
		}
		if err := logrus.ValidateOptions(opts); err == nil {
			t.Error("Expected error for invalid recovery values, got nil")
		}
	})

	t.Run("Config Conflicts", func(t *testing.T) {
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{"stdout"},
			AsyncConfig: logrus.AsyncConfig{
				Enable:      true,
				BlockOnFull: true,
			},
			RecoveryConfig: logrus.RecoveryConfig{
				Enable:     true,
				MaxRetries: 3,
			},
		}
		if err := logrus.ValidateOptions(opts); err == nil {
			t.Error("Expected error for conflicting async and recovery configs, got nil")
		}
	})
}

func TestValidateLevel(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	invalidLevels := []string{"", "trace", "warning", "critical"}

	for _, level := range validLevels {
		t.Run(level, func(t *testing.T) {
			if !logrus.ValidateLevel(level) {
				t.Errorf("Expected %s to be a valid log level", level)
			}
		})
	}

	for _, level := range invalidLevels {
		t.Run(level, func(t *testing.T) {
			if logrus.ValidateLevel(level) {
				t.Errorf("Expected %s to be an invalid log level", level)
			}
		})
	}
}
