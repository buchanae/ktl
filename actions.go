package ktl

import (
	"context"
	"fmt"
)

type RestartStepOptions struct {
	BatchID        string
	StepID         string
  // KeepDownstream directs the restart action to avoid restarting downstream steps.
	KeepDownstream bool
}

func RestartStep(ctx context.Context, db Database, opts RestartStepOptions) error {
	if opts.BatchID == "" {
		return fmt.Errorf("empty batch ID")
	}
	if opts.StepID == "" {
		return fmt.Errorf("empty step ID")
	}

	batch, err := db.GetBatch(ctx, opts.BatchID)
	if err != nil {
		return fmt.Errorf("getting batch: %s", err)
	}

  target := FindStepByID(batch.Steps, opts.StepID)
	if target == nil {
		return ErrNotFound
	}

	toRestart := []*Step{target}

	if !opts.KeepDownstream {
		d := NewDAG(batch.Steps)
		toRestart = append(toRestart, AllDownstream(d, target)...)
	}

	for _, step := range toRestart {
		cpy := *step
		cpy.History = nil
		step.History = append(step.History, &cpy)

		step.State = Waiting
		step.Logs = nil
		step.Version++
	}

	err = db.UpdateBatch(ctx, batch)
	if err != nil {
		return fmt.Errorf("updating batch: %s", err)
	}
	return nil
}

type PutStepOptions struct {
  BatchID string
  Step *Step
  // TODO KeepDownstream?
}

func PutStep(ctx context.Context, db Database, opts PutStepOptions) error {
	if opts.BatchID == "" {
		return fmt.Errorf("empty batch ID")
	}

	batch, err := db.GetBatch(ctx, opts.BatchID)
  if err == ErrNotFound {
    batch = &Batch{ID: opts.BatchID}
	} else if err != nil {
		return fmt.Errorf("getting batch: %s", err)
	}

  target := FindStepByID(batch.Steps, opts.Step.ID)
  if target == nil {
    batch.Steps = append(batch.Steps, opts.Step)
  } else {
    // TODO check step hash before replacing to make PUT idempotent
		cpy := *target
		cpy.History = nil
		target.History = append(target.History, &cpy)

		target.State = Waiting
		target.Logs = nil
		target.Version++
  }

	err = db.UpdateBatch(ctx, batch)
	if err != nil {
		return fmt.Errorf("updating batch: %s", err)
	}
	return nil
}

func FindStepByID(steps []*Step, id string) *Step {
	for _, step := range steps {
		if step.ID == id {
      return step
		}
	}
  return nil
}
