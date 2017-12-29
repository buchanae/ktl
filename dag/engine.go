package dag

import (
	structpb "github.com/golang/protobuf/ptypes/struct"
	"log"
	"sync"
)

type DAGEngine interface {
	Start(chan Event) chan Event
	ActiveCount() int
	GetStep(string) Step
}

type MemoryDAG struct {
	steps       map[string]Step
	states      map[string]EventType
	deps        map[string][]string
	rev         map[string][]string
	out         chan Event
	state_mutex sync.Mutex
}

func (self *MemoryDAG) ActiveCount() int {
	self.state_mutex.Lock()
	i := 0
	for _, x := range self.states {
		switch x {
		case EventType_NEW, EventType_READY, EventType_RUNNING:
			i += 1
		}
	}
	self.state_mutex.Unlock()
	return i
}

func (self *MemoryDAG) GetStep(id string) Step {
	return self.steps[id]
}

func (self *MemoryDAG) process_NEW(i Event) {
	log.Printf("Process New: %#v", i)
	self.steps[i.StepId] = Step{StepId: i.StepId, Depends: i.Depends, Inputs: i.Inputs}
	depends := []string{}
	in_error := false
	for _, d := range i.Depends {
		self.state_mutex.Lock()
		if x, ok := self.states[d]; ok {
			switch x {
			case EventType_NEW, EventType_READY, EventType_RUNNING:
				depends = append(depends, d)
			case EventType_UNKNOWN, EventType_FAILURE, EventType_CANCELED:
				in_error = true
			case EventType_SUCCESS:
			}
		} else {
			log.Printf("Depends %s Not Found: %s", i.StepId, d)
		}
		self.state_mutex.Unlock()
	}
	if in_error {
		self.state_mutex.Lock()
		self.states[i.StepId] = EventType_FAILURE
		self.state_mutex.Unlock()
		self.out <- Event{StepId: i.StepId, Event: EventType_FAILURE}
	} else {
		if len(depends) > 0 {
			log.Printf("Depends %s : %s", i.StepId, depends)
			self.state_mutex.Lock()
			self.states[i.StepId] = EventType_NEW
			self.state_mutex.Unlock()
			self.deps[i.StepId] = depends
			for _, j := range depends {
				if x, ok := self.rev[j]; ok {
					self.rev[j] = append(x, i.StepId)
				} else {
					self.rev[j] = []string{i.StepId}
				}
			}
		} else {
			self.state_mutex.Lock()
			self.states[i.StepId] = EventType_RUNNING
			self.state_mutex.Unlock()
			self.out <- Event{StepId: i.StepId, Event: EventType_READY}
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

func (self *MemoryDAG) start_Step(stepId string) {
	self.states[stepId] = EventType_RUNNING
	log.Printf("Starting: %s", stepId)
	params := structpb.Struct{Fields: map[string]*structpb.Value{}}
	for _, i := range self.steps[stepId].Inputs {
		res := self.steps[i.SrcStepId].Results
		if res != nil {
			params.Fields[i.ParamName] = res.Fields[i.SrcParamName]
		}
	}
	log.Printf("Step %s Params: %s", stepId, params)
	self.out <- Event{StepId: stepId, Event: EventType_READY, Params: &params}
}

func (self *MemoryDAG) process_SUCCESS(i Event) {
	log.Printf("success: %s", i.StepId)
	self.state_mutex.Lock()
	self.states[i.StepId] = EventType_SUCCESS
	s := self.steps[i.StepId]
	s.Results = i.Params
	self.steps[i.StepId] = s
	log.Printf("Step %s Results: %s", i.StepId, s.Results)
	self.state_mutex.Unlock()
	//check for steps that were dependent on this one, that can now be started
	if x, ok := self.rev[i.StepId]; ok {
		log.Printf("Success: %s resolves dependency for %s", i.StepId, x)
		for _, d := range x {
			nd := remove(self.deps[d], i.StepId)
			if len(nd) == 0 {
				delete(self.deps, d)
				self.start_Step(d)
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
	self.state_mutex = sync.Mutex{}

	go func() {
		quit := false
		defer close(self.out)
		for i := range input {
			switch i.Event {
			case EventType_NEW:
				self.process_NEW(i)
			case EventType_SUCCESS:
				self.process_SUCCESS(i)
			case EventType_CLOSE:
				log.Printf("Starting Close")
				quit = true
			default:
				log.Printf("Unknown Event")
			}
			if quit {
				if self.ActiveCount() == 0 {
					close(input)
				}
			}
		}
	}()
	return self.out
}