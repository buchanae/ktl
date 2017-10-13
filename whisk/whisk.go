package whisk

import (
  "fmt"
  "errors"
  "time"
  "sync"
)

type Workflow interface {
  GetTask(id string) (Task, error)
  GetTasks(ids []string) ([]Task, error)
  CreateTask(task Task) error

  MustGetTask(id string) Task
  MustGetTasks(ids []string) []Task

  NS(id string) Workflow
  MustGo(f func())
  Wait() error
  Recover()
}


func NewWorkflow(id string) Workflow {
  return &workflow{
    id: id,
    db: memdb{},
    tes: NewTES(),
    wg: new(sync.WaitGroup),
  }
}


type workflow struct {
  id string
  db database
  tes TES
  wg *sync.WaitGroup
}

func (wf *workflow) NS(id string) Workflow {
  return &workflow{
    id: wf.id + "/" + id,
    db: wf.db,
    tes: wf.tes,
    wg: wf.wg,
  }
}

func (wf *workflow) Recover() {
  if r := recover(); r != nil {
    fmt.Println("Recover", r)
  }
}

func (wf *workflow) Wait() error {
  wf.wg.Wait()
  return nil
}

func (wf *workflow) MustGo(f func()) {
  wf.wg.Add(1)
  go func() {
    defer wf.Recover()
    f()
    wf.wg.Done()
  }()
}


func (wf *workflow) GetTask(id string) (Task, error) {
  ticker := time.NewTicker(time.Second)

  // Check every tick until the task can be found.
  for range ticker.C {
    // Look up id -> task ID mapping
    if tid, ok := wf.db.get(id); ok {
      t := wf.tes.GetTask(tid)
      if t.State != Done {
        continue
      }
      return t, nil
    }
  }

  return Task{}, errors.New("TODO")
}


func (wf *workflow) GetTasks(ids []string) ([]Task, error) {
  var tasks []Task
  for _, id := range ids {
    t, err := wf.GetTask(id)
    if err != nil {
      return nil, err
    }
    tasks = append(tasks, t)
  }
  return tasks, nil
}


func (wf *workflow) CreateTask(task Task) error {
  fmt.Println("Create", task.ID)
  // If the task already exists, do nothing.
  if _, ok := wf.db.get(task.ID); ok {
    return nil
  }

  // Create the task. Save the mapping from whisk ID to TES ID.
  id, err := wf.tes.CreateTask(task)
  if err != nil {
    return err
  }

  wf.db.put(task.ID, id)
  wf.wg.Add(1)

  // Wait for the task to complete
  go func() {
    defer wf.wg.Done()
    ticker := time.NewTicker(time.Second)

    for range ticker.C {
      t := wf.tes.GetTask(id)
      if t.State == Done {
        return
      }
    }
  }()

  return nil
}


func (wf *workflow) MustGetTask(id string) Task {
  t, err := wf.GetTask(id)
  if err != nil {
    panic(err)
  }
  return t
}

func (wf *workflow) MustGetTasks(ids []string) []Task {
  t, err := wf.GetTasks(ids)
  if err != nil {
    panic(err)
  }
  return t
}

/*
func (wf *workflow) MustCreateTask(task Task) {
  err := wf.CreateTask(task)
  if err != nil {
    panic(err)
  }
}
*/
