package task

import (
	"context"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/ohsu-comp-bio/ktl"
	"github.com/ohsu-comp-bio/tes"
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

func (d *Driver) Check(ctx context.Context, spec *ktl.DriverSpec) (*ktl.CheckResult, error) {
	taskdat := taskData{}
	err := mapstructure.Decode(spec.Logs, &taskdat)
	if err != nil {
		return nil, fmt.Errorf("decoding task data: %s", err)
	}

	if taskdat.ID == "" {
		return &ktl.CheckResult{}, nil
	}

	task, err := d.cli.GetTask(ctx, &tes.GetTaskRequest{
		Id:   taskdat.ID,
		View: tes.Minimal,
	})
	if err == tes.ErrNotFound {
		// TODO "unknown" would be more descriptive
		// TODO would be good to clear the task ID from the step logs?
		//      or keep a mapping of version to task ID?
		return &ktl.CheckResult{
			State:   ktl.Waiting,
			Reason:  "task not found",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting task: %s", err)
	}

	switch task.State {
	case tes.Complete:
		return &ktl.CheckResult{
			State:   ktl.Success,
		}, nil

	case tes.SystemError:
		return &ktl.CheckResult{
			State:   ktl.Failed,
			Reason:  "task system error",
		}, nil

	case tes.ExecutorError:
		return &ktl.CheckResult{
			State:   ktl.Failed,
			Reason:  "task executor error",
		}, nil

	case tes.Canceled:
		return &ktl.CheckResult{
			State:   ktl.Failed,
			Reason:  "task canceled",
		}, nil

	case tes.Queued, tes.Paused, tes.Initializing, tes.Running:
		return &ktl.CheckResult{
			State:   ktl.Active,
		}, nil
	}
	return nil, nil
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

	spec.Logs = taskData{ID: resp.Id}
	return nil
}

func (d *Driver) Stop(ctx context.Context, spec *ktl.DriverSpec) error {
	taskdat := taskData{}
	err := mapstructure.Decode(spec.Logs, &taskdat)
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
	ID      string
}
