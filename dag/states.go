package dag

const (
	Waiting State = iota
	Paused
	Active
	Success
	Failed
)

// State describes the state of a step.
type State int

// Done returns true if the state is Success or Failed.
func (s State) Done() bool {
	return s == Success || s == Failed
}

//go:generate enumer -type=State -text


func FilterByState(nodes []Node, states ...State) []Node {
  var filtered []Node
	for _, node := range nodes {
    for _, state := range states {
      if node.NodeState() == state {
        filtered = append(filtered, node)
        break
      }
    }
  }
  return filtered
}

// AllState returns true if all the given nodes have the given state.
func AllState(nodes []Node, state State) bool {
  if len(nodes) == 0 {
    return false
  }
  for _, node := range nodes {
    if node.NodeState() != state {
      return false
    }
  }
  return true
}

// AllDone returns true if all of the nodes are done.
// See Done() and State.Done().
func AllDone(nodes []Node) bool {
  if len(nodes) == 0 {
    return false
  }
	for _, node := range nodes {
    if !node.NodeState().Done() {
			return false
		}
	}
	return true
}

// Done returns nodes which are done,
// i.e. they have a state of either Success of Failed.
func Done(nodes []Node) []Node {
	var done []Node
	for _, node := range nodes {
    if node.NodeState().Done() {
			done = append(done, node)
		}
	}
	return done
}

// Ready returns nodes which are ready to be run. See IsReady().
func Ready(dag *DAG, nodes []Node) []Node {
	var ready []Node
	for _, node := range nodes {
		if IsReady(dag, node) {
			ready = append(ready, node)
		}
	}
	return ready
}

// IsReady returns true if a node is ready to be run.
// A node is ready to be run if it has a state of Waiting,
// and all its upstream dependencies have a Success state.
func IsReady(dag *DAG, node Node) bool {
  if node.NodeState() != Waiting {
		return false
	}
	for _, dep := range dag.Upstream[node] {
    if dep.NodeState() != Success {
			return false
		}
	}
	return true
}

// Blocked returns nodes which are blocked. See IsBlocked().
func Blocked(dag *DAG, nodes []Node) []Node {
	var blocked []Node
	for _, node := range nodes {
		if IsBlocked(dag, node) {
			blocked = append(blocked, node)
		}
	}
	return blocked
}

// IsBlocked returns true if the given node is blocked.
// A node is blocked when it has a state of Waiting and one of its upstream dependencies
// has a Failed or Paused state.
//
// IsBlocked is different than !IsReady: a blocked node cannot be run because one of its
// dependencies has failed or is paused, while a node that isn't ready might only be 
// waiting on a dependency to complete.
//
// TODO allow Paused nodes to be blocked?
func IsBlocked(dag *DAG, node Node) bool {
  if node.NodeState() != Waiting {
    return false
  }
	for _, upstream := range AllUpstream(dag, node) {
    state := upstream.NodeState()
    if state == Failed || state == Paused {
			return true
		}
	}
	return false
}

func Blockers(dag *DAG, nodes []Node) []Node {
  track := map[Node]bool{}
  for _, node := range nodes {
    for _, upstream := range AllUpstream(dag, node) {
      state := upstream.NodeState()
      if state == Failed || state == Paused {
        track[upstream] = true
      }
    }
  }
  var blockers []Node
  for node, _ := range track {
    blockers = append(blockers, node)
  }
  return blockers
}

func AllBlocked(dag *DAG, nodes []Node) bool {
  if len(nodes) == 0 {
    return false
  }
  for _, node := range nodes {
    if !IsBlocked(dag, node) {
      return false
    }
  }
  return true
}

type Categories struct {
  Waiting, Paused, Active, Success, Failed, Ready, Done, Blocked []Node
}

func Categorize(dag *DAG, nodes []Node) Categories {
	return Categories{
    Waiting: FilterByState(nodes, Waiting),
    Paused: FilterByState(nodes, Paused),
    Active: FilterByState(nodes, Active),
    Success: FilterByState(nodes, Success),
    Failed: FilterByState(nodes, Failed),
		Ready:   Ready(dag, nodes),
		Done:    Done(nodes),
		Blocked: Blocked(dag, nodes),
	}
}

type Counts struct {
	Total   int `json:"total"`
	Waiting    int `json:"waiting"`
	Paused  int `json:"paused"`
	Active  int `json:"active"`
  Success int `json:"success"`
  Failed int `json:"failed"`
	Ready   int `json:"ready"`
	Done    int `json:"done"`
	Blocked int `json:"blocked"`
}

func Count(dag *DAG, nodes []Node) Counts {
	c := Categorize(dag, nodes)
	return Counts{
		Total:   len(nodes),
		Waiting:    len(c.Waiting),
		Paused:  len(c.Paused),
		Active:  len(c.Active),
    Success: len(c.Success),
    Failed: len(c.Failed),
		Ready:   len(c.Ready),
		Done:    len(c.Done),
		Blocked: len(c.Blocked),
	}
}
