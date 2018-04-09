package dag

import (
	"fmt"
	"github.com/kr/pretty"
	"reflect"
	"testing"
)

func TestIdle(t *testing.T) {
	nodes := []Node{
		tnode{"01", false, false, nil},
		tnode{"02", true, false, nil},
		tnode{"03", false, true, nil},
		tnode{"04", false, false, nil},
		tnode{"05", false, false, nil},
	}
	expected := []Node{nodes[0], nodes[3], nodes[4]}
	idle := Idle(nodes)

	if !reflect.DeepEqual(idle, expected) {
		t.Error("unexpected idle")
		pretty.Ldiff(t, idle, expected)
	}
}

func TestRunning(t *testing.T) {
	nodes := []Node{
		tnode{"01", false, false, nil},
		tnode{"02", true, false, nil},
		tnode{"03", false, true, nil},
		tnode{"04", false, false, nil},
		tnode{"05", false, false, nil},
	}
	expected := []Node{nodes[2]}
	running := Running(nodes)

	if !reflect.DeepEqual(running, expected) {
		t.Error("unexpected running")
		pretty.Ldiff(t, running, expected)
	}
}

func TestAllDone(t *testing.T) {
	nodes := []Node{
		tnode{"01", false, false, nil},
		tnode{"02", true, false, nil},
		tnode{"03", false, true, nil},
		tnode{"04", false, false, nil},
		tnode{"05", false, false, nil},
	}
	if AllDone(nodes) {
		t.Error("nodes should not be all done")
	}

	done := []Node{
		tnode{"done-01", true, false, nil},
		tnode{"done-02", true, false, nil},
		tnode{"done-03", true, false, nil},
	}
	if !AllDone(done) {
		t.Error("nodes should be all done")
	}
}

func TestDone(t *testing.T) {
	nodes := []Node{
		tnode{"01", false, false, nil},
		tnode{"02", true, false, nil},
		tnode{"03", false, true, nil},
		tnode{"04", false, false, nil},
		tnode{"05", true, false, nil},
	}
	expected := []Node{nodes[1], nodes[4]}
	done := Done(nodes)
	if !reflect.DeepEqual(done, expected) {
		t.Errorf("unexpected done")
		pretty.Ldiff(t, done, expected)
	}
}

func TestFailed(t *testing.T) {
	nodes := []Node{
		tnode{"01", false, false, nil},
		tnode{"02", true, false, nil},
		tnode{"03", false, false, fmt.Errorf("err")},
	}
	expected := []Node{nodes[2]}
	failed := Failed(nodes)

	if !reflect.DeepEqual(failed, expected) {
		t.Errorf("unexpected failed")
		pretty.Ldiff(t, failed, expected)
	}
}

func TestErrors(t *testing.T) {
	err1 := fmt.Errorf("err 1")
	err2 := fmt.Errorf("err 2")
	nodes := []Node{
		tnode{"01", false, false, nil},
		tnode{"02", true, false, nil},
		tnode{"03", false, false, err1},
		tnode{"04", false, false, err2},
		tnode{"05", false, false, err1},
	}
	expected := []error{err1, err2, err1}
	errors := Errors(nodes)

	if !reflect.DeepEqual(errors, expected) {
		pretty.Ldiff(t, errors, expected)
	}
}

func TestLinks(t *testing.T) {
	d, _ := dag1()

	// Test upstream links for node "07"
	up07 := d.Upstream[d.Nodes["07"]]
	ex07 := d.GetNodes("05", "06")
	if !reflect.DeepEqual(up07, ex07) {
		t.Error("unexpected upstream")
		pretty.Ldiff(t, up07, ex07)
	}

	// Test downstream links for "03"
	down03 := d.Downstream[d.Nodes["03"]]
	ex03 := d.GetNodes("06", "08")
	if !reflect.DeepEqual(down03, ex03) {
		t.Error("unexpected downstream")
		pretty.Ldiff(t, down03, ex03)
	}
}

func TestAllUpstream(t *testing.T) {
	d, _ := dag1()
	up := AllUpstream(d, d.Nodes["07"])
	ex := d.GetNodes("05", "01", "06", "02", "03")
	if !reflect.DeepEqual(up, ex) {
		t.Error("unexpected all upstream")
		pretty.Ldiff(t, up, ex)
	}
}

