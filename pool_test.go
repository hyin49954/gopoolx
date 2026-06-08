package gopoolx

import (
	"context"
	"errors"
	"testing"
)

func TestSubmitUsesProvidedTaskID(t *testing.T) {
	pool := New(1)
	ctx := context.Background()
	wantTaskID := TaskID("custom-task-id")
	taskErr := errors.New("custom task failed")

	pool.Run(ctx)

	taskID, err := pool.Submit(func(ctx context.Context) (any, error) {
		return nil, taskErr
	}, wantTaskID)
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if taskID != wantTaskID {
		t.Fatalf("Submit() task ID = %s, want %s", taskID, wantTaskID)
	}

	pool.Wait()

	got, ok := pool.Error(wantTaskID)
	if !ok {
		t.Fatalf("Error(%s) not found", wantTaskID)
	}
	if !errors.Is(got.Err, taskErr) {
		t.Fatalf("Error(%s).Err = %v, want %v", wantTaskID, got.Err, taskErr)
	}
}

func TestPoolErrorByTaskID(t *testing.T) {
	pool := New(1)
	ctx := context.Background()
	taskErr := errors.New("task failed")

	pool.Run(ctx)

	taskID, err := pool.Submit(func(ctx context.Context) (any, error) {
		return nil, taskErr
	})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if taskID == "" {
		t.Fatal("Submit() task ID is empty")
	}

	pool.Wait()

	got, ok := pool.Error(taskID)
	if !ok {
		t.Fatalf("Error(%s) not found", taskID)
	}
	if !errors.Is(got.Err, taskErr) {
		t.Fatalf("Error(%s).Err = %v, want %v", taskID, got.Err, taskErr)
	}

	all := pool.Errors()
	if len(all) != 1 {
		t.Fatalf("len(Errors()) = %d, want 1", len(all))
	}
	if all[0].TaskID != taskID {
		t.Fatalf("Errors()[0].TaskID = %s, want %s", all[0].TaskID, taskID)
	}
	if !errors.Is(all[0].Err, taskErr) {
		t.Fatalf("Errors()[0].Err = %v, want %v", all[0].Err, taskErr)
	}
}

func TestTaskErrorWithData(t *testing.T) {
	pool := New(1)
	ctx := context.Background()
	taskErr := errors.New("task failed")
	wantData := map[string]int{"processed": 3}

	pool.Run(ctx)

	taskID, err := pool.Submit(func(ctx context.Context) (any, error) {
		return wantData, taskErr
	})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	pool.Wait()

	got, ok := pool.Error(taskID)
	if !ok {
		t.Fatalf("Error(%s) not found", taskID)
	}
	if !errors.Is(got.Err, taskErr) {
		t.Fatalf("Error(%s).Err = %v, want %v", taskID, got.Err, taskErr)
	}

	data, ok := got.Data.(map[string]int)
	if !ok {
		t.Fatalf("Error(%s).Data type = %T, want map[string]int", taskID, got.Data)
	}
	if data["processed"] != wantData["processed"] {
		t.Fatalf("Error(%s).Data = %v, want %v", taskID, data, wantData)
	}
}

