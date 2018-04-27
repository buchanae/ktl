package main

import (
  "context"
  "github.com/spf13/cobra"
  "github.com/buchanae/ktl"
)

var stepCmd = &cobra.Command{
  Use: "step",
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
}
