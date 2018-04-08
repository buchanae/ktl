package main

import (
  "os"
  "github.com/spf13/cobra"
)

var root = cobra.Command{
  Use: "ktl",
	SilenceUsage:  true,
}

func main() {
  if err := root.Execute(); err != nil {
    os.Exit(1)
  }
}
