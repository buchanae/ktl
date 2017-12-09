package cwl

import (
	"fmt"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/ohsu-comp-bio/ktl/pbutil"
	"log"
	"sort"
)

type Environment struct {
	DefaultImage string
	Inputs       pbutil.JSONDict
	Outputs      pbutil.JSONDict
	Runtime      pbutil.JSONDict
}

type OutputMapping struct {
	Id   string
	Glob []string
}

func getDockerImage(m map[string]interface{}) string {
	if x, ok := m["id"]; ok {
		if x == "DockerRequirement" {
			return m["dockerPull"].(string)
		}
	}
	if x, ok := m["class"]; ok {
		if x == "DockerRequirement" {
			return m["dockerPull"].(string)
		}
	}
	return ""
}

func (self CommandLineTool) GetImageName() string {
	out := ""
	for _, i := range self.Hints {
		m := pbutil.AsMap(i)
		s := getDockerImage(m)
		if s != "" {
			return s
		}
	}
	for _, i := range self.Requirements {
		m := pbutil.AsMap(i)
		s := getDockerImage(m)
		if s != "" {
			return s
		}
	}
	return out
}

func (self CommandLineTool) SetDefaults(env Environment) Environment {
	out := env
	for _, x := range self.Inputs {
		if _, ok := env.Inputs[x.Id]; !ok {
			if x.Default != nil {
				if y, ok := x.Default.GetKind().(*structpb.Value_StringValue); ok {
					out.Inputs[x.Id] = y.StringValue
				} else if y, ok := x.Default.GetKind().(*structpb.Value_StructValue); ok {
					out.Inputs[x.Id] = pbutil.AsMap(y.StructValue)
				}
			}
		}
	}
	return out
}

func (self CommandLineTool) GetOutputMapping(env Environment) ([]OutputMapping, error) {
	//Outputs
	out := []OutputMapping{}
	for _, x := range self.Outputs {
		o := OutputMapping{Id: x.Id}
		if x.OutputBinding != nil {
			eval := JSEvaluator{Inputs: env.Inputs, Outputs: env.Outputs, Runtime: env.Runtime}
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

func (self CommandLineTool) Render(mapper FileMapper, env Environment) ([]string, error) {

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
		for _, y := range new_args.ToArray() {
			args = append(args, JobArgument{*x, "", y})
		}
	}

	//Inputs
	for _, x := range self.Inputs {
		new_args, err := x.Evaluate(mapper, env)
		if err != nil {
			log.Printf("Input Error: %s", err)
			return []string{}, err
		}
		for _, y := range new_args.ToArray() {
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

func (self CommandLineTool) RenderStdinPath(mapper FileMapper, env Environment) (string, error) {
	if self.Stdin == "" {
		return self.Stdin, nil
	}
	eval := JSEvaluator{Inputs: env.Inputs, Outputs: env.Outputs, Runtime: env.Runtime}
	s, err := eval.EvaluateExpressionString(self.Stdin, nil)
	return mapper.Storage2Volume(s), err
}

type JobArgument struct {
	CommandLineBinding
	Id    string
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

func (self CommandLineBinding) Evaluate(env Environment) (StringTree, error) {
	//log.Printf("binding: %#v", self)
	out := NewStringTree()
	if len(self.Prefix) > 0 {
		out = out.Append(self.Prefix)
	}
	eval := JSEvaluator{Inputs: env.Inputs, Outputs: env.Outputs, Runtime: env.Runtime}
	result, err := eval.EvaluateExpressionString(self.ValueFrom, nil)
	if err != nil {
		return NewStringTree(), err
	}
	out = out.Append(result)
	return out, nil
}

func (self CommandInputParameter) Evaluate(mapper FileMapper, env Environment) (StringTree, error) {
	out := NewStringTree()
	if self.InputBinding != nil {
		if self.Type.GetName() == "boolean" {
			if env.Inputs[self.Id].(bool) {
				out = out.Append(self.InputBinding.Prefix)
			}
		} else {
			if len(self.InputBinding.Prefix) > 0 {
				out = out.Append(self.InputBinding.Prefix)
			}
			if len(self.InputBinding.ValueFrom) > 0 {
				eval := JSEvaluator{Inputs: env.Inputs, Outputs: env.Outputs, Runtime: env.Runtime}
				result, err := eval.EvaluateExpressionString(self.InputBinding.ValueFrom, nil)
				if err != nil {
					return NewStringTree(), err
				}
				out = out.Append(result)
			} else {
				v := self.Type.Evaluate(env.Inputs[self.Id], mapper)
				if len(self.InputBinding.ItemSeparator) > 0 {
					v = v.SetSeperator(self.InputBinding.ItemSeparator)
				}
				out = out.Extend(v)
			}
		}
	}
	return out, nil
}

func (self TypeRecord) Evaluate(v interface{}, mapper FileMapper) StringTree {
	switch r := self.GetType().(type) {
	case *TypeRecord_Name:
		if x, ok := v.(pbutil.JSONDict); ok {
			if y, ok := x["class"]; ok {
				if y == "File" {
					if z, ok := x["path"]; ok {
						return String2Tree(mapper.Storage2Volume(z.(string)))
					} else if z, ok := x["location"]; ok {
						return String2Tree(mapper.Storage2Volume(z.(string)))
					}
				}
			}
		}
		if isString(v) {
			return String2Tree(fmt.Sprintf("%s", v))
		}
		if isInt(v) {
			return String2Tree(fmt.Sprintf("%d", v))
		}
	case *TypeRecord_Array:
		out := NewStringTree()
		data := v.([]interface{})
		for i := range data {
			if self.InputBinding != nil {
				if self.InputBinding.Prefix != "" {
					out = out.Append(self.InputBinding.Prefix)
				}
			}
			s := r.Array.Items.Evaluate(data[i], mapper)
			out = out.Extend(s)
		}
		return out
	default:
		log.Printf("Missing TypeRecord Evaluate %T %T", self.GetType(), v)
	}
	return NewStringTree()
}
