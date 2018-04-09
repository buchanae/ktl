package task

import (
  "fmt"
  "context"
  "time"
	"github.com/ohsu-comp-bio/tes"
	"github.com/ohsu-comp-bio/ktl"
	"github.com/mitchellh/mapstructure"
)

type Driver struct {
  cli *tes.Client
}

func NewDriver() (*Driver, error) {
	cli, err := tes.NewClient("http://localhost:8000")
  if err != nil {
    return nil, err
  }
  return &Driver{cli: cli}, nil
}

func (d *Driver) Check(ctx context.Context, spec *ktl.DriverSpec) error {
  taskdat := taskData{}
  err := mapstructure.Decode(spec.Logs, &taskdat)
  if err != nil {
    return fmt.Errorf("decoding task data: %s", err)
  }

  if taskdat.ID == "" {
    return nil
  }

  task, err := d.cli.GetTask(ctx, &tes.GetTaskRequest{
    Id:   taskdat.ID,
    View: tes.Minimal,
  })
  if err != nil {
    return fmt.Errorf("getting task: %s", err)
  }

  switch task.State {
  case tes.Complete:
    spec.State = ktl.Success
  case tes.SystemError:
    spec.State = ktl.Failed
    spec.Reason = "task system error"
  case tes.ExecutorError:
    spec.State = ktl.Failed
    spec.Reason = "task executor error"
  case tes.Canceled:
    spec.State = ktl.Failed
    spec.Reason = "task canceled"
  }
  return nil
}

func (d *Driver) Start(ctx context.Context, spec *ktl.DriverSpec) error {

  task := &tes.Task{}
  err := mapstructure.Decode(spec.Config, task)
  if err != nil {
    return fmt.Errorf("decoding task config: %s", err)
  }

  resp, err := d.cli.CreateTask(ctx, task)
  if err != nil {
    return fmt.Errorf("creating task: %s", err)
  }

  //startTime := time.Now()
  //spec.State = ktl.Active
  //step.StartedAt = &startTime
  spec.Logs = taskData{ID: resp.Id}
  return nil
}

func (d *Driver) Stop(ctx context.Context, step ktl.Step) error {
  taskdat := taskData{}
  err := mapstructure.Decode(step.Logs, &taskdat)
  if err != nil {
    return fmt.Errorf("decoding task data: %s", err)
  }

  if taskdat.ID == "" {
    return fmt.Errorf("step doesn't have task ID data")
  }

  _, err = d.cli.CancelTask(ctx, &tes.CancelTaskRequest{Id: taskdat.ID})
  if err != nil {
    return fmt.Errorf("canceling task: %s", err)
  }
  return nil
}

type taskData struct {
	ID string
}
