package errors

import (
	"gobase/pkg/errors/codes"
)

// 任务处理错误码 (2600-2699)

// NewTaskError 创建任务错误
func NewTaskError(message string, cause error) error {
	return NewError(codes.TaskError, message, cause)
}

// NewJobError 创建作业错误
func NewJobError(message string, cause error) error {
	return NewError(codes.JobError, message, cause)
}

// NewScheduleError 创建调度错误
func NewScheduleError(message string, cause error) error {
	return NewError(codes.ScheduleError, message, cause)
}

// NewExecutionError 创建执行错误
func NewExecutionError(message string, cause error) error {
	return NewError(codes.ExecutionError, message, cause)
}
