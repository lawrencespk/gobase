package viper

import (
	"encoding/json"

	"gobase/pkg/config/types"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

// 确保 Parser 实现了 types.Parser 接口
var _ types.Parser = (*Parser)(nil)

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
		return errors.NewError(codes.ConfigNotFound, "config key not found", nil)
	}

	// 将配置数据转换为JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.NewError(codes.ConfigInvalid, "invalid config data", err)
	}

	// 解析JSON到结构体
	if err := json.Unmarshal(jsonData, out); err != nil {
		return errors.NewError(codes.ConfigInvalid, "invalid config format", err)
	}

	return nil
}