func TestAllDownstream(t *testing.T) {
	d, _ := dag1()
	dn := AllDownstream(d, d.Nodes["03"])
	ex := d.GetNodes("06", "07", "08")

	if !reflect.DeepEqual(dn, ex) {
		t.Error("unexpected all downstream")
		pretty.Ldiff(t, dn, ex)
	}
}

func TestReady(t *testing.T) {
	d, nodes := dag1()
	ready := Ready(d, nodes)
	ex := d.GetNodes("03", "04")

	if !reflect.DeepEqual(ready, ex) {
		t.Error("unexpected ready")
		pretty.Ldiff(t, ready, ex)
	}
}

func TestBlocked(t *testing.T) {
	d, nodes := dag1()
	blocked := Blocked(d, nodes)
	ex := d.GetNodes("07")

	if !reflect.DeepEqual(blocked, ex) {
		t.Error("unexpected blocked")
		pretty.Ldiff(t, blocked, ex)
	}
}

func TestTerminals(t *testing.T) {
	d, nodes := dag1()
	term := Terminals(d, nodes)
	ex := d.GetNodes("04", "07", "08")

	if !reflect.DeepEqual(term, ex) {
		t.Error("unexpected term")
		pretty.Ldiff(t, term, ex)
	}
}

func TestCounts(t *testing.T) {
	d, nodes := dag1()
	counts := Count(d, nodes)

	if counts.Total != 8 {
		t.Errorf("expected total to be 8, but got %d", counts.Total)
	}
	if counts.Idle != 4 {
		t.Errorf("expected idle to be 4, but got %d", counts.Idle)
	}
	if counts.Ready != 2 {
		t.Errorf("expected ready to be 2, but got %d", counts.Ready)
	}
	if counts.Running != 1 {
		t.Errorf("expected running to be 1, but got %d", counts.Running)
	}
	if counts.Done != 3 {
		t.Errorf("expected done to be 3, but got %d", counts.Done)
	}
	if counts.Blocked != 1 {
		t.Errorf("expected blocked to be 1, but got %d", counts.Blocked)
	}
	if counts.Failed != 1 {
		t.Errorf("expected failed to be 1, but got %d", counts.Failed)
	}
}

// Test that a node may be added to the dag multiple times with a different ID,
// as a useful mechanism for creating aliases. Uses Ready() to test this, but
// should work in general.
//
// This useful, for example, for indexing a workflow node by the IDs of the multiple
// output files it produces.
func TestAlias(t *testing.T) {
	d, nodes := dag1()
	d.AddNode("03-alias", d.GetNodes("03")[0])
	ready := Ready(d, nodes)
	ex := d.GetNodes("03", "04")

	if !reflect.DeepEqual(ready, ex) {
		t.Error("unexpected ready")
		pretty.Ldiff(t, ready, ex)
	}
}

func dag1() (*DAG, []Node) {
	/*
	   01 (D) ---05 (E) ---07
	                      /
	   02 (D) ---06 (R) --
	             /
	   03 (I) --
	       \
	        08 (I)

	   04 (I)

	   I = Idle
	   D = Done
	   R = Running
	   E = Error
	*/
	err1 := fmt.Errorf("err 1")
	nodes := []Node{
		tnode{"01", true, false, nil},
		tnode{"02", true, false, nil},
		tnode{"03", false, false, nil},
		tnode{"04", false, false, nil},
		tnode{"05", true, false, err1},
		tnode{"06", false, true, nil},
		tnode{"07", false, false, nil},
		tnode{"08", false, false, nil},
	}
	dag := NewDAG()

	for _, node := range nodes {
		t := node.(tnode)
		dag.AddNode(t.id, node)
	}
	must(dag.AddDep("05", "01"))
	must(dag.AddDep("07", "05"))
	must(dag.AddDep("07", "06"))
	must(dag.AddDep("06", "02"))
	must(dag.AddDep("06", "03"))
	must(dag.AddDep("08", "03"))
	return dag, nodes
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type tnode struct {
	id            string
	done, running bool
	err           error
}

func (ts tnode) Done() bool {
	return ts.done
}
func (ts tnode) Running() bool {
	return ts.running
}
func (ts tnode) Error() error {
	return ts.err
}
func (ts tnode) ID() string {
	return ts.id
}
