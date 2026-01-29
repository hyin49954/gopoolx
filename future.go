package gopoolx

import "context"

// Future 表示一个异步计算结果的占位符。
// 典型用法为：提交带返回值的任务，得到 *Future[T]，随后在需要时调用 Get 获取结果。
type Future[T any] struct {
	// result 保存任务执行成功时的返回值
	result T
	// err 保存任务执行的错误（如果存在）
	err error
	// done 在任务完成（无论成功或失败）时关闭，用于通知等待方
	done chan struct{}
}

// newFuture 创建一个尚未完成的 Future。
func newFuture[T any]() *Future[T] {
	return &Future[T]{
		done: make(chan struct{}),
	}
}

// complete 在任务结束时由生产者调用，用于设置结果并通知所有等待方。
func (f *Future[T]) complete(res T, err error) {
	f.result = res
	f.err = err
	close(f.done)
}

// Get 阻塞等待任务完成或上下文结束。
//   - 若 ctx 先结束，则返回零值和 ctx.Err()
//   - 若任务先完成，则返回任务的 result 与 err
func (f *Future[T]) Get(ctx context.Context) (T, error) {
	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	case <-f.done:
		return f.result, f.err
	}
}
