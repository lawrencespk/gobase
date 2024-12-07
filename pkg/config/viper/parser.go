package viper

import (
	"encoding/json"
	"fmt"

	"gobase/pkg/errors"
)

// Parser 配置解析器
type Parser struct {
	loader *Loader
}

// NewParser 创建配置解析器
func NewParser(loader *Loader) *Parser {
	return &Parser{
		loader: loader,
	}
}

// Parse 解析配置到结构体
func (p *Parser) Parse(key string, out interface{}) error {
	data := p.loader.Get(key)
	if data == nil {
		return errors.NewConfigNotFoundError(
			fmt.Sprintf("config key not found: %s", key),
			nil,
		)
	}

	// 将配置数据转换为JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.NewConfigInvalidError(
			fmt.Sprintf("failed to marshal config data: %s", key),
			err,
		)
	}

	// 解析JSON到结构体
	if err := json.Unmarshal(jsonData, out); err != nil {
		return errors.NewConfigInvalidError(
			fmt.Sprintf("failed to unmarshal config data: %s", key),
			err,
		)
	}

	return nil
}
