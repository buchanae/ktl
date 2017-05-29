package whisk

import (
  "fmt"
  "time"
  "sync"
)

// Mock TES service
// Most of this is just placeholder code.

type TaskState int
const (
  Init TaskState = iota
  Running
  Done
)

type Task struct {
  ID string
  Cmd string
  Stdout string
  State TaskState
}

type TES interface {
  GetTask(id string) Task
  CreateTask(task Task) (string, error)
}

func NewTES() TES {
  return &tes{tasks: map[string]*Task{}}
}

type tes struct {
  id int
  mtx sync.Mutex
  tasks map[string]*Task
}
func (t *tes) GetTask(id string) Task {
  t.mtx.Lock()
  defer t.mtx.Unlock()
  return *t.tasks[id]
}

func (t *tes) CreateTask(task Task) (string, error) {
  t.mtx.Lock()
  defer t.mtx.Unlock()

  t.id++
  i := fmt.Sprintf("%d", t.id)

  t.tasks[i] = &task
  fmt.Println("RUN", task.Cmd)

  // Simulate task completing after 5 seconds
  go func() {
    time.Sleep(time.Second * 5)
    t.mtx.Lock()
    t.tasks[i].State = Done
    t.mtx.Unlock()
  }()

  return i, nil
}
