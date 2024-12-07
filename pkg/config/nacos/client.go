package nacos

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"

	"gobase/pkg/errors"
)

// Client Nacos客户端
type Client struct {
	config *Config
	client config_client.IConfigClient
	mutex  sync.RWMutex
	// 配置监听回调
	watchers map[string][]func(string, string)
}

// NewClient 创建Nacos客户端
func NewClient(config *Config) (*Client, error) {
	// 创建客户端配置
	clientConfig := config.ToClientConfig()
	serverConfigs := config.ToServerConfig()

	// 创建客户端
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, errors.NewNacosConnectError(
			fmt.Sprintf("failed to create nacos client: %v", err),
			err,
		)
	}

	return &Client{
		config:   config,
		client:   client,
		watchers: make(map[string][]func(string, string)),
	}, nil
}

// GetConfig 获取配置
func (c *Client) GetConfig(dataID, group string) (string, error) {
	content, err := c.client.GetConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
	})
	if err != nil {
		return "", errors.NewNacosConfigError(
			fmt.Sprintf("failed to get config, dataID: %s, group: %s", dataID, group),
			err,
		)
	}
	return content, nil
}

// PublishConfig 发布配置
func (c *Client) PublishConfig(dataID, group, content string) error {
	success, err := c.client.PublishConfig(vo.ConfigParam{
		DataId:  dataID,
		Group:   group,
		Content: content,
	})
	if err != nil {
		return errors.NewNacosPublishError(
			fmt.Sprintf("failed to publish config, dataID: %s, group: %s", dataID, group),
			err,
		)
	}
	if !success {
		return errors.NewNacosPublishError(
			fmt.Sprintf("publish config failed, dataID: %s, group: %s", dataID, group),
			nil,
		)
	}
	return nil
}

// WatchConfig 监听配置变化
func (c *Client) WatchConfig(dataID, group string, callback func(string, string)) error {
	c.mutex.Lock()
	if _, exists := c.watchers[dataID]; !exists {
		c.watchers[dataID] = make([]func(string, string), 0)
	}
	c.watchers[dataID] = append(c.watchers[dataID], callback)
	c.mutex.Unlock()

	err := c.client.ListenConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			c.mutex.RLock()
			defer c.mutex.RUnlock()
			if callbacks, ok := c.watchers[dataId]; ok {
				for _, cb := range callbacks {
					cb(dataId, data)
				}
			}
		},
	})
	if err != nil {
		return errors.NewNacosWatchError(
			fmt.Sprintf("failed to watch config, dataID: %s, group: %s", dataID, group),
			err,
		)
	}
	return nil
}

// Close 关闭客户端
func (c *Client) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 创建一个 channel 来接收关闭完成信号
	done := make(chan struct{}, 1)

	go func() {
		c.client.CloseClient()
		done <- struct{}{}
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return errors.NewNacosTimeoutError("close nacos client timeout", ctx.Err())
	}
}
