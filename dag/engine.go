
package dag

import (
  //"sync"
)

type DAGEngine interface {
  Start() chan Event
}

type MemoryDAG struct {
  requests chan Step
  events chan Event

  steps  map[string]Step
  states map[string]EventType
  rev    map[string][]string
}


func NewMemoryDAG(requests chan Step, events chan Event) DAGEngine {
  return &MemoryDAG{
    requests:requests,
    events:events,
    steps:map[string]Step{},
    states:map[string]EventType{},
    rev:map[string][]string{},
  }
}

func (self *MemoryDAG) Start() chan Event {

  out := make(chan Event, 10)
  go func() {
    for i:= range self.requests {
      self.steps[i.StepId] = i
      self.states[i.StepId] = EventType_WAITING
      for _, j := range i.Depends {
        if x, ok := self.rev[j]; ok {
          self.rev[j] = append(x, i.StepId)
        } else {
          self.rev[j] = []string{i.StepId}
        }
      }
    }
  }()

  return out

}