func TestSubmitQueueFullErrorHasTaskID(t *testing.T) {
	pool := New(
		1,
		WithQueueSize(1),
		WithQueueFullPolicy(QueueFullReturnError),
	)

	firstTaskID, err := pool.Submit(func(ctx context.Context) (any, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatalf("first Submit() error = %v", err)
	}
	if firstTaskID == "" {
		t.Fatal("first Submit() task ID is empty")
	}

	secondTaskID, err := pool.Submit(func(ctx context.Context) (any, error) {
		return nil, nil
	})
	if !errors.Is(err, ErrQueueFull) {
		t.Fatalf("second Submit() error = %v, want %v", err, ErrQueueFull)
	}
	if secondTaskID == "" {
		t.Fatal("second Submit() task ID is empty")
	}
	if secondTaskID == firstTaskID {
		t.Fatalf("second task ID = %s, want different from %s", secondTaskID, firstTaskID)
	}

	got, ok := pool.Error(secondTaskID)
	if !ok {
		t.Fatalf("Error(%s) not found", secondTaskID)
	}
	if !errors.Is(got.Err, ErrQueueFull) {
		t.Fatalf("Error(%s).Err = %v, want %v", secondTaskID, got.Err, ErrQueueFull)
	}
}

func TestDrainErrorsClearsOnlyCurrentErrors(t *testing.T) {
	pool := New(1)
	oldTaskID := TaskID("old-task")
	newTaskID := TaskID("new-task")
	oldErr := errors.New("old error")
	newErr := errors.New("new error")

	pool.errs.Add(oldTaskID, oldErr, "old-data")

	drained := pool.DrainErrors()
	if len(drained) != 1 {
		t.Fatalf("len(DrainErrors()) = %d, want 1", len(drained))
	}
	if drained[0].TaskID != oldTaskID {
		t.Fatalf("DrainErrors()[0].TaskID = %s, want %s", drained[0].TaskID, oldTaskID)
	}
	if drained[0].Data != "old-data" {
		t.Fatalf("DrainErrors()[0].Data = %v, want old-data", drained[0].Data)
	}
	if _, ok := pool.Error(oldTaskID); ok {
		t.Fatalf("Error(%s) found after DrainErrors()", oldTaskID)
	}
	if got := pool.Errors(); len(got) != 0 {
		t.Fatalf("len(Errors()) after DrainErrors() = %d, want 0", len(got))
	}

	pool.errs.Add(newTaskID, newErr, nil)

	got, ok := pool.Error(newTaskID)
	if !ok {
		t.Fatalf("Error(%s) not found after DrainErrors()", newTaskID)
	}
	if !errors.Is(got.Err, newErr) {
		t.Fatalf("Error(%s).Err = %v, want %v", newTaskID, got.Err, newErr)
	}
}

func TestSubmitWithResultUsesProvidedTaskID(t *testing.T) {
	pool := New(1)
	ctx := context.Background()
	wantTaskID := TaskID("custom-result-task-id")

	pool.Run(ctx)

	taskID, future, err := SubmitWithResult(pool, func(ctx context.Context) (int, error) {
		return 42, nil
	}, wantTaskID)
	if err != nil {
		t.Fatalf("SubmitWithResult() error = %v", err)
	}
	if taskID != wantTaskID {
		t.Fatalf("SubmitWithResult() task ID = %s, want %s", taskID, wantTaskID)
	}

	got, err := future.Get(ctx)
	if err != nil {
		t.Fatalf("Future.Get() error = %v", err)
	}
	if got != 42 {
		t.Fatalf("Future.Get() result = %d, want 42", got)
	}

	pool.Wait()
}

func TestSubmitWithResultErrorIncludesData(t *testing.T) {
	pool := New(1)
	ctx := context.Background()
	taskErr := errors.New("result task failed")

	pool.Run(ctx)

	taskID, future, err := SubmitWithResult(pool, func(ctx context.Context) (int, error) {
		return 7, taskErr
	})
	if err != nil {
		t.Fatalf("SubmitWithResult() error = %v", err)
	}

	res, err := future.Get(ctx)
	if !errors.Is(err, taskErr) {
		t.Fatalf("Future.Get() error = %v, want %v", err, taskErr)
	}
	if res != 7 {
		t.Fatalf("Future.Get() result = %d, want 7", res)
	}

	pool.Wait()

	got, ok := pool.Error(taskID)
	if !ok {
		t.Fatalf("Error(%s) not found", taskID)
	}
	if !errors.Is(got.Err, taskErr) {
		t.Fatalf("Error(%s).Err = %v, want %v", taskID, got.Err, taskErr)
	}
	if got.Data != 7 {
		t.Fatalf("Error(%s).Data = %v, want 7", taskID, got.Data)
	}
}
