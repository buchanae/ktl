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

var restart bool

var runCmd = &cobra.Command{
  Use: "run [taskdir...]",
  RunE: func(cmd *cobra.Command, args []string) error {
    if len(args) == 0 {
      return cmd.Help()
    }

    conn, err := grpc.Dial(
      "funnel_server_1:9090",
      grpc.WithInsecure(),
      grpc.WithBlock(),
    )
    defer conn.Close()
    if err != nil {
      panic(err)
    }
    cli := NewTaskServiceClient(conn)

    for _, arg := range args {
      runSeq(globTasks(arg), cli)
    }
    return nil
  },
}

func init() {
  f := runCmd.Flags()
  f.BoolVar(&restart, "restart", restart, "Restart failed tasks")
}

func runSeq(args []string, cli TaskServiceClient) {

  run := runner{cli: cli}

  for _, arg := range args {
    id := loadID(arg)

    r, err := run.cli.GetTask(context.Background(), &GetTaskRequest{Id: id})
    if err != nil && !isNotFound(err) {
      panic(err)
    }

    switch r.GetState() {
    case State_QUEUED, State_INITIALIZING, State_RUNNING:
      fmt.Println("Already running", arg)
      return

    case State_ERROR, State_SYSTEM_ERROR, State_CANCELED:
      if restart {
        run.startTask(arg)
      }
      return

    case State_UNKNOWN:
      run.startTask(arg)
      return
    }
  }
}

type runner struct {
  cli TaskServiceClient
}

func (run *runner) startTask(arg string) {
  fmt.Println("Starting", arg)

  f, err := os.Open(arg)
  defer f.Close()

  if err != nil {
    panic(err)
  }

  task, err := loadTask(f)
  if err != nil {
    panic(err)
  }

  r, err := run.cli.CreateTask(context.Background(), task)
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

