package main

import (
  "context"
  "fmt"
  "encoding/json"
  "time"
  "os"
  "github.com/ohsu-comp-bio/ktl"
  "github.com/spf13/cobra"
)

var batchCmd = &cobra.Command{
  Use: "batch",
}

var createBatchCmd = &cobra.Command{
  Use: "create",
  RunE: func(cmd *cobra.Command, args []string) error {
    cli := ktl.NewClient("http://"+ktl.DefaultListen)
    ctx := context.Background()

    b := &ktl.Batch{
      Steps: []*ktl.Step{
        {
          ID: "test-1",
          Type: "Task",
        },
        {
          ID: "test-2",
          Type: "Task",
          Dependencies: []string{"test-1"},
          Timeout: 5 * time.Second,
        },
      },
    }

    resp, err := cli.CreateBatch(ctx, b)
    if err != nil {
      return err
    }

    fmt.Println(resp.ID)
    return nil
  },
}

var listBatchCmd = &cobra.Command{
  Use: "list",
  RunE: func(cmd *cobra.Command, args []string) error {
    cli := ktl.NewClient("http://"+ktl.DefaultListen)
    ctx := context.Background()

    resp, err := cli.ListBatches(ctx)
    if err != nil {
      return err
    }

    enc := json.NewEncoder(os.Stdout)
    return enc.Encode(resp)
  },
}

var getBatchCmd = &cobra.Command{
  Use: "get",
  Args: cobra.ExactArgs(1),
  RunE: func(cmd *cobra.Command, args []string) error {
    cli := ktl.NewClient("http://"+ktl.DefaultListen)
    ctx := context.Background()

    resp, err := cli.GetBatch(ctx, args[0])
    if err != nil {
      return err
    }

    enc := json.NewEncoder(os.Stdout)
    return enc.Encode(resp)
  },
}

func init() {
  root.AddCommand(batchCmd)
  batchCmd.AddCommand(createBatchCmd)
  batchCmd.AddCommand(listBatchCmd)
  batchCmd.AddCommand(getBatchCmd)
}
