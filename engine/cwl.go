package engine

import (
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/cmd/client"
	"github.com/ohsu-comp-bio/ktl/cwl"
	"github.com/ohsu-comp-bio/ktl/tes"
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

	tes_doc, err := tes.Render(cmd, mapper, env)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Command line render failed %s\n", err))
		os.Exit(1)
	}

	log.Printf("Submitting %s", tes_doc)
	m := jsonpb.Marshaler{}
	tmes, _ := m.MarshalToString(&tes_doc)

	resp, err := self.client.CreateTask([]byte(tmes))
	if err != nil {
		log.Printf("Error: %s", err)
		return cwl.JSONDict{}, err
	}

	self.client.WaitForTask(resp.Id)
	task_result, _ := self.client.GetTask(resp.Id, "FULL")

	log.Printf("Response: %s", task_result)
	return cwl.JSONDict{}, nil
}
