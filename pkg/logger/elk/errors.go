package elk

import (
	"gobase/pkg/errors"
)

// HandleELKError 处理 ELK 相关错误
func HandleELKError(err error) {
	// 根据错误类型进行不同的处理
	if errors.Is(err, errors.NewELKConnectionError("", nil)) {
		// 处理连接错误
	} else if errors.Is(err, errors.NewELKIndexError("", nil)) {
		// 处理索引错误
	} else if errors.Is(err, errors.NewELKQueryError("", nil)) {
		// 处理查询错误
	} else if errors.Is(err, errors.NewELKBulkError("", nil)) {
		// 处理批量操作错误
	}
}
