package redis

import (
	"bytes"
	"context"
	"fmt"
	"gobase/pkg/logger/types"
	"io"
	"sync"
	"time"

	"gobase/pkg/trace/jaeger"

	"github.com/go-redis/redis/v8"
)

// 添加命令对象池
var cmdPool = sync.Pool{
	New: func() interface{} {
		slice := make([]redis.Cmder, 0, 10)
		return &slice // 返回切片的指针
	},
}

// TxPipeline 创建一个事务管道
func (c *client) TxPipeline() Pipeline {
	var metrics *pipelineMetrics
	if c.options != nil && c.options.EnableMetrics {
		metrics = newPipelineMetrics(c.options.MetricsNamespace)
	}

	return &redisPipeline{
		pipeline: c.client.Pipeline(),
		tracer:   c.tracer,
		logger:   c.logger,
		metrics:  metrics,
		cmdBuf:   &bytes.Buffer{},
	}
}

// redisPipeline Redis管道实现
type redisPipeline struct {
	pipeline redis.Pipeliner
	tracer   *jaeger.Provider
	cmds     []redis.Cmder
	logger   types.Logger
	metrics  *pipelineMetrics
	cmdBuf   *bytes.Buffer // 添加命令缓冲区
	commands []*Command
}

// withPipelineOperation Pipeline操作的统一包装函数
func (p *redisPipeline) withPipelineOperation(ctx context.Context, operation string, cmd redis.Cmder) {
	// 使用对象池获取命令列表
	if p.cmds == nil {
		slicePtr := cmdPool.Get().(*[]redis.Cmder)
		p.cmds = *slicePtr // 解引用获取实际的切片
	}

	// 记录指标
	if p.metrics != nil {
		p.metrics.commandsTotal.WithLabelValues(operation).Inc()
	}

	// 记录日志
	p.logger.WithFields(
		types.Field{Key: "operation", Value: operation},
		types.Field{Key: "command_type", Value: cmd.Name()},
		types.Field{Key: "args", Value: cmd.Args()},
		types.Field{Key: "total_commands", Value: len(p.cmds) + 1},
	).Debug(ctx, "adding command to pipeline")

	// 添加命令到队列
	p.cmds = append(p.cmds, cmd)
}

// Exec 实现 Pipeline.Exec 方法
func (p *redisPipeline) Exec(ctx context.Context) ([]Cmder, error) {
	defer func() {
		// 归还命令列表到对象池
		if p.cmds != nil {
			p.cmds = p.cmds[:0]
			slicePtr := &p.cmds
			cmdPool.Put(slicePtr)
			p.cmds = nil
		}
	}()

	var span *jaeger.Span
	if p.tracer != nil {
		var newCtx context.Context
		span, newCtx = startSpan(ctx, p.tracer, "redis.Pipeline.Exec")
		ctx = newCtx
		defer span.Finish()
	}

	// 开始时间
	start := time.Now()

	// 确保清理命令列表
	defer func() {
		p.cmds = nil
	}()

	// 记录初始状态
	if p.logger != nil {
		p.logger.WithFields(
			types.Field{Key: "total_commands", Value: len(p.cmds)},
			types.Field{Key: "event", Value: "pipeline_exec_start"},
		).Debug(ctx, "starting pipeline execution")
	}

	// 如果没有命令要执行，直接返回
	if len(p.cmds) == 0 {
		if p.logger != nil {
			p.logger.Debug(ctx, "no commands to execute")
		}
		return nil, nil
	}

	// 执行管道命令
	_, err := p.pipeline.Exec(ctx)
	if err != nil {
		// 记录错误指标
		if p.metrics != nil {
			p.metrics.errorTotal.WithLabelValues("exec", err.Error()).Inc()
			p.metrics.executionLatency.WithLabelValues("error").Observe(time.Since(start).Seconds())
		}

		if p.logger != nil {
			p.logger.WithFields(
				types.Field{Key: "error", Value: err},
				types.Field{Key: "duration", Value: time.Since(start)},
				types.Field{Key: "event", Value: "pipeline_exec_error"},
			).Error(ctx, "pipeline execution failed")
		}

		return nil, handleRedisError(err, errPipelineFailed)
	}

	// 检查每个命令的执行结果
	result := make([]Cmder, len(p.cmds))
	for i, cmd := range p.cmds {
		if err := cmd.Err(); err != nil && err != redis.Nil {
			// 记录错误指标
			if p.metrics != nil {
				p.metrics.errorTotal.WithLabelValues(cmd.Name(), err.Error()).Inc()
				p.metrics.executionLatency.WithLabelValues("error").Observe(time.Since(start).Seconds())
			}

			if p.logger != nil {
				p.logger.WithFields(
					types.Field{Key: "command", Value: cmd.Name()},
					types.Field{Key: "error", Value: err},
					types.Field{Key: "duration", Value: time.Since(start)},
					types.Field{Key: "event", Value: "command_error"},
				).Error(ctx, "pipeline command failed")
			}

			return nil, handleRedisError(err, fmt.Sprintf("command %s failed", cmd.Name()))
		}
		result[i] = cmd
	}

	// 记录成功指标
	if p.metrics != nil {
		p.metrics.executionLatency.WithLabelValues("success").Observe(time.Since(start).Seconds())
	}

	// 记录成功日志
	if p.logger != nil {
		p.logger.WithFields(
			types.Field{Key: "duration", Value: time.Since(start)},
			types.Field{Key: "total_commands", Value: len(result)},
			types.Field{Key: "event", Value: "pipeline_exec_success"},
		).Debug(ctx, "pipeline execution completed successfully")
	}

	return result, nil
}

