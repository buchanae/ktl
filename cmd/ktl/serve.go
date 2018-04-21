package main

import (
  "context"
  "log"
  "fmt"
  "time"
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
      go func() {
        // TODO configurable
        for range ticker(ctx, 2 * time.Second) {
          err := proc.Process(ctx)
          if err != nil {
            log.Println("error: ", err)
          }
        }
      }()

      return ktl.Serve(db)
    },
  }
  root.AddCommand(cmd)
}

func ticker(ctx context.Context, d time.Duration) <-chan time.Time {
  out := make(chan time.Time)
  go func() {
    out <- time.Now()
    ticker := time.NewTicker(d)
    defer ticker.Stop()
    for {
      select {
      case <-ctx.Done():
        return
      case t := <-ticker.C:
        out <- t
      }
    }
  }()
  return out
}
