package dag

import (
	"fmt"
)

/* TODO
- create new version of DAG by modifying a node
- manually invalidate a node
- when an intermediate, finished node is invalidated by a new version,
  how are the downstream nodes invalidated?
  - how are active, invalidated, downstream nodes stopped?

- change links between nodes in dag?
  - or just create a new dag at that point? but lose caching?

misc:
- how are task retries handled?
- if a node change version, but that version ends up creating the same outputs as
  the previous version, it's possible to optimize and sort of re-cache this new version.
  how would this work? Does a node need to include its output hashes in its verison hash?
*/

type Node interface {
	DAGNodeState() State
}

type DAG struct {
	Nodes      map[string]Node
	Upstream   map[Node][]Node
	Downstream map[Node][]Node
}

func NewDAG() *DAG {
	return &DAG{
		Nodes:      map[string]Node{},
		Upstream:   map[Node][]Node{},
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

func Terminals(dag *DAG, nodes []Node) []Node {
	var terminals []Node
	for _, node := range nodes {
		if len(dag.Downstream[node]) == 0 {
			terminals = append(terminals, node)
		}
	}
	return terminals
}
