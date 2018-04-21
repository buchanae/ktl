package main

import (
  "context"
  "fmt"
  "github.com/ohsu-comp-bio/ktl"
  "github.com/ohsu-comp-bio/ktl/driver/task"
  "github.com/ohsu-comp-bio/ktl/database/mongodb"
  "github.com/spf13/cobra"
)

func init() {
  cmd := &cobra.Command{
    Use: "serve",
    Args: cobra.NoArgs,
    RunE: func(cmd *cobra.Command, args []string) error {

      ctx := context.Background()
      db, err := mongodb.NewMongoDB(mongodb.DefaultConfig())
      if err != nil {
        return err
      }

      taskDriver, err := task.NewDriver()
      if err != nil {
        return fmt.Errorf("creating task driver: %s", err)
      }

      drivers := map[string]ktl.Driver{
        "Task": taskDriver,
      }

      proc := ktl.NewProcessor(db, drivers)
      go proc.Process(ctx)

      return ktl.Serve(db)
    },
  }
  root.AddCommand(cmd)
}
