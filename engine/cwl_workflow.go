package engine

import (
	"log"
	"fmt"
	//"time"
	"github.com/ohsu-comp-bio/ktl/cwl"
	"github.com/ohsu-comp-bio/ktl/dag"
	"github.com/ohsu-comp-bio/ktl/pbutil"
	"strings"
)


func (self Engine) processEvents(in_events, out_events chan dag.Event, wf cwl.Workflow, graph cwl.CWLGraph, mapper cwl.FileMapper, env cwl.Environment) {
	for e := range out_events {
		log.Printf("Out: %s", e)
		if strings.HasPrefix(e.StepId, "/") {
			param_name := e.StepId[1:len(e.StepId)]
			output := pbutil.JSONDict{
				param_name : env.Inputs[param_name],
			}
			in_events <- dag.Event{
				StepId: e.StepId,
				Event:  dag.EventType_SUCCESS,
				Params: output.AsStruct(),
			}
		} else {
			step := wf.GetStep(e.StepId)
			var cmd *cwl.CommandLineTool = nil
			if x, ok := step.Run.Run.(*cwl.RunRecord_Commandline); ok {
				cmd = x.Commandline
			}
			if x, ok := step.Run.Run.(*cwl.RunRecord_Path); ok {
				c, err := graph.Elements[x.Path].CommandLineTool()
				if err == nil {
					cmd = &c
				}
			}
			cmd_env := cmd.SetDefaults(cwl.Environment{Inputs: pbutil.AsMap(e.Params), DefaultImage:env.DefaultImage})
			out, err := self.RunCommandLine(*cmd, mapper, cmd_env)
			if err == nil {
				in_events <- dag.Event{
					StepId: e.StepId,
					Event:  dag.EventType_SUCCESS,
					Params: out.AsStruct(),
				}
			} else {
				in_events <- dag.Event{
					StepId: e.StepId,
					Event:  dag.EventType_FAILURE,
				}
			}
		}
	}
}

func (self Engine) RunWorkflow(wf cwl.Workflow, graph cwl.CWLGraph, mapper cwl.FileMapper, env cwl.Environment) (pbutil.JSONDict, error) {
	log.Printf("Starting Workflow")

	NWORKERS := 4
	md := dag.MemoryDAG{}

	in_events := make(chan dag.Event, 100)
	out_events := md.Start(in_events)
	quit := make(chan bool, NWORKERS)
	for i := 0; i < NWORKERS; i++ {
		go func() {
			self.processEvents(in_events, out_events, wf, graph, mapper, env)
			quit <- true
		}()
	}


	steps := map[string]bool{}

	//Add inputs into event dag
	for _, i := range wf.Inputs {
		//Input files are recorded as events with no inputs that then produce
		//in input file as their output
		event_name := fmt.Sprintf("/%s", i.Id)
		de := dag.Event{
			StepId:  event_name,
			Event:   dag.EventType_NEW,
		}
		in_events <- de
		steps[event_name] = true
	}

	//Add steps into event dag
	for added := true; added; {
		added = false
		for _, s := range wf.Steps {
			if _, ok := steps[s.Id]; !ok {
				deps := map[string]bool{}
				ins := []*dag.InputMapping{}
				for _, i := range s.In {
					if len(i.Source) > 0 {
						p := strings.Split(i.Source[0], "/")
						if len(p) == 2 {
							d := dag.InputMapping{
								SrcStepId:p[0],
								SrcParamName:p[1],
								ParamName:i.Id,
							}
							ins = append(ins, &d)
							deps[p[0]] = true
						} else if len(p) == 1 {
							event_name := fmt.Sprintf("/%s", p[0])
							d := dag.InputMapping{
								SrcStepId:event_name,
								SrcParamName:p[0],
								ParamName:i.Id,
							}
							ins = append(ins, &d)
							deps[event_name] = true
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
						StepId:  s.Id,
						Event:   dag.EventType_NEW,
						Depends: da,
						Inputs: ins,
					}
					log.Printf("Add %s", de)
					in_events <- de
					steps[s.Id] = true
					added = true
				}
			}
		}
	}
	de := dag.Event{Event: dag.EventType_CLOSE}
	in_events <- de
	for i := 0; i < NWORKERS; i++ {
		<-quit
	}
	return pbutil.JSONDict{}, nil
}
