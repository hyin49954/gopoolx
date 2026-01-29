package gopoolx

import "fmt"

// panicError 将任务中的 panic 包装为 error，方便统一处理。
// 目前主要用于 SubmitWithResult 中，将 panic 传递给 Future。
func panicError(r any) error {
	return fmt.Errorf("task panic: %v", r)
}
