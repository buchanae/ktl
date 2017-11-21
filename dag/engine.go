
package dag


type DAGEngine interface {
  SetRequestChan(steps chan Step)
  SetEventChan(evants chan Event)
  SetResultChan(results chan Event)
  
  Start()
}

type MemoryDAG struct {
  requests chan Step
  events chan Event
  results chan Event
}


func (self *MemoryDAG) Start() {
  
  
  
  
}