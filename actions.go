package ktl

import (
	"context"
	"fmt"
)

type RestartStepOptions struct {
	BatchID        string
	StepID         string
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

	var target *Step
	for _, step := range batch.Steps {
		if step.ID == opts.StepID {
			target = step
		}
	}
	if target == nil {
		return ErrNotFound
	}

	toRestart := []*Step{target}

	if !opts.KeepDownstream {
		d := NewDAG(batch.Steps)
		toRestart = append(toRestart, AllDownstream(d, target)...)
	}

	for _, step := range toRestart {
		copy := *step
		copy.History = nil
		step.History = append(step.History, &copy)

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
