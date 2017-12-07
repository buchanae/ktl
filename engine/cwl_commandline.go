package engine

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/client"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/ktl/cwl"
	"github.com/ohsu-comp-bio/ktl/pbutil"
	"log"
	"os"
	"path"
)

type Engine struct {
	client *client.Client
}

func NewEngine(host string) Engine {
	return Engine{client.NewClient(host)}
}

func (self Engine) RunCommandLine(cmd cwl.CommandLineTool, mapper cwl.FileMapper, env cwl.Environment) (pbutil.JSONDict, error) {
	log.Printf("Running CommandLineTool")
	tes_doc, post, err := Render(cmd, mapper, env)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Command line render failed %s\n", err))
		os.Exit(1)
	}

	resp, err := self.client.CreateTask(context.Background(), &tes_doc)
	if err != nil {
		log.Printf("Error: %s", err)
		return pbutil.JSONDict{}, err
	}

	self.client.WaitForTask(context.Background(), resp.Id)
	task_result, _ := self.client.GetTask(context.Background(), &tes.GetTaskRequest{Id: resp.Id, View: tes.TaskView_FULL})

	out := pbutil.JSONDict{}
	for _, i := range post.Steps {
		if x, ok := i.Step.(*PostProcessStep_GlobOutput); ok {
			for _, g := range x.GlobOutput.Glob {
				for _, j := range task_result.Logs[len(task_result.Logs)-1].Outputs {
					m, _ := path.Match(path.Join(cwl.DOCKER_WORK_DIR, g), j.Path)
					if m {
						out[x.GlobOutput.ParamName] = pbutil.JSONDict{
							"class": "File",
							"url":   j.Url,
						}
					}
				}
			}
		}
	}
	log.Printf("Response: %s", task_result)
	return out, nil
}
