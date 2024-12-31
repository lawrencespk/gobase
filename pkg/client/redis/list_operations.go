package redis

import "context"

// LPop 从列表左端弹出元素
func (c *client) LPop(ctx context.Context, key string) (string, error) {
	var result string
	err := c.withOperation(ctx, "LPop", func() error {
		var err error
		result, err = c.client.LPop(ctx, key).Result()
		return err
	})
	return result, err
}

// LPush 从列表左端推入元素
func (c *client) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	var result int64
	err := c.withOperation(ctx, "LPush", func() error {
		var err error
		result, err = c.client.LPush(ctx, key, values...).Result()
		return err
	})
	return result, err
}
