package gopoolx

import (
	"errors"
	"time"
)

// QueueFullPolicy 定义队列满时的处理策略类型
type QueueFullPolicy int

const (
	// QueueFullWait 队列满时等待，直到有空位再插入（默认策略）
	QueueFullWait QueueFullPolicy = iota
	// QueueFullDiscard 队列满时直接丢弃任务
	QueueFullDiscard
	// QueueFullReturnError 队列满时返回错误，任务计入失败
	QueueFullReturnError
)

// ErrQueueFull 表示队列已满的错误
var ErrQueueFull = errors.New("task queue is full")

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
	// queueFullPolicy 定义队列满时的处理策略：
	//   - QueueFullWait: 等待，直到有空位再插入（默认）
	//   - QueueFullDiscard: 直接丢弃任务
	//   - QueueFullReturnError: 返回错误，任务计入失败
	queueFullPolicy QueueFullPolicy
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
		retry:           0,
		retryDelay:      0,
		queueSize:       0,             // 0 = 无缓冲（最安全）
		queueFullPolicy: QueueFullWait, // 默认等待策略
	}
}

// WithQueueSize 设置任务队列通道的缓冲大小。
func WithQueueSize(n int) Option {
	return func(o *Options) {
		o.queueSize = n
	}
}

// WithQueueFullPolicy 设置队列满时的处理策略。
// 可选策略：
//   - QueueFullWait: 等待，直到有空位再插入（默认）
//   - QueueFullDiscard: 直接丢弃任务，不返回错误
//   - QueueFullReturnError: 返回错误，任务计入失败
func WithQueueFullPolicy(policy QueueFullPolicy) Option {
	return func(o *Options) {
		o.queueFullPolicy = policy
	}
}
