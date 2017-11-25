package engine

import (
  "log"
  "time"
	"github.com/ohsu-comp-bio/ktl/cwl"
  "github.com/ohsu-comp-bio/ktl/dag"
	"strings"
)


func (self Engine) RunWorkflow(wf cwl.Workflow, mapper cwl.FileMapper, env cwl.Environment) (cwl.JSONDict, error) {
	log.Printf("Starting Workflow")

  md := dag.MemoryDAG{}

  in_events := make(chan dag.Event, 100)
  out_events := md.Start(in_events)
  go func() {
    for e := range out_events {
      log.Printf("Out: %s", e)
    }
  }()

	steps := map[string]bool{}
	for added := true; added; {
		added = false
		for _, s := range wf.Steps {
			if _, ok := steps[s.Id]; !ok {
				deps := map[string]bool{}
				for _, i := range s.In {
					if len(i.Source) > 0 {
						p := strings.Split(i.Source[0], "/")
						if len(p) == 2 {
							deps[p[0]] = true
						}
					}
				}
				ready := true
				for d := range deps {
					if _, ok := steps[d]; !ok {
						ready = false
					}
				}
				if ready {
          da := []string{}
          for k := range deps {
            da = append(da, k)
          }
          de := dag.Event{
            StepId: s.Id,
            Event: dag.EventType_NEW,
            Depends: da,
          }
          log.Printf("Add %s", de)
          in_events <- de
					steps[s.Id] = true
					added = true
				}
			}
		}
	}

  log.Printf("ActiveCount: %d", md.ActiveCount())
  for md.ActiveCount() > 0 {
		time.Sleep(500 * time.Microsecond)
		log.Printf("ActiveCount: %d", md.ActiveCount())
	}
	close(in_events)

	return cwl.JSONDict{}, nil
}
