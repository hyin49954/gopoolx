## gopoolx

An engineering-level goroutine pool for running massive concurrent tasks **safely and efficiently**.

Designed for real-world production scenarios, not just toy examples.

---

## Features

- **Fixed worker pool**  
  Control concurrency and prevent goroutine explosion.

- **Context-aware execution**  
  Built-in support for `context.Context` cancellation and timeouts.

- **Retry with configurable backoff**  
  - `WithRetry(n)` – number of retries on failure  
  - `WithRetryDelay(d)` – delay between retries

- **Centralized error collection**  
  All task errors are collected and can be obtained via `pool.Errors()`.

- **Panic recovery**  
  - Panics inside tasks (both normal and result-returning) are safely recovered
  - Converted into `error` so they do not crash workers

- **Generic Future results**  
  Use `SubmitWithResult` + `Future[T]` to run tasks that return values.

- **Queue full policy**  
  Three strategies when the queue is full:
  - `QueueFullWait` (default): Block until space is available
  - `QueueFullDiscard`: Drop the task silently
  - `QueueFullReturnError`: Return an error and record the failure

- **Simple, production-friendly API**

---

## Install

```bash
go get github.com/yourname/gopoolx
```

---

## Quick Start

### Fixed-size pool with retry

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
        // your task here
        return nil
    })
}

pool.Wait()

for _, err := range pool.Errors() {
    log.Println("task error:", err)
}
```

### Using Future for tasks with return values

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

v1, _ := f1.Get(ctx) // can be controlled via ctx for timeout/cancel
v2, _ := f2.Get(ctx)

fmt.Println(v1, v2)

pool.Wait()
```

### Queue full policy

gopoolx provides three strategies when the task queue is full:

#### 1. Wait (default)

```go
pool := gopoolx.New(
    10,
    gopoolx.WithQueueSize(1000),
    // QueueFullWait is the default, so this is optional
    gopoolx.WithQueueFullPolicy(gopoolx.QueueFullWait),
)

pool.Run(context.Background())

// Submit will block if the queue is full until space is available
pool.Submit(func(ctx context.Context) error {
    return nil
})
```

#### 2. Discard

```go
pool := gopoolx.New(
    10,
    gopoolx.WithQueueSize(1000),
    gopoolx.WithQueueFullPolicy(gopoolx.QueueFullDiscard), // drop tasks when queue is full
)

pool.Run(context.Background())

// Submit will return nil immediately, dropped tasks are not executed
err := pool.Submit(func(ctx context.Context) error {
    return nil
})
// err is always nil in discard mode
```

> Note: In discard mode, dropped tasks are **not executed** and will **not** appear in `pool.Errors()`.

#### 3. Return error

```go
pool := gopoolx.New(
    10,
    gopoolx.WithQueueSize(1000),
    gopoolx.WithQueueFullPolicy(gopoolx.QueueFullReturnError), // return error when queue is full
)

pool.Run(context.Background())

// Submit will return ErrQueueFull if the queue is full
err := pool.Submit(func(ctx context.Context) error {
    return nil
})
if err != nil {
    // Handle queue full error
    log.Println("Failed to submit task:", err)
}
```

> Note: In return error mode, failed submissions are recorded in `pool.Errors()`.

---

## Design Highlights

- `Pool` uses a fixed number of workers and a `chan Task` as the task queue.
- `executeWithRetry` is responsible for:
  - retry logic
  - panic recovery (converting to `error`)
- `ErrorCollector` provides concurrency-safe error aggregation.
- `Future[T]` exposes a type-safe async result API.

---

