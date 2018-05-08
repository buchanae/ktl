package main

import (
  "context"
  "fmt"
  "io/ioutil"

  "github.com/ghodss/yaml"
  "github.com/spf13/cobra"
  "github.com/buchanae/ktl"
)

var stepCmd = &cobra.Command{
  Use: "step",
}

var putStepCmd = &cobra.Command{
  Use: "put <batch> <step>",
  Args: cobra.ExactArgs(2),
  RunE: func(cmd *cobra.Command, args []string) error {
    cli := ktl.NewClient("http://"+ktl.DefaultServeOpts.Listen)
    ctx := context.Background()
    b, err := ioutil.ReadFile(args[1])
    if err != nil {
      return fmt.Errorf("reading step file: %s")
    }

    step := &ktl.Step{}
    err = yaml.Unmarshal(b, step)
    if err != nil {
      return fmt.Errorf("unmarshaling step file: %s", err)
    }

    return cli.PutStep(ctx, args[0], step)
  },
}

var restartStepCmd = &cobra.Command{
  Use: "restart <batch> <step>",
  Args: cobra.ExactArgs(2),
  RunE: func(cmd *cobra.Command, args []string) error {
    cli := ktl.NewClient("http://"+ktl.DefaultServeOpts.Listen)
    ctx := context.Background()
    return cli.RestartStep(ctx, args[0], args[1])
  },
}

func init() {
  root.AddCommand(stepCmd)
  stepCmd.AddCommand(restartStepCmd)
  stepCmd.AddCommand(putStepCmd)
}
