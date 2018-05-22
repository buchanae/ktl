package ktl

import (
	"context"
	"fmt"
	"log"
	"time"
)

// DriverSpec is used to pass limited step information to drivers,
// so it's more clear what information a driver is expected to access/modify.
type DriverSpec struct {
	Config interface{}
	Logs   interface{}
}

type CheckResult struct {
	State  State
	Reason string
}

// Driver is the interface fulfilled by a step driver.
// Drivers are responsible for managing the state of a step.
// There are many types of drivers: start a task, wait for an event, etc.
type Driver interface {
	Check(context.Context, *DriverSpec) (*CheckResult, error)
	Start(context.Context, *DriverSpec) error
	Stop(context.Context, *DriverSpec) error
}

type Processor struct {
	db      Database
	drivers map[string]Driver
}

func NewProcessor(db Database, drivers map[string]Driver) *Processor {
	return &Processor{db: db, drivers: drivers}
}

// Process is the main control loop, responsible for managing the state
// of batches and their steps. Process periodically checks for active batches
// and calls the step drivers to manage step state.
func (p *Processor) Process(ctx context.Context) error {

	// Get all active batches
	// TODO pagination
	// TODO separate batch list from main control code below
	batches, err := p.db.ListBatches(ctx, &BatchListOptions{})
	if err != nil {
		return fmt.Errorf("listing batches: %s", err)
	}

	// For each batch, reconcile any state changes.
	for _, batch := range batches {

		for _, step := range batch.Steps {
			err := p.checkActual(ctx, step)
			if err != nil {
				log.Println("error checking step:", err)
				continue
			}

			// Stop old versions of steps
			for _, old := range step.History {
				err := p.checkActual(ctx, old)
				if err != nil {
					log.Println("error checking step:", err)
					continue
				}
				if !old.State.Done() {
					old.State = Failed
				}
				err = p.reconcile(ctx, old)
				if err != nil {
					log.Println(err)
				}
			}
		}

		// Process the batch's DAG and update the step states.
		// e.g. when a step's dependencies are ready, mark the step as active.
		processDAG(batch.Steps, batch.Mode)

		// TODO concurrency checks
		err = p.db.UpdateBatch(ctx, batch)
		if err != nil {
			log.Println("error updating batch state:", err)
			continue
		}

		// Reconcile the state of the steps with the driver entities
		// e.g. start/stop a task
		for _, step := range batch.Steps {
			err := p.reconcile(ctx, step)
			if err != nil {
				log.Println(err)
			}
		}

		// TODO concurrency checks
		err = p.db.UpdateBatch(ctx, batch)
		if err != nil {
			log.Println("error updating batch state after driving:", err)
			continue
		}
	}
	return nil
}

func (p *Processor) checkActual(ctx context.Context, step *Step) error {
	driver, ok := p.drivers[step.Type]
	if !ok {
		return fmt.Errorf("unknown step driver type: %s", step.Type)
	}

	spec := &DriverSpec{
		Config: step.Config,
		Logs:   step.Logs,
	}

	res, err := driver.Check(ctx, spec)
	if err != nil {
		return fmt.Errorf("checking step %s: %s", step.ID, err)
	}

	step.actual = res
	// TODO this doesn't belong here in the long run.
	p.processTimeLimits(step)
	return nil
}

func (p *Processor) processTimeLimits(s *Step) {

	// TODO while the process loop might happen often, it might not, and if it's slow
	//      deadlines and timeouts might be imprecise. would be nice to have a system
	//      with high precision.
	//
	//      At some point, this should probably move to an independent, async controller.

	// Check the step deadline.
	if s.Deadline != nil && s.Deadline.Sub(time.Now()) < 0 {
		s.State = Failed
		s.Reason = "deadline exceeded"

		// Check the s timeout.
	} else if s.StartedAt != nil && s.Timeout > 0 &&
		time.Now().Sub(*s.StartedAt) > time.Duration(s.Timeout) {

		s.State = Failed
		s.Reason = "timeout exceeded"
	}
}

func (p *Processor) reconcile(ctx context.Context, step *Step) error {
	driver, ok := p.drivers[step.Type]
	if !ok {
		return fmt.Errorf("unknown step driver type: %s", step.Type)
	}
	if step.actual == nil {
		return nil
	}
	actual := step.actual.State

	spec := &DriverSpec{
		Config: step.Config,
		Logs:   step.Logs,
	}

	switch {
	case step.State == Active && actual == Waiting:
		err := driver.Start(ctx, spec)
		if err != nil {
			return err
		}
		now := time.Now()
		step.StartedAt = &now
		step.Logs = spec.Logs

	case step.State == Failed && actual == Active:
		err := driver.Stop(ctx, spec)
		if err != nil {
			return err
		}

	case step.State == Active && actual == Failed:
		step.State = Failed

	case step.State == Active && actual == Success:
		step.State = Success
	}

	return nil
}

// processDAG processes the steps in a DAG, determining which steps are ready to run.
func processDAG(steps []*Step, mode Mode) {

	d := NewDAG(steps)
	all := d.AllSteps()
	failed := FilterByState(all, Failed)

	if AllDone(all) {
		return
	}

	// In fail fast mode, the processing stops one the first error encountered,
	// stopping any steps which are running.
	if mode == FailFast && failed != nil {

		// Stop any active steps.
		for _, step := range steps {
			if step.State == Active {
				step.State = Failed
				step.Reason = "failed fast"
			}
		}
		return
	}

	ready := Ready(d, all)
	active := FilterByState(all, Active)
	terminals := Terminals(d, all)
	blockers := Blockers(d, terminals)

	// If all terminal steps are blocked, and nothing is running, and nothing is ready,
	// there's nothing to do. If all the remaining steps are blocked, we consider
	// the batch as failed.
	if active == nil && ready == nil && AllBlocked(d, terminals) &&
		AllState(blockers, Failed) {
		return
	}

	// Execute next steps.
	for _, step := range ready {
		step.State = Active
	}
}
