package main

import (
  "context"
  "fmt"
  "encoding/json"
  "io/ioutil"
  "os"
  "github.com/buchanae/ktl"
  "github.com/ghodss/yaml"
  "github.com/spf13/cobra"
)

var batchCmd = &cobra.Command{
  Use: "batch",
}

var createBatchCmd = &cobra.Command{
  Use: "create",
  Args: cobra.ExactArgs(1),
  RunE: func(cmd *cobra.Command, args []string) error {
    id, err := createBatch(args[0])
    if err != nil {
      return err
    }
    fmt.Println(id)
    return nil
  },
}

func createBatch(path string) (id string, err error) {
  cli := ktl.NewClient("http://"+ktl.DefaultServeOpts.Listen)
  ctx := context.Background()

  b, err := ioutil.ReadFile(path)
  if err != nil {
    return "", fmt.Errorf("reading batch file: %s", err)
  }

  batch := &ktl.Batch{}
  err = yaml.Unmarshal(b, batch)
  if err != nil {
    return "", fmt.Errorf("unmarshaling batch file: %s", err)
  }

  err = ktl.ValidateBatch(batch)
  if err != nil {
    return "", fmt.Errorf("validating batch: %s", err)
  }

  resp, err := cli.CreateBatch(ctx, batch)
  if err != nil {
    return "", err
  }
  return resp.ID, nil
}

var listBatchCmd = &cobra.Command{
  Use: "list",
  RunE: func(cmd *cobra.Command, args []string) error {
    cli := ktl.NewClient("http://"+ktl.DefaultServeOpts.Listen)
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
    cli := ktl.NewClient("http://"+ktl.DefaultServeOpts.Listen)
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
