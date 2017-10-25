package cwl


import (
  "log"
  "fmt"
  "sort"
)

type Environment struct {
  Inputs JSONDict
  Outputs JSONDict
  Runtime JSONDict
}

type OutputMapping struct {
  Id string
  Glob []string
}

func (self CommandLineTool) SetDefaults(env Environment) Environment {
  out := env
  for _, x := range self.Inputs {
    if _, ok := env.Inputs[x.Id]; !ok {
      if x.Default != nil {
        out.Inputs[x.Id] = x.Default.GetStringValue() //BUG: This could be a none string value
      }
    }
  }
  return out
}

func (self CommandLineTool) GetOutputMapping(env Environment) ([]OutputMapping, error) {
  	//Outputs
    out := []OutputMapping{}
  	for _, x := range self.Outputs {
      o := OutputMapping{Id:x.Id}
      if x.OutputBinding != nil {
        eval := JSEvaluator{Inputs:env.Inputs,Outputs:env.Outputs,Runtime:env.Runtime}
        result, err := eval.EvaluateExpressionString(x.OutputBinding.Glob[0], nil)
        if err != nil {
          return []OutputMapping{}, err
        }
        o.Glob = []string{result}
      }
      out = append(out, o)
  	}
    return out, nil
}

func (self CommandLineTool) Render(env Environment) ([]string, error) {

  log.Printf("CommandLineTool Evalute")

	args := make(jobArgArray, 0, len(self.Arguments)+len(self.Inputs))

	for _, x := range self.BaseCommand {
		args = append(args, JobArgument{CommandLineBinding{Position: -10000}, "", x})
	}

	//Arguments
	for _, x := range self.Arguments {
		new_args, err := x.Evaluate(env)
		if err != nil {
			log.Printf("Argument Error: %s", err)
			return []string{}, err
		}
    for _, y := range new_args {
		    args = append(args, JobArgument{*x, "", y})
    }
	}

	//Inputs
	for _, x := range self.Inputs {
		new_args, err := x.Evaluate(env)
		if err != nil {
			log.Printf("Input Error: %s", err)
			return []string{}, err
		}
    for _, y := range new_args {
		    args = append(args, JobArgument{*x.InputBinding, "", y})
    }
	}

	sort.Stable(args)
  out := make([]string, len(args))
  for i := range args {
    out[i] = args[i].Value
  }
	//log.Printf("Out: %v", args)
	return out, nil
}

type JobArgument struct {
  CommandLineBinding
  Id string
  Value string
}

type jobArgArray []JobArgument

func (self jobArgArray) Len() int {
	return len(self)
}

func (self jobArgArray) Less(i, j int) bool {
	if (self)[i].Position == (self)[j].Position {
		return (self)[i].Id < (self)[j].Id
	}
	return (self)[i].Position < (self)[j].Position
}

func (self jobArgArray) Swap(i, j int) {
	(self)[i], (self)[j] = (self)[j], (self)[i]
}



func (self CommandLineBinding) Evaluate(env Environment) ([]string, error) {
  //log.Printf("binding: %#v", self)
  out := []string{}
  if len(self.Prefix) > 0 {
    out = append(out, self.Prefix)
  }
  eval := JSEvaluator{Inputs:env.Inputs,Outputs:env.Outputs,Runtime:env.Runtime}
  result, err := eval.EvaluateExpressionString(self.ValueFrom, nil)
  if err != nil {
    return []string{}, err
  }
  out = append(out, result)
  return out, nil
}

func (self CommandInputParameter) Evaluate(env Environment) ([]string, error) {
  out := []string{}
  if self.InputBinding != nil {
    if len(self.InputBinding.Prefix) > 0 {
      out = append(out, self.InputBinding.Prefix)
    }
    if len(self.InputBinding.ValueFrom) > 0 {
      eval := JSEvaluator{Inputs:env.Inputs,Outputs:env.Outputs,Runtime:env.Runtime}
      result, err := eval.EvaluateExpressionString(self.InputBinding.ValueFrom, nil)
      if err != nil {
        return []string{}, err
      }
      out = append(out, result)
    } else {
      for _, s := range self.Type.Evaluate(env.Inputs[self.Id]) {
          out = append(out, s)
      }
    }
  }
  return out, nil
}

func (self TypeRecord) Evaluate(v interface{}) []string {
  if x, ok := v.(JSONDict); ok {
    if y, ok := x["class"]; ok {
        if y == "File" {
          return []string{fmt.Sprintf("%s", x["path"])}
        }
      }
    }
  return []string{fmt.Sprintf("%s", v)}
}
