package gopoolx

import (
	"context"
	"sync"
	"time"
)

// Pool 表示一个固定 worker 数量的 goroutine 池，用于并发执行 Task。
// 使用方式一般为：
//  1. 通过 New 创建池实例
//  2. 调用 Run(ctx) 启动 worker
//  3. 使用 Submit 提交任务
//  4. 调用 Wait 等待所有任务完成并关闭池
type Pool struct {
	// workerNum 是并发执行任务的 worker 数量
	workerNum int
	// tasks 是任务队列，worker 会从该通道中取出任务执行
	tasks chan Task
	// wg 用于等待所有提交的任务执行完成
	wg sync.WaitGroup
	// once 用于确保任务通道只会被关闭一次，避免多次 Wait 调用导致 panic
	once sync.Once

	// opts 存放池的配置项（重试次数、队列大小等）
	opts *Options
	// errs 收集所有执行失败的任务错误
	errs *ErrorCollector
}

// New 创建一个新的 Pool。
//   - workerNum: worker 的数量（应为正数）
//   - opts: 可选配置，例如重试次数、队列大小等
func New(workerNum int, opts ...Option) *Pool {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	var ch chan Task
	if o.queueSize > 0 {
		ch = make(chan Task, o.queueSize)
	} else {
		ch = make(chan Task)
	}

	return &Pool{
		workerNum: workerNum,
		tasks:     ch,
		opts:      o,
		errs:      &ErrorCollector{},
	}
}

// Submit 提交一个任务到池中，内部会递增 WaitGroup 计数。
// 根据配置的队列满策略，行为如下：
//   - QueueFullWait: 队列满时阻塞等待，直到有空位再插入（默认）
//   - QueueFullDiscard: 队列满时直接丢弃任务，不返回错误
//   - QueueFullReturnError: 队列满时返回 ErrQueueFull 错误，任务计入失败
func (p *Pool) Submit(task Task) error {
	p.wg.Add(1)

	switch p.opts.queueFullPolicy {
	case QueueFullDiscard:
		// 队列满时直接丢弃任务
		select {
		case p.tasks <- task:
			// 正常入队，由 worker 负责执行并在结束时调用 wg.Done
		default:
			// 队列已满：撤销之前的 Add，保持 WaitGroup 计数正确
			p.wg.Done()
		}
		return nil

	case QueueFullReturnError:
		// 队列满时返回错误，任务计入失败
		select {
		case p.tasks <- task:
			// 正常入队，由 worker 负责执行并在结束时调用 wg.Done
		default:
			// 队列已满：撤销之前的 Add，将错误加入错误收集器，并返回错误
			p.wg.Done()
			p.errs.Add(ErrQueueFull)
			return ErrQueueFull
		}
		return nil

	case QueueFullWait:
		fallthrough
	default:
		// 默认等待模式：在任务队列满时阻塞，直到有空间写入
		p.tasks <- task
		return nil
	}
}

// Run 启动指定数量的 worker。
// ctx 结束时（超时、取消等），worker 会自动退出。
func (p *Pool) Run(ctx context.Context) {
	for i := 0; i < p.workerNum; i++ {
		go p.worker(ctx)
	}
}

// worker 是实际执行 Task 的 worker 循环。
// 它会根据 ctx 或任务通道关闭而退出。
func (p *Pool) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			p.executeWithRetry(ctx, task)
			p.wg.Done()
		}
	}
}

// executeWithRetry 根据配置执行任务，并在失败时进行重试。
// 当超过最大重试次数后，会将最终错误加入错误收集器。
func (p *Pool) executeWithRetry(ctx context.Context, task Task) {
	var err error
	// 统一 panic 恢复：无论是否开启重试，任务中的 panic
	// 都会被转换为 error 并加入错误收集器，避免 worker 整体崩溃。
	defer func() {
		if r := recover(); r != nil {
			p.errs.Add(panicError(r))
			return
		}
		// 非 panic 场景下，如果最终仍有错误，则收集错误
		if err != nil {
			p.errs.Add(err)
		}
	}()

	for i := 0; i <= p.opts.retry; i++ {
		err = task(ctx)
		if err == nil {
			return
		}
		if p.opts.retryDelay > 0 {
			time.Sleep(p.opts.retryDelay)
		}
	}
}

// Wait 阻塞等待所有已提交任务执行完成，并在首次调用时关闭任务通道。
// 多次调用是安全的（多次调用只会在第一次时真正关闭通道）。
func (p *Pool) Wait() {
	p.wg.Wait()
	// 通过 once 保证 tasks 只会被关闭一次，避免调用方误多次调用 Wait 时 panic。
	p.once.Do(func() {
		close(p.tasks)
	})
}

// Errors 返回一个包含所有任务执行错误的切片副本。
// 返回的是拷贝，调用方可以安全地在外部修改。
func (p *Pool) Errors() []error {
	return p.errs.Errors()
}
