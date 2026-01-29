package gopoolx

import "sync"

// ErrorCollector 用于在并发环境下收集任务执行错误。
// 通过内部互斥锁保证在多 goroutine 下安全地写入和读取错误切片。
type ErrorCollector struct {
	mu sync.Mutex
	// errs 存放所有收集到的错误
	errs []error
}

// Add 将一个错误加入收集器。
// 若 err 为 nil，会被直接忽略。
func (e *ErrorCollector) Add(err error) {
	if err == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.errs = append(e.errs, err)
}

// Errors 返回一个包含已收集错误的切片副本。
// 返回副本是为了避免调用方修改内部状态。
func (e *ErrorCollector) Errors() []error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]error(nil), e.errs...)
}
