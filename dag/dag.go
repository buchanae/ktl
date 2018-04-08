package dag

import (
	"fmt"
)

/* TODO
- create new version of DAG by modifying a node
- manually invalidate a node
- when an intermediate, finished node is invalidated by a new version,
  how are the downstream nodes invalidated?
  - how are running, invalidated, downstream nodes stopped?

- change links between nodes in dag?
  - or just create a new dag at that point? but lose caching?

misc:
- how are task retries handled?
- timeouts
- if a node change version, but that version ends up creating the same outputs as
  the previous version, it's possible to optimize and sort of re-cache this new version.
  how would this work? Does a node need to include its output hashes in its verison hash?
*/

// TODO serializable DAG

type Node interface {
	Done() bool
	Running() bool
	Error() error

	//Start()
	//Stop()
	//Reset()
}

type DAG struct {
  Nodes map[string]Node
	Upstream   map[Node][]Node
	Downstream map[Node][]Node
}

func NewDAG() *DAG {
  return &DAG{
    Nodes: map[string]Node{},
    Upstream: map[Node][]Node{},
    Downstream: map[Node][]Node{},
  }
}

func (l *DAG) AddNode(id string, node Node) {
  l.Nodes[id] = node
}

func (l *DAG) AllNodes() []Node {
  var nodes []Node
  for _, node := range l.Nodes {
    nodes = append(nodes, node)
  }
  return nodes
}

func (l *DAG) GetNodes(ids ...string) []Node {
  var nodes []Node
  for _, id := range ids {
    nodes = append(nodes, l.Nodes[id])
  }
  return nodes
}

func (l *DAG) AddDep(nodeID, depID string) error {
  dep, ok := l.Nodes[depID]
  if !ok {
    return fmt.Errorf(`missing dependency "%s"`, depID)
  }

  node, ok := l.Nodes[nodeID]
  if !ok {
    return fmt.Errorf(`missing node "%s"`, nodeID)
  }

  l.Upstream[node] = append(l.Upstream[node], dep)
  l.Downstream[dep] = append(l.Downstream[dep], node)
  return nil
}

func Idle(nodes []Node) []Node {
	var idle []Node
	for _, node := range nodes {
		if !node.Done() && !node.Running() {
			idle = append(idle, node)
		}
	}
	return idle
}

func Running(nodes []Node) []Node {
	var running []Node
	for _, node := range nodes {
		if node.Running() {
			running = append(running, node)
		}
	}
	return running
}

func AllDone(nodes []Node) bool {
	for _, node := range nodes {
		if !node.Done() {
			return false
		}
	}
	return true
}

func Done(nodes []Node) []Node {
	var done []Node
	for _, node := range nodes {
		if node.Done() {
			done = append(done, node)
		}
	}
	return done
}

func Failed(nodes []Node) []Node {
	var failed []Node
	for _, node := range nodes {
		if node.Error() != nil {
			failed = append(failed, node)
		}
	}
	return failed
}

func Errors(nodes []Node) error {
	var errors []error
	for _, node := range nodes {
		if err := node.Error(); err != nil {
			errors = append(errors, err)
		}
	}
  if errors == nil {
    return nil
  }
	return &ErrorList{errors}
}

func AllUpstream(dag *DAG, node Node) []Node {
	var upstream []Node
	for _, up := range dag.Upstream[node] {
		upstream = append(upstream, up)
		upstream = append(upstream, AllUpstream(dag, up)...)
	}
	return upstream
}

func AllDownstream(dag *DAG, node Node) []Node {
	var downstream []Node
	for _, down := range dag.Downstream[node] {
		downstream = append(downstream, down)
		downstream = append(downstream, AllDownstream(dag, down)...)
	}
	return downstream
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
		if upstream.Error() != nil {
			return true
		}
	}
	return false
}

func IsReady(dag *DAG, node Node) bool {
	if node.Done() || node.Running() {
		return false
	}
	for _, dep := range dag.Upstream[node] {
		if !dep.Done() || dep.Error() != nil {
			return false
		}
	}
	return true
}

func Terminals(dag *DAG, nodes []Node) []Node {
	var terminals []Node
	for _, node := range nodes {
		if len(dag.Downstream[node]) == 0 {
			terminals = append(terminals, node)
		}
	}
	return terminals
}

type Categories struct {
	Idle,
	Ready,
	Running,
	Done,
	Blocked,
	Failed []Node
}

func Categorize(dag *DAG, nodes []Node) Categories {
	return Categories{
		Idle:    Idle(nodes),
		Ready:   Ready(dag, nodes),
		Running: Running(nodes),
		Done:    Done(nodes),
		Blocked: Blocked(dag, nodes),
		Failed:  Failed(nodes),
	}
}

type Counts struct {
	Total int `json:"total"`
	Idle int `json:"idle"`
	Ready int `json:"ready"`
	Running int `json:"running"`
	Done int `json:"done"`
	Blocked int `json:"blocked"`
	Failed int `json:"failed"`
}

func Count(dag *DAG, nodes []Node) Counts {
	c := Categorize(dag, nodes)
	return Counts{
		Total:   len(nodes),
		Idle:    len(c.Idle),
		Ready:   len(c.Ready),
		Running: len(c.Running),
		Done:    len(c.Done),
		Blocked: len(c.Blocked),
		Failed:  len(c.Failed),
	}
}

// TODO FailFast is more interesting if it stops/cancels
//      running tasks on the first error.
// TODO need to check Blocked or something to
// ensure that if no nodes are ready, will there ever be one
// ready?
func FailFast(dag *DAG, nodes []Node) ([]Node, error) {
	errs := Errors(nodes)
	if errs != nil {
    return nil, errs
	}
	return Ready(dag, nodes), nil
}

var ErrBlocked = fmt.Errorf("all remaining nodes are blocked")

// TODO need to check Blocked or something to
// ensure that if no nodes are ready, will there ever be one
// ready?
func BestEffort(dag *DAG, nodes []Node) ([]Node, error) {
  return Ready(dag, nodes), nil
}

type ErrorList struct {
	Errors []error
}

func (e *ErrorList) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}

	s := "Errors:"
	for _, err := range e.Errors {
		s += "- " + err.Error()
	}
	return s
}
