package engine

import (
	"fmt"
	"context"
	"github.com/ohsu-comp-bio/funnel/client"
	"github.com/ohsu-comp-bio/ktl/cwl"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"log"
	"os"
)

type Engine struct {
	client *client.Client
}

func NewEngine(host string) Engine {
	return Engine{client.NewClient(host)}
}

func (self Engine) Run(cmd cwl.CommandLineTool, mapper cwl.FileMapper, env cwl.Environment) (cwl.JSONDict, error) {

	tes_doc, err := Render(cmd, mapper, env)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Command line render failed %s\n", err))
		os.Exit(1)
	}

	resp, err := self.client.CreateTask(context.Background(), &tes_doc)
	if err != nil {
		log.Printf("Error: %s", err)
		return cwl.JSONDict{}, err
	}

	self.client.WaitForTask(context.Background(), resp.Id)
	task_result, _ := self.client.GetTask(context.Background(), &tes.GetTaskRequest{Id:resp.Id, View:tes.TaskView_FULL})

	log.Printf("Response: %s", task_result)
	return cwl.JSONDict{}, nil
}
