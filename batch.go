package ktl

import (
	"github.com/ohsu-comp-bio/ktl/dag"
	"github.com/rs/xid"
  "fmt"
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

tasks:
- retries
- manual restart
- canned tasks: run notebook, etc.

step types:
- wait for file
- wait for query change
- somehow use task stdout to generate next step?
- wait for task
- wait for event
-- what happens when the event comes in twice? and the last is still running?
   map to separate tasks? restart?
   -- this is more like a task that starts a new batch. some code, whether template
      driven or not, would need to map the event data to a task/workflow/batch.
      dynamic batches (batches that can modify themselves)?
- task with timeout
- curl/import
- galaxy
- cwl

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

	Timeout  time.Duration `json:"timeout,omitempty"`
	Deadline *time.Time    `json:"deadline,omitempty"`

	// Status
	State     State       `json:"state"`
  Reason string      `json:"reason,omitempty"`
	StartedAt *time.Time  `json:"startedAt,omitempty"`
	Logs      interface{} `json:"logs,omitempty"`
}

func (s *Step) Done() bool {
	return s.State == Success || s.State == Failed
}

func (s *Step) Running() bool {
	return s.State == Running
}

func (s *Step) Error() error {
  if s.State == Failed {
    return fmt.Errorf(s.Reason)
  }
  return nil
}

func NewBatchID() string {
	return xid.New().String()
}

// CreateBatchResponse is returned from the CreateBatch API endpoint,
// describing the ID of the batch created.
type CreateBatchResponse struct {
	ID string `json:"id"`
}

type BatchListOptions struct {
	State    []State           `json:"state"`
	Tags     map[string]string `json:"tags"`
	Page     string            `json:"page"`
	PageSize int               `json:"pageSize"`
}

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

func UpdateBatchCounts(batch *Batch) {
	d := BatchDAG(batch)
	batch.Counts = dag.Count(d, d.AllNodes())
}
