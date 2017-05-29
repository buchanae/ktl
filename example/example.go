package main

import (
  "fmt"
  "github.com/ohsu-comp-bio/whisk"
  "bytes"
  "text/template"
  "encoding/json"
)

// This is an overly simple example of a workflow definition in JSON,
// used to demonstrate the structure of a whisk workflow driver.
//
// In reality, this would be a CWL/WDL/etc. document and the driver
// code would be more complex.
var exampleRaw = `
{ "steps": [
  {"id": "task1", "cmd": "TASK 1"},
  {"id": "task2", "deps": ["task1"], "cmd": "TASK 2 {{ .task1.Stdout }}"},
  {"id": "task3", "cmd": "TASK 3"},
  {"id": "task4", "deps": ["task2"], "cmd": "TASK 4"},
  {"id": "task5", "deps": ["task2", "task3"], "cmd": "TASK 5"}
]}
`

func main() {
  // The ID passed to NewWorkflow is used as the basis for call caching.
  wf := whisk.NewWorkflow("example-1")
  RunExample(wf, exampleRaw)
}

// ExampleDoc represents an example workflow document.
// w.r.t. CWL, this would be a parsed CWL document.
type ExampleDoc struct {
  Steps []ExampleStep
}

// ExampleStep represents a step in the example workflow document.
type ExampleStep struct {
  ID string
  // CLI command string to run. This is executed as a Go template,
  // to resolve variables, e.g. 'echo {{ dep1.stdout }}'
  Cmd string
  // List of dependency step IDs.
  Deps []string
}

// Given a example workflow string, parse it and create workflow tasks.
func RunExample(wf whisk.Workflow, raw string) error {

  // For simplicity, every error in this example is a panic.
  // This handles those panics and fails the workflow.
  defer wf.Recover()

  doc := MustParse(raw)
  fmt.Println("RunEx", len(doc.Steps))

  for _, step := range doc.Steps {
    RunExampleStep(wf.NS(step.ID), step)
  }

  return wf.Wait()
}

// Create/manage tasks for an ExampleStep instance
func RunExampleStep(wf whisk.Workflow, step ExampleStep) {
  wf.MustGo(func() {
    fmt.Println("RunStep", step.Deps)

    // Get data for command template
    data := map[string]whisk.Task{}
    for _, t := range wf.MustGetTasks(step.Deps) {
      data[t.ID] = t
    }

    // Create task with rendered command template
    wf.CreateTask(whisk.Task{
      ID: step.ID,
      Cmd: MustGetCmd(step.Cmd, data),
    })
  })
}

// Render a command template
func MustGetCmd(raw string, data map[string]whisk.Task) string {
  var b bytes.Buffer
  tpl := template.Must(template.New("cmd").Parse(raw))
  err := tpl.Execute(&b, data)
  if err != nil {
    panic(err)
  }
  return b.String()
}

// Parse a JSON string into ExampleDoc
func MustParse(raw string) ExampleDoc {
  doc := ExampleDoc{}
  err := json.Unmarshal([]byte(raw), &doc)
  if err != nil {
    panic(err)
  }
  return doc
}
