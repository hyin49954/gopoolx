package gopoolx

import "context"

// SubmitWithResult 提交一个带返回值的任务到指定的 Pool 中，并返回一个 Future 用于异步获取结果。
// 说明：
//   - fn 会在池中的 worker goroutine 中执行
//   - 若 fn 正常返回，其结果与错误会写入 Future
//   - 若 fn 发生 panic，会被捕获并转换为 error 返回到 Future
func SubmitWithResult[T any](
	pool *Pool,
	fn func(ctx context.Context) (T, error),
) *Future[T] {

	future := newFuture[T]()

	// 将带返回值的函数包装成 Pool 所需的 Task 形式
	if err := pool.Submit(func(ctx context.Context) error {
		var (
			res T
			err error
		)

		// 通过 defer 捕获 panic，保证无论成功、失败还是 panic，
		// Future 都能被正确标记为"已完成"并唤醒等待方。
		defer func() {
			if r := recover(); r != nil {
				var zero T
				future.complete(zero, panicError(r))
				return
			}
			future.complete(res, err)
		}()

		res, err = fn(ctx)
		return err
	}); err != nil {
		// 如果提交失败（如队列满且策略为返回错误），立即完成 Future 并返回错误
		var zero T
		future.complete(zero, err)
	}

	return future
}
