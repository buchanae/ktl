package cwl

import (
	"fmt"
)

func (self CommandLineTool) CommandLineTool() (CommandLineTool, error) {
	return self, nil
}

func (self CommandLineTool) Workflow() (Workflow, error) {
	return Workflow{}, fmt.Errorf("Not Workflow")
}

func (self Workflow) CommandLineTool() (CommandLineTool, error) {
	return CommandLineTool{}, fmt.Errorf("Not CommandLineTools")
}

func (self Workflow) Workflow() (Workflow, error) {
	return self, nil
}

func (self Workflow) GetStep(n string) *WorkflowStep {
	for _, i := range self.Steps {
		if n == i.Id {
			return i
		}
	}
	return nil
}