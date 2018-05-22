package ktl

func FilterByState(steps []*Step, states ...State) []*Step {
	var filtered []*Step
	for _, step := range steps {
		for _, state := range states {
			if step.State == state {
				filtered = append(filtered, step)
				break
			}
		}
	}
	return filtered
}

// AllState returns true if all the given steps have the given state.
func AllState(steps []*Step, state State) bool {
	if len(steps) == 0 {
		return false
	}
	for _, step := range steps {
		if step.State != state {
			return false
		}
	}
	return true
}

// AllDone returns true if all of the steps are done.
// See Done() and State.Done().
func AllDone(steps []*Step) bool {
	if len(steps) == 0 {
		return false
	}
	for _, step := range steps {
		if !step.State.Done() {
			return false
		}
	}
	return true
}

// Done returns steps which are done,
// i.e. they have a state of either Success of Failed.
func Done(steps []*Step) []*Step {
	var done []*Step
	for _, step := range steps {
		if step.State.Done() {
			done = append(done, step)
		}
	}
	return done
}

// Ready returns steps which are ready to be run. See IsReady().
func Ready(dag *DAG, steps []*Step) []*Step {
	var ready []*Step
	for _, step := range steps {
		if IsReady(dag, step) {
			ready = append(ready, step)
		}
	}
	return ready
}

// IsReady returns true if a step is ready to be run.
// A step is ready to be run if it has a state of Waiting,
// and all its upstream dependencies have a Success state.
func IsReady(dag *DAG, step *Step) bool {
	if step.State != Waiting {
		return false
	}
	for _, dep := range dag.Upstream[step] {
		if dep.State != Success {
			return false
		}
	}
	return true
}

// Blocked returns steps which are blocked. See IsBlocked().
func Blocked(dag *DAG, steps []*Step) []*Step {
	var blocked []*Step
	for _, step := range steps {
		if IsBlocked(dag, step) {
			blocked = append(blocked, step)
		}
	}
	return blocked
}

// IsBlocked returns true if the given step is blocked.
// A step is blocked when it has a state of Waiting and one of its upstream dependencies
// has a Failed or Paused state.
//
// IsBlocked is different than !IsReady: a blocked step cannot be run because one of its
// dependencies has failed or is paused, while a step that isn't ready might only be
// waiting on a dependency to complete.
//
// TODO allow Paused steps to be blocked?
func IsBlocked(dag *DAG, step *Step) bool {
	if step.State != Waiting {
		return false
	}
	for _, upstream := range AllUpstream(dag, step) {
		state := upstream.State
		if state == Failed || state == Paused {
			return true
		}
	}
	return false
}

func Blockers(dag *DAG, steps []*Step) []*Step {
	track := map[*Step]bool{}
	for _, step := range steps {
		for _, upstream := range AllUpstream(dag, step) {
			if upstream.State == Failed || upstream.State == Paused {
				track[upstream] = true
			}
		}
	}
	var blockers []*Step
	for step := range track {
		blockers = append(blockers, step)
	}
	return blockers
}

func AllBlocked(dag *DAG, steps []*Step) bool {
	if len(steps) == 0 {
		return false
	}
	for _, step := range steps {
		if !IsBlocked(dag, step) {
			return false
		}
	}
	return true
}

type Categories struct {
	Waiting, Paused, Active, Success, Failed, Ready, Done, Blocked []*Step
}

func Categorize(dag *DAG, steps []*Step) Categories {
	return Categories{
		Waiting: FilterByState(steps, Waiting),
		Paused:  FilterByState(steps, Paused),
		Active:  FilterByState(steps, Active),
		Success: FilterByState(steps, Success),
		Failed:  FilterByState(steps, Failed),
		Ready:   Ready(dag, steps),
		Done:    Done(steps),
		Blocked: Blocked(dag, steps),
	}
}

type Counts struct {
	Total   int `json:"total"`
	Waiting int `json:"waiting"`
	Paused  int `json:"paused"`
	Active  int `json:"active"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
	Ready   int `json:"ready"`
	Done    int `json:"done"`
	Blocked int `json:"blocked"`
}

func Count(dag *DAG, steps []*Step) Counts {
	c := Categorize(dag, steps)
	return Counts{
		Total:   len(steps),
		Waiting: len(c.Waiting),
		Paused:  len(c.Paused),
		Active:  len(c.Active),
		Success: len(c.Success),
		Failed:  len(c.Failed),
		Ready:   len(c.Ready),
		Done:    len(c.Done),
		Blocked: len(c.Blocked),
	}
}
