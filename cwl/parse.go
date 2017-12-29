package cwl

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/ktl/pbutil"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

func mapNormalize(v interface{}) interface{} {
	if base, ok := v.(map[interface{}]interface{}); ok {
		out := pbutil.JSONDict{}
		for k, v := range base {
			out[k.(string)] = mapNormalize(v)
		}
		return out
	} else if base, ok := v.(map[string]interface{}); ok {
		out := map[string]interface{}{}
		for k, v := range base {
			out[k] = mapNormalize(v)
		}
		return out
	} else if base, ok := v.([]interface{}); ok {
		out := make([]interface{}, len(base))
		for i, v := range base {
			out[i] = mapNormalize(v)
		}
		return out
	}
	return v
}

func YamlLoad(path string) (pbutil.JSONDict, error) {
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	doc := make(map[interface{}]interface{})
	err = yaml.Unmarshal(source, &doc)
	out := mapNormalize(doc)
	return out.(pbutil.JSONDict), nil
}

func InputParse(path string, mapper FileMapper) (pbutil.JSONDict, error) {
	doc, err := YamlLoad(path)

	x, _ := filepath.Abs(path)
	base_path := filepath.Dir(x)

	out := SetInputAbsPath(doc, base_path).(pbutil.JSONDict)
	return out, err
}

func Parse(cwl_path string) (CWLGraph, error) {
	doc, err := YamlLoad(cwl_path)
	if err != nil {
		return CWLGraph{}, fmt.Errorf("Unable to parse file: %s", err)
	}
	if base, ok := doc["$graph"]; ok {
		log.Printf("%s\n", base)
		//return parser.NewGraph(base)
	} else if class, ok := doc["class"]; ok {
		if class == "CommandLineTool" {
			fixed_doc := FixCommandLineTool(doc)
			jdoc, err := json.MarshalIndent(fixed_doc, "", "   ")
			if err != nil {
				log.Printf("%s", err)
			}
			log.Printf("Parsed: %s\n", jdoc)
			cmd := CommandLineTool{}

			umarsh := jsonpb.Unmarshaler{AllowUnknownFields: true}
			err = umarsh.Unmarshal(strings.NewReader(string(jdoc)), &cmd)
			if err != nil {
				log.Printf("SchemaParseError: %s", err)
				return CWLGraph{}, fmt.Errorf("Unable to parse file %s", cwl_path)
			}
			return CWLGraph{Main: "#", Elements: map[string]CWLDoc{"#": cmd}}, nil
		} else if class == "Workflow" {
			fixed_doc := FixWorkflow(doc)
			jdoc, err := json.MarshalIndent(fixed_doc, "", "   ")
			if err != nil {
				log.Printf("%s", err)
			}
			log.Printf("Parsed: %s\n", jdoc)
			wf := Workflow{}
			umarsh := jsonpb.Unmarshaler{AllowUnknownFields: true}
			err = umarsh.Unmarshal(strings.NewReader(string(jdoc)), &wf)
			if err != nil {
				log.Printf("SchemaParseError: %s", err)
				return CWLGraph{}, fmt.Errorf("Unable to parse file %s", cwl_path)
			}
			out := CWLGraph{Main: "#", Elements: map[string]CWLDoc{"#": wf}}
			for _, s := range wf.Steps {
				if s.Run.GetPath() != "" {
					g, err := Parse(s.Run.GetPath())
					if err != nil {
						return CWLGraph{}, err
					}
					out.Elements[s.Run.GetPath()] = g.Elements["#"]
				}
			}
			return out, nil
		} else {
			return CWLGraph{}, fmt.Errorf("Unknown class %s", class)
		}
		//return parser.NewClass(doc)
	}
	return CWLGraph{}, fmt.Errorf("Unable to parse file")
}

func isDict(i interface{}) bool {
	if _, ok := i.(map[string]interface{}); ok {
		return true
	}
	if _, ok := i.(pbutil.JSONDict); ok {
		return true
	}
	return false
}

func isString(i interface{}) bool {
	_, ok := i.(string)
	return ok
}

func isFloat(i interface{}) bool {
	_, ok := i.(float32)
	return ok
}

func isInt(i interface{}) bool {
	_, ok := i.(int)
	return ok
}

func isList(i interface{}) bool {
	_, ok := i.([]interface{})
	return ok
}

func contains(a []string, v string) bool {
	for _, i := range a {
		if i == v {
			return true
		}
	}
	return false
}

