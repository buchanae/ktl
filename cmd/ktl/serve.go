package main

import (
  "context"
  "log"
  "fmt"
  "time"
  "github.com/buchanae/ktl"
  "github.com/buchanae/ktl/driver/task"
  "github.com/buchanae/ktl/database/mongodb"
  "github.com/spf13/cobra"
)

func init() {
  opts := ktl.DefaultServeOpts

  cmd := &cobra.Command{
    Use: "serve",
    Args: cobra.NoArgs,
    RunE: func(cmd *cobra.Command, args []string) error {

      ctx := context.Background()
      db, err := mongodb.NewMongoDB(mongodb.DefaultConfig())
      if err != nil {
        return err
      }

      taskDriver, err := task.NewDriver(opts.TaskAPI)
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

      return ktl.Serve(db, opts)
    },
  }

  f := cmd.Flags()
  f.StringVar(&opts.Listen, "listen", opts.Listen, "Address for server to listen on.")
  f.StringVar(&opts.TaskAPI, "taskapi", opts.TaskAPI, "Address of the Task API (i.e. Funnel).")
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
