package main

import (
  "log"
  "fmt"
  "os"
  "github.com/golang/protobuf/jsonpb"
  "github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(editCmd)
}

var editCmd = &cobra.Command{
    Use:   "edit [task.json ...]",
    RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) == 0 {
            return cmd.Help()
        }
        for _, arg := range args {
          doEdit(globTasks(arg))
        }
        return nil
    },
}


var editFlags = struct {
  cpus int
  write bool
  name string
}{}


func init() {
  f := editCmd.Flags()
  f.IntVar(&editFlags.cpus, "cpus", -1, "Set CPUs")
  f.StringVar(&editFlags.name, "name", editFlags.name, "Set name")
  f.BoolVar(&editFlags.write, "write", editFlags.write, "Write in place")
}

func doEdit(args []string) {

  mar := jsonpb.Marshaler{
    Indent: "  ",
  }

  for _, arg := range args {
    f, err := os.Open(arg)
    if err != nil {
      log.Print(err)
      continue
    }

    task, err := loadTask(f)
    if err != nil {
      log.Print(err)
      continue
    }
    f.Close()

    if editFlags.cpus > -1 {
      task.Resources.CpuCores = uint32(editFlags.cpus)
    }
    if editFlags.name != "" {
      task.Name = editFlags.name
    }

    s, err := mar.MarshalToString(task)
    if err != nil {
      log.Print(err)
      continue
    }

    fmt.Println(s)
    if editFlags.write {
      f, err := os.OpenFile(arg, os.O_RDWR, 0664)
      if err != nil {
        log.Print(err)
        continue
      }
      err = mar.Marshal(f, task)
      if err != nil {
        log.Print(err)
        continue
      }
      f.Close()
    }
  }
}
