package main

/*
Lessons:
- checking the state of multiple, sequential tasks in parallel would help.
*/

import (
  "context"
  "io"
  "time"
  "fmt"
  "os"
  "github.com/golang/protobuf/jsonpb"
  "github.com/spf13/cobra"
)

var restart bool

var runCmd = &cobra.Command{
  Use: "run [taskdir...]",
  RunE: func(cmd *cobra.Command, args []string) error {
    if len(args) == 0 {
      return cmd.Help()
    }
    runSeq(globTasks(args))
    return nil
  },
}

func init() {
  f := runCmd.Flags()
  f.BoolVar(&restart, "restart", restart, "Restart failed tasks")
}

func runSeq(args []string) {
  cli, err := newTaskClient("funnel_server_1:9090")
  if err != nil {
    panic(err)
  }

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


func waitForTask(ctx context.Context, client TaskServiceClient, id string) error {
  for {
    r, err := client.GetTask(ctx, &GetTaskRequest{Id: id})
    if err != nil {
      return err
    }

    switch r.State {
    case State_ERROR, State_SYSTEM_ERROR:
      return fmt.Errorf("Task error")
    case State_CANCELED:
      return fmt.Errorf("Task canceled")
    case State_COMPLETE:
      return nil
    default:
      fmt.Println("State:", r.State.String())
    }
    time.Sleep(time.Second)
  }
}

func loadTask(r io.Reader) (*Task, error) {
  t := Task{}
  err := jsonpb.Unmarshal(r, &t)
  if err != nil {
    return nil, err
  }
  return &t, nil
}

