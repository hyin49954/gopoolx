package gopoolx

import (
	"fmt"
	"sync"
)

// TaskError 表示某个任务执行失败时记录的错误。
type TaskError struct {
	TaskID TaskID
	Err    error
	// Data 是任务失败时附带的额外返回值。
	Data any
}

func (e TaskError) Error() string {
	return fmt.Sprintf("task %s: %v", e.TaskID, e.Err)
}

func (e TaskError) Unwrap() error {
	return e.Err
}

// ErrorCollector 用于在并发环境下收集任务执行错误。
// 通过内部互斥锁保证在多 goroutine 下安全地写入和读取错误切片。
type ErrorCollector struct {
	mu sync.Mutex
	// errs 存放所有收集到的错误
	errs []TaskError
	// byTask 支持按任务 ID 快速查询错误
	byTask map[TaskID]TaskError
}

// Add 将一个错误加入收集器。
// 若 err 为 nil，会被直接忽略。
func (e *ErrorCollector) Add(taskID TaskID, err error, data any) {
	if err == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.byTask == nil {
		e.byTask = make(map[TaskID]TaskError)
	}
	taskErr := TaskError{
		TaskID: taskID,
		Err:    err,
		Data:   data,
	}
	e.errs = append(e.errs, taskErr)
	e.byTask[taskID] = taskErr
}

// Errors 返回一个包含已收集错误的切片副本。
// 返回副本是为了避免调用方修改内部状态。
func (e *ErrorCollector) Errors() []TaskError {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]TaskError(nil), e.errs...)
}

// DrainErrors 原子地取出当前所有错误并清空收集器。
// 调用后，已取出的错误不会再被 Error 或 Errors 返回。
func (e *ErrorCollector) DrainErrors() []TaskError {
	e.mu.Lock()
	defer e.mu.Unlock()

	errs := e.errs
	e.errs = nil
	e.byTask = nil
	return errs
}

// Error 返回指定任务 ID 对应的错误记录。
func (e *ErrorCollector) Error(taskID TaskID) (TaskError, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	taskErr, ok := e.byTask[taskID]
	return taskErr, ok
}
