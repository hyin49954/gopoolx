package gopoolx

import "context"

// Task 是提交到 Pool 中执行的基本任务类型。
// 参数为上层传入的上下文，允许任务根据 ctx 进行超时或取消控制。
type Task func(ctx context.Context) error
