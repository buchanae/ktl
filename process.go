package ktl

import (
	"context"
  "fmt"
	"github.com/ohsu-comp-bio/ktl/dag"
	"log"
	"time"
)

type Driver interface {
  Check(context.Context, *Step) error
  Start(context.Context, *Step) error
  Stop(context.Context, *Step) error
}

func Process(db Database, drivers map[string]Driver) {
	ctx := context.Background()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {

		// Get all active batches
		batches, err := db.ListBatches(ctx, &BatchListOptions{
			State: []State{Idle, Running},
		})
		if err != nil {
			log.Println("error listing batches", err)
			continue
		}

		// For each batch, reconcile any state changes.
		for _, batch := range batches {
			processBatch(ctx, batch, db, drivers)

			err = db.UpdateBatch(ctx, batch)
			if err != nil {
				log.Println("error updating batch state:", err)
			}
		}
	}
}

func processBatch(ctx context.Context, batch *Batch, db Database, drivers map[string]Driver) {
	defer UpdateBatchCounts(batch)

	for _, step := range batch.Steps {
		if step.Done() {
			continue
		}

    driver, ok := drivers[step.Type]
    if !ok {
      step.State = Failed
      step.Reason = fmt.Sprint(`unknown driver "%s"`, step.Type)
      continue
    }

    // TODO drivers should be canceling tasks in this situation
		if step.Deadline != nil && step.Deadline.Sub(time.Now()) < 0 {
			step.State = Failed
			step.Reason = "deadline exceeded"

      // TODO think about best error handling
      err := driver.Stop(ctx, step)
      if err != nil {
        log.Println("error stopping step %s: %s", step.ID, err)
      }
			continue
		}

		if step.StartedAt != nil && step.Timeout > 0 && time.Now().Sub(*step.StartedAt) > step.Timeout {
			step.State = Failed
			step.Reason = "timeout exceeded"

      // TODO think about best error handling
      err := driver.Stop(ctx, step)
      if err != nil {
        log.Println("error stopping step %s: %s", step.ID, err)
      }
			continue
		}

    // TODO better error handling?
    err := driver.Check(ctx, step)
    if err != nil {
      log.Println("error checking step %s: %s", step.ID, err)
    }
	}

	// Calculate the next available steps
	d := BatchDAG(batch)
  ready := dag.Ready(d, d.AllNodes())
  err := dag.Errors(d.AllNodes())

	if dag.AllDone(d.AllNodes()) {
    if err != nil {
      batch.State = Failed
      batch.Reason = err.Error()
    } else {
		  batch.State = Success
    }
		return
	}

  if batch.Mode == FailFast && err != nil {
    batch.State = Failed
    batch.Reason = err.Error()
    return
  }

	// Execute next steps, using available step drivers.
	for _, node := range ready {
		step := node.(*Step)
    driver, ok := drivers[step.Type]
    if !ok {
      step.State = Failed
      step.Reason = fmt.Sprint(`unknown driver "%s"`, step.Type)
      continue
    }

    // TODO think about best error handling
    err := driver.Start(ctx, step)
    if err != nil {
      log.Println("error starting step %s: %s", step.ID, err)
    }
	}
}
