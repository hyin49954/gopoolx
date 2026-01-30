## gopoolx

🚀 **gopoolx** 是一个工程级的 Goroutine 并发任务池，用于在 Go 项目中**安全、可控、高效**地执行大量并发任务。

它不是 Demo，也不是单纯的协程池，而是为真实生产场景设计的任务执行框架。

---

## ✨ 特性

- **固定 Worker 数量**：限制并发度，防止 goroutine 爆炸
- **统一上下文控制**：基于 `context.Context` 的取消 / 超时控制
- **失败自动重试**：支持设置重试次数与重试间隔（`WithRetry` / `WithRetryDelay`）
- **统一错误收集**：所有任务执行错误集中到 `pool.Errors()` 中
- **panic 自动恢复**：
  - 普通任务与带返回值任务中的 panic 都会被安全捕获并转换为 `error`
  - 不会打爆整个 worker 协程
- **Future 泛型结果**：通过 `SubmitWithResult` + `Future[T]` 支持有返回值任务的异步等待
- **队列满策略**：
  提供三种队列满时的处理策略：
  - `QueueFullWait`（默认）：阻塞等待直到有空位
  - `QueueFullDiscard`：直接丢弃任务
  - `QueueFullReturnError`：返回错误并记录失败
- **简单、清晰、工程化 API**：贴近真实业务代码的使用方式

---

## 📦 安装

```bash
go get github.com/yourname/gopoolx
```

---

## 🚀 快速上手

### 固定大小协程池 + 失败重试

```go
pool := gopoolx.New(
    5,
    gopoolx.WithRetry(2),
    gopoolx.WithRetryDelay(200*time.Millisecond),
    gopoolx.WithQueueSize(100),
)

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

pool.Run(ctx)

for i := 0; i < 1000; i++ {
    pool.Submit(func(ctx context.Context) error {
        // 执行你的任务
        return nil
    })
}

pool.Wait()

for _, err := range pool.Errors() {
    log.Println("task error:", err)
}
```

### 使用 Future 提交有返回值任务

```go
pool := gopoolx.New(3)
ctx := context.Background()

pool.Run(ctx)

f1 := gopoolx.SubmitWithResult(pool, func(ctx context.Context) (int, error) {
    time.Sleep(time.Second)
    return 100, nil
})

f2 := gopoolx.SubmitWithResult(pool, func(ctx context.Context) (string, error) {
    return "hello gopoolx", nil
})

v1, _ := f1.Get(ctx) // 支持通过 ctx 控制等待超时/取消
v2, _ := f2.Get(ctx)

fmt.Println(v1, v2)

pool.Wait()
```

### 队列满策略

gopoolx 提供了三种队列满时的处理策略：

#### 1. 等待模式（默认）

```go
pool := gopoolx.New(
    10,
    gopoolx.WithQueueSize(1000),
    // QueueFullWait 是默认策略，此配置可选
    gopoolx.WithQueueFullPolicy(gopoolx.QueueFullWait),
)

pool.Run(context.Background())

// 队列满时 Submit 会阻塞，直到有空位再插入
pool.Submit(func(ctx context.Context) error {
    return nil
})
```

#### 2. 丢弃模式

```go
pool := gopoolx.New(
    10,
    gopoolx.WithQueueSize(1000),
    gopoolx.WithQueueFullPolicy(gopoolx.QueueFullDiscard), // 队列满时直接丢弃任务
)

pool.Run(context.Background())

// Submit 会立即返回 nil，被丢弃的任务不会执行
err := pool.Submit(func(ctx context.Context) error {
    return nil
})
// 丢弃模式下 err 始终为 nil
```

> 注意：丢弃模式下，被丢弃的任务不会执行，也不会出现在 `pool.Errors()` 中。

#### 3. 返回错误模式

```go
pool := gopoolx.New(
    10,
    gopoolx.WithQueueSize(1000),
    gopoolx.WithQueueFullPolicy(gopoolx.QueueFullReturnError), // 队列满时返回错误
)

pool.Run(context.Background())

// 队列满时 Submit 会返回 ErrQueueFull 错误
err := pool.Submit(func(ctx context.Context) error {
    return nil
})
if err != nil {
    // 处理队列满错误
    log.Println("提交任务失败:", err)
}
```

> 注意：返回错误模式下，提交失败的任务会被记录到 `pool.Errors()` 中。

---

## ⚙️ 设计要点

- `Pool` 使用固定数量的 worker 与 `chan Task` 作为任务队列
- `executeWithRetry` 统一处理：
  - 失败自动重试
  - panic 恢复并转为 `error`
- `ErrorCollector` 提供并发安全的错误收集能力
- `Future[T]` 提供类型安全的异步结果获取接口

---

