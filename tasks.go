package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/satori/go.uuid"
)

const (
	StatusPending = iota // pending task status.
	StatusQueued         // queued task status.
)

// TaskStat stores a runtime for a task.
type TaskStat struct {
	// Created is the task stat creation timestamp.
	// Key is the task key.
	// RunTime is the task run time in seconds.
	Created time.Time `json:"created"`
	Key     string    `json:"key"`
	RunTime float64   `json:"runtime"`
}

// NewTaskStat returns an initialized task instance.
func NewTaskStat(key string, runtime float64) *TaskStat {
	return &TaskStat{time.Now(), key, runtime}
}

// Save creates a new document for the task stat in the database.
func (taskStat *TaskStat) Save(taskModel Model) (DocumentMeta, error) {
	return taskModel.Save(taskStat)
}

// Task is a unit of work that is queued in the priority queue.
type Task struct {
	// Created is the task creation timestamp.
	// Id is the unique version 1 uuid assigned for task identification.
	// Key is the resource key for the task.
	// Meta is user defined data that can be added to the task.
	// Priority is the queue priority order.
	// RunAt is a static point in time execution time.
	// Status is the execution status of the task.
	Created  time.Time       `json:"created"`
	Id       string          `json:"_key"`
	Key      string          `json:"key"`
	Meta     json.RawMessage `json:"meta;omitempty"`
	Priority float64         `json:"priority"`
	RunAt    time.Time       `json:"runAt;omitempty"`
	Status   int             `json:"status"`
}

// NewTask returns an initialized task instance.
func NewTask(data []byte) *Task {
	task := &Task{Created: time.Now(), Status: StatusPending, Id: uuid.NewV1().String()}
	json.Unmarshal(data, task)
	return task
}

// ChangeStatus changes the status of the task and saves the task
func (task *Task) ChangeStatus(taskModel Model, status int) error {
	task.Status = status
	_, err := task.Save(taskModel)
	return err
}

// GetAverageRunTime returns the average of, up to, the 10 most recent task execution times
func (task *Task) GetAverageRunTime(taskStatModel Model) (float64, error) {
	var avg float64 = 0.0
	q := fmt.Sprintf(
		"FOR t IN %s FILTER t.key == @key LIMIT 10 SORT t.created DESC RETURN t",
		CollectionTaskStats,
	)
	taskStats, err := taskStatModel.Query(q, map[string]interface{}{"key": task.Key})
	if err != nil {
		return float64(-1), err
	}
	for _, taskStat := range taskStats {
		v, _ := taskStat.(*TaskStat)
		avg += v.RunTime
	}
	return avg / float64(len(taskStats)), nil
}

// Save writes the task to the database
func (task *Task) Save(taskModel Model) (DocumentMeta, error) {
	return taskModel.Save(task)
}
