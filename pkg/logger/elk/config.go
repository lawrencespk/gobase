package elk

import (
	"gobase/pkg/config"
)

// ElkConfig 定义与 Elasticsearch 连接的配置
type ElkConfig struct {
	Addresses []string
	Username  string
	Password  string
	Index     string
	Timeout   int
}

// DefaultElkConfig 返回默认配置
func DefaultElkConfig() *ElkConfig {
	conf := config.GetConfig()
	if conf == nil || len(conf.ELK.Addresses) == 0 {
		return &ElkConfig{
			Addresses: []string{"http://localhost:9200"},
			Username:  "",
			Password:  "",
			Index:     "default-index",
			Timeout:   30,
		}
	}

	return &ElkConfig{
		Addresses: conf.ELK.Addresses,
		Username:  conf.ELK.Username,
		Password:  conf.ELK.Password,
		Index:     conf.ELK.Index,
		Timeout:   conf.ELK.Timeout,
	}
}
