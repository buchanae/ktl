

package tes

import (
  "github.com/ohsu-comp-bio/ktl/cwl"
)

func Render(cmd cwl.CommandLineTool, env cwl.Environment) (Task, error) {
  cmd_line, err := cmd.Render(env)
  if err != nil {
    return Task{}, err
  }
  
  out := Task{}
  exec := Executor{}
  exec.Cmd = cmd_line
  exec.ImageName = cmd.GetImageName()
  out.Executors = []*Executor{&exec}
  
  for _, i := range cmd.GetMappedInputs(env) {
    input := TaskParameter{
      Url: i.StoragePath,
      Path: i.MappedPath,
    }
    out.Inputs = append(out.Inputs, &input)
  }
  
  return out, nil
}