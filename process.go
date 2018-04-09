package ktl

import (
	"context"
	"github.com/ohsu-comp-bio/ktl/dag"
	"log"
	"time"
)

const (
	Start EventType = iota
	Stop
)

// EventType enumerates the types of events available.
type EventType int

//go:generate enumer -type=EventType -text

// Event describes an event occurring during the lifetime of a step,
// usually due to a state change, such as "start" or "stop".
// Events are used while processing step drivers.
type Event struct {
	Type      EventType `json:"type"`
	CreatedAt time.Time `json:"createdAt"`
	// Processed is set to true when the event has been successfully
	// processed by the step driver.
	Processed bool `json:"processed"`
	// Attempts records the number of times event processing has been
	// attempted. The batch processor may decide to give up if the
	// event has been attempted too many times.
	Attempts int `json:"attempts"`
}

// NewEvent creates a new event of the given type, with the timestamp
// set to now.
func NewEvent(t EventType) Event {
	return Event{Type: t, CreatedAt: time.Now()}
}

// DriverSpec is used to pass limited step information to drivers,
// so it's more clear what information a driver is expected to access/modify.
type DriverSpec struct {
	BatchID string
	StepID  string
	State   State
	Reason  string
	Config  interface{}
	Logs    interface{}
}

// Driver is the interface fulfilled by a step driver.
// Drivers are responsible for managing the state of a step.
// There are many types of drivers: start a task, wait for an event, etc.
type Driver interface {
	Check(context.Context, *DriverSpec) error
	Start(context.Context, *DriverSpec) error
	Stop(context.Context, *DriverSpec) error
}

// Process is the main control loop, responsible for managing the state
// of batches and their steps. Process periodically checks for active batches
// and calls the step drivers to manage step state.
func Process(db Database, drivers map[string]Driver) {
	ctx := context.Background()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {

		// Get all active batches
		batches, err := db.ListBatches(ctx, &BatchListOptions{
			State: []State{Waiting, Ready, Active},
		})
		if err != nil {
			log.Println("error listing batches", err)
			continue
		}

		// For each batch, reconcile any state changes.
		for _, batch := range batches {
			processBatch(ctx, batch)

			err = db.UpdateBatch(ctx, batch)
			if err != nil {
				log.Println("error updating batch state:", err)
			}
		}
	}
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
			step.Events = append(step.Events, Stop)
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

	// TODO check how this behaves with blocked nodes.
	d := BatchDAG(batch)
	if dag.AllDone(d.AllNodes()) {
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

	// Execute next steps.
	ready := dag.Ready(d, d.AllNodes())
	for _, node := range ready {
		step := node.(*Step)
		step.State = Ready
		step.Events = append(step.Events, NewEvent(Start))
	}
}