// Close 实现 Pipeline.Close 方法
func (p *redisPipeline) Close() error {
	// 记录日志
	p.logger.Debug(context.Background(), "closing pipeline")

	// 先丢弃所有待执行的命令
	if err := p.pipeline.Discard(); err != nil {
		p.logger.WithError(err).Error(context.Background(), "failed to discard pipeline")
		return handleRedisError(err, "failed to discard pipeline")
	}

	// 关闭底层 pipeline
	if closer, ok := p.pipeline.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			p.logger.WithError(err).Error(context.Background(), "failed to close pipeline")
			return handleRedisError(err, "failed to close pipeline")
		}
	}

	// 清空命令列表
	p.cmds = nil
	return nil
}

// Del 删除一个或多个键
func (p *redisPipeline) Del(ctx context.Context, keys ...string) (int64, error) {
	p.withPipelineOperation(ctx, "Del", p.pipeline.Del(ctx, keys...))
	return 0, nil
}

// Get 获取键值
func (p *redisPipeline) Get(ctx context.Context, key string) (string, error) {
	p.withPipelineOperation(ctx, "Get", p.pipeline.Get(ctx, key))
	return "", nil
}

// HDel 删除哈希表中的字段
func (p *redisPipeline) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	p.withPipelineOperation(ctx, "HDel", p.pipeline.HDel(ctx, key, fields...))
	return 0, nil
}

// HGet 获取哈希表中的字段值
func (p *redisPipeline) HGet(ctx context.Context, key, field string) (string, error) {
	p.withPipelineOperation(ctx, "HGet", p.pipeline.HGet(ctx, key, field))
	return "", nil
}

// HSet 设置哈希表字段值
func (p *redisPipeline) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	p.withPipelineOperation(ctx, "HSet", p.pipeline.HSet(ctx, key, values...))
	return 0, nil
}

// Set 设置键值
func (p *redisPipeline) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	p.withPipelineOperation(ctx, "Set", p.pipeline.Set(ctx, key, value, expiration))
	return nil
}

// SAdd 向集合添加元素
func (p *redisPipeline) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	p.withPipelineOperation(ctx, "SAdd", p.pipeline.SAdd(ctx, key, members...))
	return 0, nil
}

// SRem 从集合中移除元素
func (p *redisPipeline) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	p.withPipelineOperation(ctx, "SRem", p.pipeline.SRem(ctx, key, members...))
	return 0, nil
}

// ZAdd 向有序集合添加元素
func (p *redisPipeline) ZAdd(ctx context.Context, key string, members ...*Z) (int64, error) {
	redisMembers := make([]*redis.Z, len(members))
	for i, member := range members {
		redisMembers[i] = &redis.Z{
			Score:  member.Score,
			Member: member.Member,
		}
	}
	p.withPipelineOperation(ctx, "ZAdd", p.pipeline.ZAdd(ctx, key, redisMembers...))
	return 0, nil
}

// ZRem 从有序集合中移除元素
func (p *redisPipeline) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	p.withPipelineOperation(ctx, "ZRem", p.pipeline.ZRem(ctx, key, members...))
	return 0, nil
}

// Command 改为使用指针类型
type Command struct {
	Name   string
	Args   []interface{}
	Result interface{}
}

// 修改 redisPipeline 结构体的方法
func (p *redisPipeline) addCommand(cmd *Command) {
	if p.commands == nil {
		p.commands = make([]*Command, 0)
	}
	p.commands = append(p.commands, cmd)
}

// 修改创建命令的方法为 redisPipeline 的方法
func (p *redisPipeline) newCommand(name string, args []interface{}) *Command {
	return &Command{
		Name: name,
		Args: args,
	}
}

// 修改 Send 方法为 redisPipeline 的方法
func (p *redisPipeline) Send(name string, args ...interface{}) {
	cmd := p.newCommand(name, args)
	p.addCommand(cmd)
}
