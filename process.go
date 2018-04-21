package ktl

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/ktl/dag"
	"log"
	"time"
)

// DriverSpec is used to pass limited step information to drivers,
// so it's more clear what information a driver is expected to access/modify.
type DriverSpec struct {
	BatchID string
	StepID  string
	Version int
	Config  interface{}
	Logs    interface{}
}

type CheckResult struct {
	State   State
	Reason  string
	Version int
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
			// TODO pagination
			batches, err := db.ListBatches(ctx, &BatchListOptions{})
			if err != nil {
				log.Println("error listing batches", err)
				continue
			}

			// For each batch, reconcile any state changes.
			for _, batch := range batches {
				var ctrls []*stepCtrl
				for _, step := range batch.Steps {

					driver, ok := drivers[step.Type]
					if !ok {
						// TODO want these errors in the database so they can be
						//      reported in UIs
						log.Println("unknown step driver type: %s", step.Type)
						continue
					}

					ctrl := &stepCtrl{
						Batch:  batch,
						Step:   step,
						Driver: driver,
					}
					ctrls = append(ctrls, ctrl)

					// Check step state via driver.
					err := ctrl.checkActual(ctx)
					if err != nil {
						log.Println("error checking step:", err)
					}

					ctrl.processTimeLimits()
				}

				// Process the batch's DAG and update the step states.
				// e.g. when a step's dependencies are ready, mark the step as active.
				processDAG(ctx, ctrls, batch.Mode)

				// TODO concurrency checks
				err = db.UpdateBatch(ctx, batch)
				if err != nil {
					log.Println("error updating batch state:", err)
					continue
				}

				// Reconcile the state of the steps with the driver entities
				// e.g. start/stop a task
				for _, ctrl := range ctrls {
					err := ctrl.reconcile(ctx)
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

type stepCtrl struct {
	*Step
	Batch  *Batch
	Driver Driver
	Actual *CheckResult
	Reason string
}

func (s *stepCtrl) checkActual(ctx context.Context) error {

	spec := &DriverSpec{
		BatchID: s.Batch.ID,
		StepID:  s.Step.ID,
		Version: s.Step.Version,
		Config:  s.Step.Config,
		Logs:    s.Step.Logs,
	}
	res, err := s.Driver.Check(ctx, spec)
	if err != nil {
		return fmt.Errorf("checking step %s: %s", s.Step.ID, err)
	}
	s.Actual = res
	return nil
}

func (s *stepCtrl) processTimeLimits() {

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

func (s *stepCtrl) reconcile(ctx context.Context) error {
	if s.Actual == nil {
		return nil
	}
	actual := s.Actual.State

	spec := &DriverSpec{
		BatchID: batchID,
		StepID:  step.ID,
		Version: step.Version,
		Config:  step.Config,
		Logs:    step.Logs,
	}

	// TODO check for incremented version

	switch {
	case s.State == Active && actual == Waiting:
		err := s.Driver.Start(ctx, spec)
		if err != nil {
			return err
		}

	case s.State == Failed && actual == Active:
		err := s.Driver.Stop(ctx, spec)
		if err != nil {
			return err
		}

	case s.State == Active && actual == Failed:
		s.State = Failed

	case s.State == Active && actual == Success:
		s.State = Success
	}

	return nil
}

// NodeState returns state information used by the dag library.
func (s *stepCtrl) NodeState() dag.State {
	state := s.Step.State
	if s.Actual.State.Done() {
		state = s.Actual.State
	}

	switch state {
	case Waiting:
		return dag.Waiting
	case Success:
		return dag.Success
	case Paused:
		return dag.Paused
	case Active:
		return dag.Active
	case Failed:
		return dag.Failed
	}
	return dag.Paused
}

// processDAG processes the steps in a DAG, determining which steps are ready to run.
func processDAG(ctx context.Context, steps []*stepCtrl, mode Mode) {

	d := newDAG(steps)
	all := d.AllNodes()
	failed := dag.FilterByState(all, dag.Failed)

	if dag.AllDone(all) {
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

	ready := dag.Ready(d, all)
	active := dag.FilterByState(all, dag.Active)
	terminals := dag.Terminals(d, all)
	blockers := dag.Blockers(d, terminals)

	// If all terminal steps are blocked, and nothing is running, and nothing is ready,
	// there's nothing to do. If all the remaining steps are blocked, we consider
	// the batch as failed.
	if active == nil && ready == nil && dag.AllBlocked(d, terminals) &&
		dag.AllState(blockers, dag.Failed) {
		return
	}

	// Execute next steps.
	for _, node := range ready {
		step := node.(*stepCtrl)
		step.State = Active
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
			step.Version++

			err := db.UpdateBatch(ctx, batch)
			if err != nil {
				return fmt.Errorf("updating batch: %s", err)
			}
			return nil
		}
	}
	return ErrNotFound
}

// newDAG builds a new DAG datastructure from the given batch's steps.
func newDAG(steps []*stepCtrl) *dag.DAG {
	d := dag.NewDAG()
	for _, step := range steps {
		d.AddNode(step.Step.ID, step)
	}

	for _, step := range steps {
		for _, dep := range step.Dependencies {
			d.AddDep(step.ID, dep)
		}
	}
	return d
}
