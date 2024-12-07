package viper

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"gobase/pkg/errors"
)

// Loader 配置加载器
type Loader struct {
	viper      *viper.Viper
	configFile string
	configType string
	envPrefix  string
	enableEnv  bool
}

// NewLoader 创建配置加载器
func NewLoader(configFile string, enableEnv bool, envPrefix string) *Loader {
	v := viper.New()

	// 设置配置文件
	if configFile != "" {
		v.SetConfigFile(configFile)
		v.SetConfigType(strings.TrimPrefix(filepath.Ext(configFile), "."))
	}

	// 设置环境变量
	if enableEnv {
		v.AutomaticEnv()
		if envPrefix != "" {
			v.SetEnvPrefix(envPrefix)
		}
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	}

	return &Loader{
		viper:      v,
		configFile: configFile,
		configType: strings.TrimPrefix(filepath.Ext(configFile), "."),
		enableEnv:  enableEnv,
		envPrefix:  envPrefix,
	}
}

// Load 加载配置
func (l *Loader) Load() error {
	if err := l.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return errors.NewConfigNotFoundError(
				fmt.Sprintf("config file not found: %s", l.configFile),
				err,
			)
		}
		return errors.NewConfigInvalidError(
			fmt.Sprintf("failed to read config file: %s", l.configFile),
			err,
		)
	}
	return nil
}

// Get 获取配置值
func (l *Loader) Get(key string) interface{} {
	return l.viper.Get(key)
}

// GetString 获取字符串配置
func (l *Loader) GetString(key string) string {
	return l.viper.GetString(key)
}

// GetInt 获取整数配置
func (l *Loader) GetInt(key string) int {
	return l.viper.GetInt(key)
}

// GetBool 获取布尔配置
func (l *Loader) GetBool(key string) bool {
	return l.viper.GetBool(key)
}

// GetFloat64 获取浮点数配置
func (l *Loader) GetFloat64(key string) float64 {
	return l.viper.GetFloat64(key)
}

// GetStringSlice 获取字符串切片配置
func (l *Loader) GetStringSlice(key string) []string {
	return l.viper.GetStringSlice(key)
}

// GetStringMap 获取字符串映射配置
func (l *Loader) GetStringMap(key string) map[string]interface{} {
	return l.viper.GetStringMap(key)
}

// GetStringMapString 获取字符串-字符串映射配置
func (l *Loader) GetStringMapString(key string) map[string]string {
	return l.viper.GetStringMapString(key)
}

// Set 设置配置值
func (l *Loader) Set(key string, value interface{}) {
	l.viper.Set(key, value)
}

// IsSet 检查配置是否存在
func (l *Loader) IsSet(key string) bool {
	return l.viper.IsSet(key)
}

// Watch 监听配置变化
func (l *Loader) Watch(key string, callback func(key string, value interface{})) error {
	l.viper.OnConfigChange(func(in fsnotify.Event) {
		if in.Op&fsnotify.Write == fsnotify.Write {
			callback(key, l.Get(key))
		}
	})
	l.viper.WatchConfig()
	return nil
}

// GetViper 获取viper实例
func (l *Loader) GetViper() *viper.Viper {
	return l.viper
}

// GetDuration 获取时间间隔配置
func (l *Loader) GetDuration(key string) time.Duration {
	return l.viper.GetDuration(key)
}
