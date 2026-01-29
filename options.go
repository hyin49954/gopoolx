package gopoolx

import "time"

// Options 封装了 Pool 的可配置项。
type Options struct {
	// retry 表示在任务执行失败时，最多额外重试的次数。
	// 实际执行总次数 = retry + 1
	retry int
	// retryDelay 是每次重试之间的等待时间。
	retryDelay time.Duration

	// queueSize 指任务队列（chan Task）的缓冲大小。
	//   - 0 表示无缓冲通道（提交和消费完全同步）
	//   - >0 表示有缓冲队列
	queueSize int
	// nonBlocking 预留字段：用于支持非阻塞提交任务。
	// 当前版本尚未在 Pool.Submit 中生效，未来可扩展为：
	//   - 提交失败时返回错误，或
	//   - 将任务丢弃并记录统计信息
	nonBlocking bool
}

// Option 是修改 Options 的函数式配置。
type Option func(*Options)

// WithRetry 设置任务执行失败时的重试次数。
func WithRetry(n int) Option {
	return func(o *Options) {
		o.retry = n
	}
}

// WithRetryDelay 设置重试之间的间隔时间。
func WithRetryDelay(d time.Duration) Option {
	return func(o *Options) {
		o.retryDelay = d
	}
}

// defaultOptions 返回 Pool 的默认配置。
func defaultOptions() *Options {
	return &Options{
		retry:      0,
		retryDelay: 0,
		queueSize:  0, // 0 = 无缓冲（最安全）
	}
}

// WithQueueSize 设置任务队列通道的缓冲大小。
func WithQueueSize(n int) Option {
	return func(o *Options) {
		o.queueSize = n
	}
}

// WithNonBlocking 启用非阻塞提交模式（预留）。
// 启用后，Submit 在任务队列已满时不会阻塞调用方，而是直接丢弃任务：
//   - 被丢弃的任务不会被执行
//   - Wait 不会因此卡死（WaitGroup 计数会被回滚）
//   - 错误收集器中也不会包含这些被丢弃的任务
func WithNonBlocking() Option {
	return func(o *Options) {
		o.nonBlocking = true
	}
}
