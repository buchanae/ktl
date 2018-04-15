package ktl

import (
  "fmt"
	"context"
	"github.com/ohsu-comp-bio/ktl/dag"
	"log"
	"time"
)

// DriverSpec is used to pass limited step information to drivers,
// so it's more clear what information a driver is expected to access/modify.
type DriverSpec struct {
	BatchID string
	StepID  string
	Config  interface{}
	Logs    interface{}
}

type CheckResult struct {
	State   State
	Reason  string
}

// Driver is the interface fulfilled by a step driver.
// Drivers are responsible for managing the state of a step.
// There are many types of drivers: start a task, wait for an event, etc.
type Driver interface {
	Check(context.Context, *DriverSpec) (*CheckResult, error)
	Start(context.Context, *DriverSpec) error
	Stop(context.Context, *DriverSpec) error
}

// Process is the main control loop, responsible for managing the state
// of batches and their steps. Process periodically checks for active batches
// and calls the step drivers to manage step state.
func Process(ctx context.Context, db Database, drivers map[string]Driver) {
  // TODO configurable
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

  for {
    select {
    case <-ctx.Done():
      return
    case <-ticker.C:

      // Get all active batches
      batches, err := db.ListBatches(ctx, &BatchListOptions{
        State: []State{Waiting, Active},
      })
      if err != nil {
        log.Println("error listing batches", err)
        continue
      }

      // For each batch, reconcile any state changes.
      for _, batch := range batches {

        // Check step state via driver.
        for _, step := range batch.Steps {
          // TODO only check steps that aren't finished.
          err := checkStep(ctx, batch.ID, step, drivers)
          if err != nil {
            log.Println("error checking step:", err)
          }
        }

        processBatch(ctx, batch)

        // TODO concurrency checks
        err = db.UpdateBatch(ctx, batch)
        if err != nil {
          log.Println("error updating batch state:", err)
          continue
        }

        for _, step := range batch.Steps {
          err := processEvents(ctx, batch.ID, step, drivers)
          if err != nil {
            log.Println(err)
          }
        }

        // TODO concurrency checks
        err = db.UpdateBatch(ctx, batch)
        if err != nil {
          log.Println("error updating batch state after driving:", err)
          continue
        }
      }
    }
  }
}

func processEvents(ctx context.Context, batchID string, step *Step, drivers map[string]Driver) error {
  driver, ok := drivers[step.Type]
  if !ok {
    return fmt.Errorf("unknown step driver type: %s", step.Type)
  }

  for _, event := range step.Events {
    if event.Processed {
      continue
    }

    spec := &DriverSpec{
      BatchID: batchID,
      StepID: step.ID,
      Config: step.Config,
      Logs: step.Logs,
    }

    switch event.Type {
    case Start:
      err := driver.Start(ctx, spec)
      if err != nil {
        return fmt.Errorf("driver.Start failed", err)
      }

    case Stop:
      err := driver.Stop(ctx, spec)
      if err != nil {
        return fmt.Errorf("driver.Stop failed", err)
      }
    }

    event.Processed = true
    step.Config = spec.Config
    step.Logs = spec.Logs
  }
  return nil
}

func checkStep(ctx context.Context, batchID string, step *Step, drivers map[string]Driver) error {

  driver, ok := drivers[step.Type]
  if !ok {
    return fmt.Errorf("unknown step driver type: %s", step.Type)
  }

  spec := &DriverSpec{
    BatchID: batchID,
    StepID: step.ID,
    Config: step.Config,
    Logs: step.Logs,
  }
  res, err := driver.Check(ctx, spec)
  if err != nil {
    return fmt.Errorf("checking step %s: %s", step.ID, err)
  }
  if res != nil {
    step.State = res.State
    step.Reason = res.Reason
  }
  return nil
}

// processBatch processes a single batch. This is where most of the work happens.
func processBatch(ctx context.Context, batch *Batch) {
	defer UpdateBatchCounts(batch)

	for _, step := range batch.Steps {
		if step.State.Done() || step.State == Paused {
			continue
		}

		// TODO while the process loop might happen often, it might not, and if it's slow
		//      deadlines and timeouts might be imprecise. would be nice to have a system
		//      with high precision.

		// Check the step deadline.
		if step.Deadline != nil && step.Deadline.Sub(time.Now()) < 0 {
			step.State = Failed
			step.Reason = "deadline exceeded"
			step.Events = append(step.Events, NewEvent(Stop))
			continue
		}

		// Check the step timeout.
		if step.StartedAt != nil && step.Timeout > 0 &&
			time.Now().Sub(*step.StartedAt) > time.Duration(step.Timeout) {

			step.State = Failed
			step.Reason = "timeout exceeded"
			step.Events = append(step.Events, NewEvent(Stop))
			continue
		}
	}

	var failed []*Step
	for _, step := range batch.Steps {
		if step.State == Failed {
			failed = append(failed, step)
		}
	}

	d := BatchDAG(batch)
  all := d.AllNodes()

	if dag.AllDone(all) {
		if failed != nil {
			batch.State = Failed
		} else {
			batch.State = Success
		}
		return
	}

	// In fail fast mode, the batch stops one the first error encountered,
	// stopping any steps which are running.
	if batch.Mode == FailFast && failed != nil {
		batch.State = Failed

		// Stop any active steps.
		for _, step := range batch.Steps {
			if step.State == Active {
				step.State = Failed
				step.Reason = "batch failed fast"
				step.Events = append(step.Events, NewEvent(Stop))
			}
		}
		return
	}

	ready := dag.Ready(d, all)
  active := dag.FilterByState(all, dag.Active)
  terminals := dag.Terminals(d, all)
  blockers := dag.Blockers(d, terminals)

  // If all terminal steps are blocked, and nothing is running, and nothing is ready,
  // there's nothing to do. If all the remaining steps are blocked, we consider
  // the batch as failed.
  if active == nil && ready == nil && dag.AllBlocked(d, terminals) &&
     dag.AllState(blockers, dag.Failed) {

    // All the remaining steps are blocked by failed nodes.
    // Consider the batch failed.
    batch.State = Failed
    return
  }

	// Execute next steps.
	for _, node := range ready {
		step := node.(*Step)
		step.State = Active
		step.Events = append(step.Events, NewEvent(Start))
	}
}

func RestartStep(ctx context.Context, db Database, batchID, stepID string) error {
  batch, err := db.GetBatch(ctx, batchID)
  if err != nil {
    return fmt.Errorf("getting batch: %s", err)
  }

  for _, step := range batch.Steps {
    if step.ID == stepID {
      step.State = Waiting
      step.Events = append(step.Events, NewEvent(Stop))
      step.Events = append(step.Events, NewEvent(Start))

      if batch.State.Done() {
        batch.State = Waiting
      }

      err := db.UpdateBatch(ctx, batch)
      if err != nil {
        return fmt.Errorf("updating batch: %s", err)
      }
      return nil
    }
  }
  return ErrNotFound
}
