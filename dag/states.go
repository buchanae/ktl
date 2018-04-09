package dag

type State struct {
	Done, Paused, Active, Failed bool
}

func (s State) Idle() bool {
	return !s.Done && !s.Active && !s.Paused
}

func Idle(nodes []Node) []Node {
	var idle []Node
	for _, node := range nodes {
		state := node.DAGNodeState()
		if state.Idle() {
			idle = append(idle, node)
		}
	}
	return idle
}

func Active(nodes []Node) []Node {
	var active []Node
	for _, node := range nodes {
		state := node.DAGNodeState()
		if state.Active {
			active = append(active, node)
		}
	}
	return active
}

func Paused(nodes []Node) []Node {
	var paused []Node
	for _, node := range nodes {
		state := node.DAGNodeState()
		if state.Paused {
			paused = append(paused, node)
		}
	}
	return paused
}

func AllDone(nodes []Node) bool {
	for _, node := range nodes {
		state := node.DAGNodeState()
		if !state.Done {
			return false
		}
	}
	return true
}

func Done(nodes []Node) []Node {
	var done []Node
	for _, node := range nodes {
		state := node.DAGNodeState()
		if state.Done {
			done = append(done, node)
		}
	}
	return done
}

func Ready(dag *DAG, nodes []Node) []Node {
	var ready []Node
	for _, node := range nodes {
		if IsReady(dag, node) {
			ready = append(ready, node)
		}
	}
	return ready
}

func Blocked(dag *DAG, nodes []Node) []Node {
	var blocked []Node
	for _, node := range nodes {
		if IsBlocked(dag, node) {
			blocked = append(blocked, node)
		}
	}
	return blocked
}

func IsBlocked(dag *DAG, node Node) bool {
	for _, upstream := range AllUpstream(dag, node) {
		state := upstream.DAGNodeState()
		if state.Failed || state.Paused {
			return true
		}
	}
	return false
}

func IsReady(dag *DAG, node Node) bool {
	state := node.DAGNodeState()
	if !state.Idle() {
		return false
	}
	for _, dep := range dag.Upstream[node] {
		state := dep.DAGNodeState()
		if !state.Done || state.Failed {
			return false
		}
	}
	return true
}

type Categories struct {
	Idle,
	Ready,
	Active,
	Paused,
	Done,
	Blocked []Node
}

func Categorize(dag *DAG, nodes []Node) Categories {
	return Categories{
		Idle:    Idle(nodes),
		Ready:   Ready(dag, nodes),
		Active:  Active(nodes),
		Paused:  Paused(nodes),
		Done:    Done(nodes),
		Blocked: Blocked(dag, nodes),
	}
}

type Counts struct {
	Total   int `json:"total"`
	Idle    int `json:"idle"`
	Ready   int `json:"ready"`
	Active  int `json:"active"`
	Paused  int `json:"paused"`
	Done    int `json:"done"`
	Blocked int `json:"blocked"`
}

func Count(dag *DAG, nodes []Node) Counts {
	c := Categorize(dag, nodes)
	return Counts{
		Total:   len(nodes),
		Idle:    len(c.Idle),
		Ready:   len(c.Ready),
		Active:  len(c.Active),
		Paused:  len(c.Paused),
		Done:    len(c.Done),
		Blocked: len(c.Blocked),
	}
}
