package ktl

import (
	"fmt"
	"github.com/ohsu-comp-bio/ktl/dag"
	"github.com/rs/xid"
	"time"
)

/*
TODO
Batch editing/versioning

actions
- force invalidate step manually
- modify step state (e.g. cancel/stop)
- modify DAG

step exec history

state reconciler
- need something that is tolerant to errors occurring while driving steps
  e.g. if driver.Stop() fails, a system should be able to revisit this
  steps and try again. also, provides an easy way for users to modify
  the desired state of the steps. also, possibly allows drivers to
  exist as clients.

tasks:
- retries
- manual restart
- canned tasks: run notebook, etc.

configuration and cli/env

dashboard

step types:
- add existing tasks, without creation
- wait for file
- wait for time.
- wait for query change
- somehow use task stdout to generate next step?
- wait for task
- wait for event
- task that waits for input files
-- what happens when the event comes in twice? and the last is still running?
   map to separate tasks? restart?
   -- this is more like a task that starts a new batch. some code, whether template
      driven or not, would need to map the event data to a task/workflow/batch.
      dynamic batches (batches that can modify themselves)?
- curl/import
- galaxy
- cwl
- can a step run multiple times?

hard:
- secret management
  e.g. at end of smc-het workflow, have a step which uploads to OICR, which
  requires secret key
*/

// Batch describes a batch of steps to be executed. Steps may have dependencies
// on each other, forming a workflow. An execution engine handles executing
// the batch of steps and monitoring their state.
type Batch struct {
	// Metadata
	ID          string            `json:"id"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`

	// Execution config
	Steps []*Step `json:"steps"`
	Mode  Mode    `json:"mode"`

	// Status
	State  State       `json:"state"`
	Reason string      `json:"reason,omitempty"`
	Logs   interface{} `json:"logs,omitempty"`
	Counts dag.Counts  `json:"counts"`
}

// Step describes a unit of work in a Batch. There are many types of steps:
// run a task, wait for an event, etc.
type Step struct {
	ID string `json:"id"`
	// Type is used to determine which step driver will handle processing
	// the step.
	Type string `json:"type"`
	// Dependencies lists the IDs of the steps this step depends on.
	Dependencies []string `json:"dependencies,omitempty"`
	// Config contains opaque, driver-specific data which each type of step
	// driver uses to define the details of the step.
	Config interface{} `json:"config,omitempty"`

	// Timeout is used to require the a step finish within a given time frame,
	// starting when step is started.
	Timeout Duration `json:"timeout,omitempty"`
	// Deadline is used to require that a step finish before a certain time.
	Deadline *time.Time `json:"deadline,omitempty"`

	State State `json:"state"`
	// Reason describes why the step failed.
	Reason    string     `json:"reason,omitempty"`
	StartedAt *time.Time `json:"startedAt,omitempty"`
	// Logs holds opaque, driver-specific log data.
	Logs interface{} `json:"logs,omitempty"`
}

// Done returns true if the step is in a final state: success or failed.
func (s *Step) Done() bool {
	return s.State == Success || s.State == Failed
}

// Running returns true if the step state is Running.
// Mostly exists to fulfill dag.Node interface.
func (s *Step) Running() bool {
	return s.State == Running
}

// Error returns an error with Step.Reason if the step state is Failed,
// otherwise Error returns nil.
func (s *Step) Error() error {
	if s.State == Failed {
		return fmt.Errorf(s.Reason)
	}
	return nil
}

// NewBatchID generates a new, globally unique ID.
func NewBatchID() string {
	return xid.New().String()
}

// CreateBatchResponse is returned from the CreateBatch API endpoint,
// describing the ID of the batch created.
type CreateBatchResponse struct {
	ID string `json:"id"`
}

// BatchListOptions describes filters and pagination options used while
// querying for a list of batches.
type BatchListOptions struct {
	State    []State           `json:"state"`
	Tags     map[string]string `json:"tags"`
	Page     string            `json:"page"`
	PageSize int               `json:"pageSize"`
}

// GetPageSize returns a page size within the allowed range [10, 500].
func (b *BatchListOptions) GetPageSize() int {
	if b.PageSize < 0 {
		return 10
	}
	if b.PageSize > 500 {
		return 500
	}
	return b.PageSize
}

// CreateBatchResponse is returned from the CreateBatch API endpoint,
// describing the ID of the batch created.
type BatchList struct {
	Batches  []*Batch `json:"batches"`
	NextPage string   `json:"nextPage"`
}

// BatchDAG builds a new DAG datastructure from the given batch's steps.
func BatchDAG(batch *Batch) *dag.DAG {
	d := dag.NewDAG()
	for _, step := range batch.Steps {
		d.AddNode(step.ID, step)
	}

	for _, step := range batch.Steps {
		for _, dep := range step.Dependencies {
			d.AddDep(step.ID, dep)
		}
	}
	return d
}

// UpdateBatchCounts modifies the given batch, updating the Batch.Counts field.
func UpdateBatchCounts(batch *Batch) {
	d := BatchDAG(batch)
	batch.Counts = dag.Count(d, d.AllNodes())
}
