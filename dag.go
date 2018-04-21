package ktl

import (
	"fmt"
)

/* TODO
- create new version of DAG by modifying a step
- manually invalidate a step
- when an intermediate, finished step is invalidated by a new version,
  how are the downstream steps invalidated?
  - how are active, invalidated, downstream steps stopped?

- change links between steps in dag?
  - or just create a new dag at that point? but lose caching?

misc:
- how are task retries handled?
- if a step change version, but that version ends up creating the same outputs as
  the previous version, it's possible to optimize and sort of re-cache this new version.
  how would this work? Does a step need to include its output hashes in its verison hash?
*/

type DAG struct {
	Steps      map[string]*Step
	Upstream   map[*Step][]*Step
	Downstream map[*Step][]*Step
}

// NewDAG builds a new DAG datastructure from the given batch's steps.
func NewDAG(steps []*Step) *DAG {
	d := &DAG{
		Steps:      map[string]*Step{},
		Upstream:   map[*Step][]*Step{},
		Downstream: map[*Step][]*Step{},
	}

	for _, step := range steps {
		d.AddStep(step.ID, step)
	}

	for _, step := range steps {
		for _, dep := range step.Dependencies {
			d.AddDep(step.ID, dep)
		}
	}
	return d
}

func (l *DAG) AddStep(id string, step *Step) {
	l.Steps[id] = step
}

func (l *DAG) AllSteps() []*Step {
	var steps []*Step
	for _, step := range l.Steps {
		steps = append(steps, step)
	}
	return steps
}

func (l *DAG) GetSteps(ids ...string) []*Step {
	var steps []*Step
	for _, id := range ids {
		steps = append(steps, l.Steps[id])
	}
	return steps
}

func (l *DAG) AddDep(stepID, depID string) error {
	dep, ok := l.Steps[depID]
	if !ok {
		return fmt.Errorf(`missing dependency "%s"`, depID)
	}

	step, ok := l.Steps[stepID]
	if !ok {
		return fmt.Errorf(`missing step "%s"`, stepID)
	}

	l.Upstream[step] = append(l.Upstream[step], dep)
	l.Downstream[dep] = append(l.Downstream[dep], step)
	return nil
}

func AllUpstream(dag *DAG, step *Step) []*Step {
	var upstream []*Step
	for _, up := range dag.Upstream[step] {
		upstream = append(upstream, up)
		upstream = append(upstream, AllUpstream(dag, up)...)
	}
	return upstream
}

func AllDownstream(dag *DAG, step *Step) []*Step {
	var downstream []*Step
	for _, down := range dag.Downstream[step] {
		downstream = append(downstream, down)
		downstream = append(downstream, AllDownstream(dag, down)...)
	}
	return downstream
}

func Terminals(dag *DAG, steps []*Step) []*Step {
	var terminals []*Step
	for _, step := range steps {
		if len(dag.Downstream[step]) == 0 {
			terminals = append(terminals, step)
		}
	}
	return terminals
}
