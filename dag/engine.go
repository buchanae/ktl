
package dag

import (
  //"log"
)

type DAGEngine interface {
  Start(chan Event) chan Event
}

type MemoryDAG struct {
  steps  map[string]Step
  states map[string]EventType
  deps   map[string][]string
  rev    map[string][]string
  out    chan Event
}


func (self *MemoryDAG) process_NEW(i Event) {
  //log.Printf("%#v", i)
  self.steps[i.StepId] = Step{StepId:i.StepId, Depends:i.Depends}
  depends := []string{}
  in_error := false
  for _, d := range i.Depends {
    if x, ok := self.states[d]; ok {
      switch x {
      case EventType_NEW, EventType_READY:
        depends = append(depends, d)
      case EventType_UNKNOWN, EventType_FAILURE, EventType_CANCELED:
        in_error = true
      case EventType_SUCCESS:
      }
    }
  }
  if in_error {
    self.states[i.StepId] = EventType_FAILURE
    self.out <- Event{StepId:i.StepId, Event:EventType_FAILURE}          
  } else {
    if len(depends) > 0 {
      self.states[i.StepId] = EventType_NEW
      self.deps[i.StepId] = depends
      for _, j := range depends {
        if x, ok := self.rev[j]; ok {
          self.rev[j] = append(x, i.StepId)
        } else {
          self.rev[j] = []string{i.StepId}
        }
      }
    } else {
      self.states[i.StepId] = EventType_READY
      self.out <- Event{StepId:i.StepId, Event:EventType_READY}
    }
  }
}


func remove(in []string, s string) []string {
  out := make([]string, 0, len(in))
  for _, x := range in {
    if x != s {
      out = append(out, x)
    }
  }
  return out
}

func (self *MemoryDAG) process_SUCCESS(i Event) {
  if x, ok := self.rev[i.StepId]; ok {
    for _, d := range x {
      nd := remove(self.deps[d], i.StepId)
      if len(nd) == 0 {
        delete(self.deps, d)
        self.out <- Event{StepId:i.StepId, Event:EventType_READY}
      } else {
        self.deps[d] = nd
      }
    }
    delete(self.rev, i.StepId)
  }
}

func (self *MemoryDAG) Start(input chan Event) chan Event {
  self.steps = map[string]Step{}
  self.states = map[string]EventType{}
  self.rev = map[string][]string{}
  self.deps = map[string][]string{}
  self.out = make(chan Event, 10)
  
  go func() {
    defer close(self.out)
    for i:= range input {
      switch i.Event {
      case EventType_NEW:
        self.process_NEW(i)
      case EventType_SUCCESS:
        self.process_SUCCESS(i)
      }
    }
  }()
  return self.out
}
