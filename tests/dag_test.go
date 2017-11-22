package dag

import (
	"log"
	//"sync"
	"fmt"
	"github.com/ohsu-comp-bio/ktl/dag"
	"math/rand"
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
		for d.ActiveCount() > 0 {
			time.Sleep(500 * time.Microsecond)
			log.Printf("ActiveCount: %d", d.ActiveCount())
		}
		close(in_events)
	}()

	//Consume events
	go func() {
		//defer close(job)
		for i := range out_events {
			log.Printf("Event: %s\n", i)
			switch i.Event {
			case dag.EventType_READY:
				go func(step_id string) {
					time.Sleep(time.Duration(rand.Int63n(1000)) * time.Microsecond)
					//log.Printf("Complete: %s", step_id)
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
}
