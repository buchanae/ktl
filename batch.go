package ktl

import (
	"fmt"
	"github.com/rs/xid"
	"time"
)

/*
TODO
Batch editing/versioning
- want to mimic kubectl create/apply. want to manage a batch by recursively scanning
  a directory of declarations. but, merge-based updates will be hard.

actions
- force invalidate step manually
- modify step state (e.g. cancel/stop)
- modify DAG

step exec history

tasks:
- retries
- manual restart
- canned tasks: run notebook, etc.

configuration and cli/env

dashboard

step types:
- task which is able to retry, gradually requesting more resources
- add existing tasks, without creation
- wait for github PR to be merged.
- task array
- wait for file
- wait for time
- wait for query change
- somehow use task stdout to generate next step?
- wait for task
- task that waits for input files
-- after a task has finished, if the file disappears, how can ktl convey this usefully?

- wait for event
-- what happens when the event comes in twice? and the last is still running?
   map to separate tasks? restart?
   -- this is more like a task that starts a new batch. some code, whether template
      driven or not, would need to map the event data to a task/workflow/batch.
      dynamic batches (batches that can modify themselves)?

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

	//Counts dag.Counts `json:"counts"`
	History []*Step `json:"-"`
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

	// Timeout is used to require the a step finish within a given time frame,
	// starting when step is started.
	Timeout Duration `json:"timeout,omitempty"`
	// Deadline is used to require that a step finish before a certain time.
	Deadline *time.Time `json:"deadline,omitempty"`

	State State `json:"state"`
	// Reason describes why the step is in its state,
	// usually it describes an error message.
	Reason    string     `json:"reason,omitempty"`
	StartedAt *time.Time `json:"startedAt,omitempty"`
	Version   int        `json:"version"`

	// Config contains opaque, driver-specific data which each type of step
	// driver uses to define the details of the step.
	Config interface{} `json:"config,omitempty"`
	// Logs holds opaque, driver-specific log data.
	Logs interface{} `json:"logs,omitempty"`
}

// Error returns an error if the step failed, with step.Reason
// as the message, otherwise it returns nil.
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
