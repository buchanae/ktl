package main

/*
Lessons:
- checking the state of multiple, sequential tasks in parallel would help.
*/

import (
  "context"
  "io"
  "fmt"
  "os"
  "github.com/golang/protobuf/jsonpb"
  "github.com/spf13/cobra"
  "google.golang.org/grpc"
)

// State variables for convenience
const (
	Unknown      = State_UNKNOWN
	Queued       = State_QUEUED
	Running      = State_RUNNING
	Paused       = State_PAUSED
	Complete     = State_COMPLETE
	Error        = State_ERROR
	SystemError  = State_SYSTEM_ERROR
	Canceled     = State_CANCELED
	Initializing = State_INITIALIZING
)

var runCmd = &cobra.Command{
  Use: "run [taskdir...]",
  RunE: func(cmd *cobra.Command, args []string) error {
    if len(args) == 0 {
      return cmd.Help()
    }

    // Loader helps load tasks in parallel.
    l, err := newLoader(serverHost)
    if err != nil {
      return err
    }

    // Load all tasks.
    for _, arg := range args {
      go l.loadDirectory(arg)
    }

    return nil
  },
}

func shouldStart(t Task, restart bool) bool {
  switch t.GetState() {
  case Queued, Initializing, Running:
    return false

  case Error, SystemError, Canceled:
    return restart

  case Unknown:
    return true
  }
  return false
}

func runSeq(args []string, cli TaskServiceClient) {

  for _, arg := range args {

    if id == "" {
      continue
    }
    fmt.Println("ID:", id)

    if shouldStart(t) {
      startTask(arg, cli)
    }
  }
}

func startTask(task Task, cli TaskServiceClient) {
  fmt.Println("Starting", arg)

  r, err := cli.CreateTask(context.Background(), task)
  if err != nil {
    panic(err)
  }

  fmt.Println("Created:", r.Id)
  saveID(arg, r.Id)
}


func saveID(path, id string) {
  f, err := os.Create(path + ".id")
  defer f.Close()

  if err != nil {
    panic(err)
  }
  f.WriteString(id)
}

func loadTask(r io.Reader) (*Task, error) {
  t := Task{}
  err := jsonpb.Unmarshal(r, &t)
  if err != nil {
    return nil, err
  }
  return &t, nil
}