func fixDict2List(doc pbutil.JSONDict, typeField string, fields ...string) pbutil.JSONDict {
	out := pbutil.JSONDict{}
	for k, v := range doc {
		if contains(fields, k) {
			if isDict(v) {
				nv := []interface{}{}
				for ek, ev := range v.(pbutil.JSONDict) {
					if isString(ev) || isList(ev) {
						i := pbutil.JSONDict{"id": ek, typeField: ev}
						nv = append(nv, i)
					} else {
						i := ev.(pbutil.JSONDict)
						i["id"] = ek
						nv = append(nv, i)
					}
				}
				out[k] = nv
			} else {
				log.Printf("Skipped: %s %T\n", k, v)
				out[k] = v
			}
		} else {
			out[k] = v
		}
	}
	return out
}

func fixForceList(doc pbutil.JSONDict, fields ...string) pbutil.JSONDict {
	out := pbutil.JSONDict{}
	for k, v := range doc {
		if contains(fields, k) {
			if !isList(v) {
				out[k] = []interface{}{v}
			} else {
				out[k] = v
			}
		} else {
			out[k] = v
		}
	}
	return out
}

func FixDataRecord(doc interface{}) pbutil.JSONDict {
	if isString(doc) {
		return pbutil.JSONDict{"string_value": doc.(string)}
	}
	if isDict(doc) {
		return pbutil.JSONDict{"struct_value": doc.(pbutil.JSONDict)}
	}
	if isList(doc) {
		return pbutil.JSONDict{"list_value": doc.([]interface{})}
	}
	if isFloat(doc) {
		return pbutil.JSONDict{"float_value": doc.(float32)}
	}
	if isInt(doc) {
		return pbutil.JSONDict{"int_value": doc}
	}
	log.Printf("Unknown Type: %#T", doc)
	return pbutil.JSONDict{}
}

func FixTypeRecord(doc interface{}) pbutil.JSONDict {
	if isString(doc) {
		doc_string := doc.(string)
		if strings.HasSuffix(doc_string, "[]") {
			return pbutil.JSONDict{"array": pbutil.JSONDict{"items": FixTypeRecord(doc_string[:len(doc_string)-2])}}
		} else {
			return pbutil.JSONDict{"name": doc}
		}
	}
	if isDict(doc) {
		doc_dict := doc.(pbutil.JSONDict)
		if doc_dict["type"] == "array" {
			o := pbutil.JSONDict{"array": pbutil.JSONDict{"items": FixTypeRecord(doc_dict["items"])}}
			if x, ok := doc_dict["inputBinding"]; ok {
				o["inputBinding"] = x
			}
			return o
		}
		if doc_dict["type"] == "record" {
			doc_fields_list := doc_dict["fields"].([]interface{})
			fields := []interface{}{}
			for _, i := range doc_fields_list {
				field_doc := i.(pbutil.JSONDict)
				fields = append(fields, pbutil.JSONDict{"name": field_doc["name"], "type": FixTypeRecord(field_doc["type"])})
			}
			return pbutil.JSONDict{"record": pbutil.JSONDict{"name": doc_dict["name"], "fields": fields}}
			if doc_dict["type"] == "enum" {
				return pbutil.JSONDict{"enum": pbutil.JSONDict{"name": doc_dict["name"], "symbols": doc_dict["symbols"]}}
			}
		}
	}
	if isList(doc) {
		doc_list := doc.([]interface{})
		t := []interface{}{}
		for _, i := range doc_list {
			t = append(t, FixTypeRecord(i))
		}
		return pbutil.JSONDict{"oneof": pbutil.JSONDict{"types": t}}
	}
	return pbutil.JSONDict{}
}

func FixInputRecordField(doc pbutil.JSONDict) pbutil.JSONDict {
	doc = fixForceList(doc, "doc", "secondaryFiles")
	doc["type"] = FixTypeRecord(doc["type"])
	if x, ok := doc["default"]; ok {
		doc["default"] = FixDataRecord(x)
	}
	return doc
}

func FixInputRecordFieldList(list []interface{}) interface{} {
	out := make([]interface{}, len(list))
	for i := range list {
		i_doc := list[i].(pbutil.JSONDict)
		out[i] = FixInputRecordField(i_doc)
	}
	return out
}

func FixCommandOutputBinding(doc pbutil.JSONDict) pbutil.JSONDict {
	doc = fixForceList(doc, "glob")
	return doc
}

