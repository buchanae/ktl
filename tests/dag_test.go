package dag

import (
	"fmt"
	"github.com/ohsu-comp-bio/ktl/dag"
	"github.com/ohsu-comp-bio/ktl/pbutil"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func choose(in []string, count int) []string {
	t := make(map[int32]bool, count)
	for len(t) < count && len(t) < len(in) {
		t[rand.Int31n(int32(len(in)))] = true
	}
	out := make([]string, 0, count)
	for i := range t {
		out = append(out, in[i])
	}
	return out
}

var STEP_COUNT int = 10000

func TestRun(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	in_events := make(chan dag.Event, 100)

	d := dag.MemoryDAG{}

	out_events := d.Start(in_events)

	quit := make(chan bool)

	//Create Job requests
	go func() {
		step_ids := []string{}
		for i := 0; i < STEP_COUNT; i++ {
			s := fmt.Sprintf("event_%d", i)
			step_ids = append(step_ids, s)
			depends := []string{}
			if i > 2 {
				dcount := int(rand.Int31n(int32((i-1)/2))) % 7
				depends = choose(step_ids, dcount)
			}
			in_events <- dag.Event{StepId: s, Event: dag.EventType_NEW, Depends: depends}
			time.Sleep(time.Duration(rand.Int63n(100)) * time.Microsecond)
		}
		in_events <- dag.Event{Event: dag.EventType_CLOSE}
	}()

	processed := make(map[string]bool, STEP_COUNT)
	processed_lock := sync.Mutex{}
	//Consume events
	go func() {
		//defer close(job)
		for i := range out_events {
			log.Printf("Event: %s\n", i)
			switch i.Event {
			case dag.EventType_READY:
				go func(step_id string) {
					time.Sleep(time.Duration(rand.Int63n(1000)) * time.Microsecond)
					processed_lock.Lock()
					processed[step_id] = true
					processed_lock.Unlock()
					in_events <- dag.Event{StepId: step_id, Event: dag.EventType_SUCCESS}
				}(i.StepId)
			default:
				log.Printf("Unknown Event: %s", i.Event)
			}
		}
		quit <- true
	}()
	log.Printf("%#v", d)
	<-quit

	//check to make sure correct number of jobs were processed
	if len(processed) != STEP_COUNT {
		for i := 0; i < STEP_COUNT; i++ {
			s := fmt.Sprintf("event_%d", i)
			if _, ok := processed[s]; !ok {
				t.Errorf("Processed %s not found", s)
			}
		}
		t.Errorf("Processed %d out of %d events", len(processed), STEP_COUNT)
	}
}


func TestInputMapping(t *testing.T) {
	in_events := make(chan dag.Event, 100)
	d := dag.MemoryDAG{}
	out_events := d.Start(in_events)

	in_events <- dag.Event{StepId: "step1", Event: dag.EventType_NEW}
	in_events <- dag.Event{StepId: "step2",
		Event: dag.EventType_NEW,
		Depends:[]string{"step1"},
		Inputs:[]*dag.InputMapping{
			&dag.InputMapping{SrcStepId:"step1",SrcParamName:"output1",ParamName:"step1_output"},
		},
	}
	in_events <- dag.Event{StepId: "step3",
		Event: dag.EventType_NEW,
		Depends:[]string{"step1", "step2"},
		Inputs:[]*dag.InputMapping{
			&dag.InputMapping{SrcStepId:"step1",SrcParamName:"output1",ParamName:"input1"},
			&dag.InputMapping{SrcStepId:"step2",SrcParamName:"output2",ParamName:"input2"},
		},
	}
	in_events <- dag.Event{Event: dag.EventType_CLOSE}

	for e := range out_events {
		if e.Event == dag.EventType_READY {
			if e.StepId == "step1" {
				in_events <- dag.Event{
					StepId: "step1",
					Event: dag.EventType_SUCCESS,
					Params: pbutil.JSONDict{"output1" : "hello world"}.AsStruct(),
				}
			} else if e.StepId == "step2" {
				if e.Params.Fields["step1_output"].GetStringValue() != "hello world" {
					t.Errorf("Incorrect Input Parameter")
				}
				log.Printf("Params: %s", e.Params)
				in_events <- dag.Event{
					StepId: "step2",
					Event: dag.EventType_SUCCESS,
					Params: pbutil.JSONDict{"output2" : "world hello"}.AsStruct(),
				}
			} else if e.StepId == "step3" {
				if e.Params.Fields["input1"].GetStringValue() != "hello world" {
					t.Errorf("Incorrect Input Parameter")
				}
				if e.Params.Fields["input2"].GetStringValue() != "world hello" {
					t.Errorf("Incorrect Input Parameter")
				}
				log.Printf("Params: %s", e.Params)
				in_events <- dag.Event{
					StepId: "step3",
					Event: dag.EventType_SUCCESS,
				}
			}
		}
	}

}
