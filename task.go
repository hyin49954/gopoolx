package gopoolx

import (
	"context"
	"crypto/rand"
	"fmt"
)

// TaskID 是每个提交任务的唯一标识。
type TaskID string

// Task 是提交到 Pool 中执行的基本任务类型。
// 参数为上层传入的上下文，允许任务根据 ctx 进行超时或取消控制。
type Task func(ctx context.Context) error

type poolTask struct {
	id   TaskID
	task Task
}

func newTaskID() (TaskID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate task id: %w", err)
	}

	// RFC 4122 version 4 UUID.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return TaskID(fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])), nil
}
