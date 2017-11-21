

package dag

import (
  "log"
  "fmt"
  "math/rand"
  "testing"
  "github.com/ohsu-comp-bio/ktl/dag"
)


func TestRun(t *testing.T) {

  requests := make(chan dag.Step, 100)
  events := make(chan dag.Event, 100)

  d := dag.NewMemoryDAG(requests, events)

  d.Start()

  step_ids := []string{}
  for i := 0; i < 100; i++ {
    step_ids = append(step_ids, fmt.Sprintf("event_%d", i))
    if i > 2 {
      dcount := rand.Int31n(int32(i/2))
      fmt.Printf("%d\n", dcount)
    }
  }

  requests <- dag.Step{StepId:"event1"}
  requests <- dag.Step{StepId:"event2"}

  log.Printf("%#v", d)

  close(requests)
  close(events)

}