func FixOutputRecordField(doc pbutil.JSONDict) pbutil.JSONDict {
	doc = fixForceList(doc, "doc", "outputSource")
	doc["type"] = FixTypeRecord(doc["type"])
	if x, ok := doc["outputBinding"]; ok {
		doc["outputBinding"] = FixCommandOutputBinding(x.(pbutil.JSONDict))
	}
	return doc
}

func FixOutputRecordFieldList(list []interface{}) interface{} {
	out := make([]interface{}, len(list))
	for i := range list {
		i_doc := list[i].(pbutil.JSONDict)
		out[i] = FixOutputRecordField(i_doc)
	}
	return out
}

func FixCommandLineBinding(doc interface{}) pbutil.JSONDict {
	if isString(doc) {
		return pbutil.JSONDict{"valueFrom": doc}
	}
	return doc.(pbutil.JSONDict)
}

func FixCommandLineBindingList(list []interface{}) interface{} {
	out := []interface{}{}
	for _, i := range list {
		out = append(out, FixCommandLineBinding(i))
	}
	return out
}

func FixCommandLineTool(doc pbutil.JSONDict) pbutil.JSONDict {
	doc = fixDict2List(doc, "type", "inputs", "outputs", "hints", "requirements")
	doc = fixForceList(doc, "baseCommand", "doc")
	if x, ok := doc["inputs"]; ok {
		doc["inputs"] = FixInputRecordFieldList(x.([]interface{}))
	}
	if x, ok := doc["outputs"]; ok {
		doc["outputs"] = FixOutputRecordFieldList(x.([]interface{}))
	}
	if x, ok := doc["arguments"]; ok {
		doc["arguments"] = FixCommandLineBindingList(x.([]interface{}))
	}
	//undo stdout capture shortcuts

	return doc
}

func FixWorkflowStepOutput(i interface{}) interface{} {
	if x, ok := i.(string); ok {
		return pbutil.JSONDict{"id": x}
	}
	return i
}

func FixWorkflowStepOutputList(x []interface{}) []interface{} {
	out := []interface{}{}
	for _, i := range x {
		out = append(out, FixWorkflowStepOutput(i))
	}
	return out
}

func FixWorkflowStepInput(x interface{}) pbutil.JSONDict {
	if i, ok := x.(string); ok {
		return pbutil.JSONDict{"source": []interface{}{i}}
	}
	if i, ok := x.(pbutil.JSONDict); ok {
		return i
	}
	return pbutil.JSONDict{}
}

func FixWorkflowStepInputList(x []interface{}) []interface{} {
	out := []interface{}{}
	for _, i := range x {
		out = append(out, FixWorkflowStepInput(i))
	}
	return out
}

func FixWorkflowStepInputMap(x map[string]interface{}) []interface{} {
	out := []interface{}{}
	for k, v := range x {
		i := FixWorkflowStepInput(v)
		i["id"] = k
		out = append(out, i)
	}
	return out
}

func FixWorkflowStep(doc pbutil.JSONDict) pbutil.JSONDict {
	if x, ok := doc["in"]; ok {
		if x_list, ok := x.([]interface{}); ok {
			doc["in"] = FixWorkflowStepInputList(x_list)
		}
		if x_map, ok := x.(pbutil.JSONDict); ok {
			doc["in"] = FixWorkflowStepInputMap(x_map)
		}
	}
	if x, ok := doc["out"]; ok {
		doc["out"] = FixWorkflowStepOutputList(x.([]interface{}))
	}
	if x, ok := doc["run"]; ok {
		if x_str, ok := x.(string); ok {
			doc["run"] = pbutil.JSONDict{"path": x_str}
		}
	}
	return doc
}

func FixWorkflowStepList(list []interface{}) interface{} {
	out := []interface{}{}
	for _, i := range list {
		out = append(out, FixWorkflowStep(i.(pbutil.JSONDict)))
	}
	return out
}

func FixWorkflow(doc pbutil.JSONDict) pbutil.JSONDict {
	doc = fixDict2List(doc, "type", "inputs", "outputs", "hints", "requirements", "steps")
	if x, ok := doc["inputs"]; ok {
		doc["inputs"] = FixInputRecordFieldList(x.([]interface{}))
	}
	if x, ok := doc["outputs"]; ok {
		doc["outputs"] = FixOutputRecordFieldList(x.([]interface{}))
	}
	if x, ok := doc["steps"]; ok {
		doc["steps"] = FixWorkflowStepList(x.([]interface{}))
	}
	return doc
}
