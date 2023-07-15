package worker

import (
	"fmt"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/heavybr/chikako/internal/task"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	DB        map[uuid.UUID]task.Task
	TaskCount int
}

func (w *Worker) CollectStats() {
	fmt.Println("Collecting stats...")
}

func (w *Worker) RunTask() {
	fmt.Println("Start/Stop a task...")
}

func (w *Worker) StopTask() {
	fmt.Println("Stopping task...")
}

func (w *Worker) StartTask() {
	fmt.Println("Starting task...")
}
