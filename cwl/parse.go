package cwl

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

func mapNormalize(v interface{}) interface{} {
	if base, ok := v.(map[interface{}]interface{}); ok {
		out := JSONDict{}
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

func YamlLoad(path string) (JSONDict, error) {
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	doc := make(map[interface{}]interface{})
	err = yaml.Unmarshal(source, &doc)
	out := mapNormalize(doc)
	return out.(JSONDict), nil
}

func InputParse(path string, mapper FileMapper) (JSONDict, error) {
	doc, err := YamlLoad(path)

	x, _ := filepath.Abs(path)
	base_path := filepath.Dir(x)

	out := AdjustInputs(doc, base_path, mapper).(JSONDict)
	return out, err
}

func AdjustInputs(input interface{}, basePath string, mapper FileMapper) interface{} {
	if base, ok := input.(JSONDict); ok {
		out := JSONDict{}
		if class, ok := base["class"]; ok {
			if class == "File" {
				for k, v := range base {
					if k == "path" {
						out["path"] = mapper.Input2Storage(basePath, v.(string))
					} else if k == "location" {
						out["location"] = mapper.Input2Storage(basePath, v.(string))
					} else {
						out[k] = v
					}
				}
			} else if class == "Directory" {
				for k, v := range base {
					if k == "path" {
						out["path"] = mapper.Input2Storage(basePath, v.(string))
					} else if k == "location" {
						out["location"] = mapper.Input2Storage(basePath, v.(string))
					} else {
						out[k] = v
					}
				}
			} else {
				log.Printf("Unknown class type: %s", class)
			}
		} else {
			for k, v := range base {
				out[k] = AdjustInputs(v, basePath, mapper)
			}
		}
		return out
	} else if base, ok := input.([]interface{}); ok {
		out := []interface{}{}
		for _, i := range base {
			out = append(out, AdjustInputs(i, basePath, mapper))
		}
		return out
	}
	return input
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
				return CWLGraph{}, fmt.Errorf("Unable to parse file")
			}
			return CWLGraph{Main: "#", Elements: map[string]CWLDoc{"#": cmd}}, nil
		} else if class == "Workflow" {
			fixed_doc := FixWorkflow(doc)
			jdoc, err := json.MarshalIndent(fixed_doc, "", "   ")
			if err != nil {
				log.Printf("%s", err)
			}
			log.Printf("Parsed: %s\n", jdoc)
			cmd := Workflow{}

			umarsh := jsonpb.Unmarshaler{AllowUnknownFields: true}
			err = umarsh.Unmarshal(strings.NewReader(string(jdoc)), &cmd)
			if err != nil {
				log.Printf("SchemaParseError: %s", err)
				return CWLGraph{}, fmt.Errorf("Unable to parse file")
			}
			return CWLGraph{Main: "#", Elements: map[string]CWLDoc{"#": cmd}}, nil
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
	if _, ok := i.(JSONDict); ok {
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

func fixDict2List(doc JSONDict, typeField string, fields ...string) JSONDict {
	out := JSONDict{}
	for k, v := range doc {
		if contains(fields, k) {
			if isDict(v) {
				nv := []interface{}{}
				for ek, ev := range v.(JSONDict) {
					if isString(ev) || isList(ev) {
						i := JSONDict{"id": ek, typeField: ev}
						nv = append(nv, i)
					} else {
						i := ev.(JSONDict)
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

func fixForceList(doc JSONDict, fields ...string) JSONDict {
	out := JSONDict{}
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

func FixDataRecord(doc interface{}) JSONDict {
	if isString(doc) {
		return JSONDict{"string_value": doc.(string)}
	}
	if isDict(doc) {
		return JSONDict{"struct_value": doc.(JSONDict)}
	}
	if isList(doc) {
		return JSONDict{"list_value": doc.([]interface{})}
	}
	if isFloat(doc) {
		return JSONDict{"float_value": doc.(float32)}
	}
	if isInt(doc) {
		return JSONDict{"int_value": doc}
	}
	log.Printf("Unknown Type: %#T", doc)
	return JSONDict{}
}

func FixTypeRecord(doc interface{}) JSONDict {
	if isString(doc) {
		doc_string := doc.(string)
		if strings.HasSuffix(doc_string, "[]") {
			return JSONDict{"array": JSONDict{"items": FixTypeRecord(doc_string[:len(doc_string)-2])}}
		} else {
			return JSONDict{"name": doc}
		}
	}
	if isDict(doc) {
		doc_dict := doc.(JSONDict)
		if doc_dict["type"] == "array" {
			o := JSONDict{"array": JSONDict{"items": FixTypeRecord(doc_dict["items"])}}
			if x, ok := doc_dict["inputBinding"]; ok {
				o["inputBinding"] = x
			}
			return o
		}
		if doc_dict["type"] == "record" {
			doc_fields_list := doc_dict["fields"].([]interface{})
			fields := []interface{}{}
			for _, i := range doc_fields_list {
				field_doc := i.(JSONDict)
				fields = append(fields, JSONDict{"name": field_doc["name"], "type": FixTypeRecord(field_doc["type"])})
			}
			return JSONDict{"record": JSONDict{"name": doc_dict["name"], "fields": fields}}
			if doc_dict["type"] == "enum" {
				return JSONDict{"enum": JSONDict{"name": doc_dict["name"], "symbols": doc_dict["symbols"]}}
			}
		}
	}
	if isList(doc) {
		doc_list := doc.([]interface{})
		t := []interface{}{}
		for _, i := range doc_list {
			t = append(t, FixTypeRecord(i))
		}
		return JSONDict{"oneof": JSONDict{"types": t}}
	}
	return JSONDict{}
}

func FixInputRecordField(doc JSONDict) JSONDict {
	doc = fixForceList(doc, "doc")
	doc["type"] = FixTypeRecord(doc["type"])
	if x, ok := doc["default"]; ok {
		doc["default"] = FixDataRecord(x)
	}
	return doc
}

func FixInputRecordFieldList(list []interface{}) interface{} {
	out := make([]interface{}, len(list))
	for i := range list {
		i_doc := list[i].(JSONDict)
		out[i] = FixInputRecordField(i_doc)
	}
	return out
}

func FixCommandOutputBinding(doc JSONDict) JSONDict {
	doc = fixForceList(doc, "glob")
	return doc
}

func FixOutputRecordField(doc JSONDict) JSONDict {
	doc = fixForceList(doc, "doc", "outputSource")
	doc["type"] = FixTypeRecord(doc["type"])
	if x, ok := doc["outputBinding"]; ok {
		doc["outputBinding"] = FixCommandOutputBinding(x.(JSONDict))
	}
	return doc
}

func FixOutputRecordFieldList(list []interface{}) interface{} {
	out := make([]interface{}, len(list))
	for i := range list {
		i_doc := list[i].(JSONDict)
		out[i] = FixOutputRecordField(i_doc)
	}
	return out
}

func FixCommandLineBinding(doc interface{}) JSONDict {
	if isString(doc) {
		return JSONDict{"valueFrom": doc}
	}
	return doc.(JSONDict)
}

func FixCommandLineBindingList(list []interface{}) interface{} {
	out := []interface{}{}
	for _, i := range list {
		out = append(out, FixCommandLineBinding(i))
	}
	return out
}

func FixCommandLineTool(doc JSONDict) JSONDict {
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
		return JSONDict{"id": x}
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

func FixWorkflowStepInput(x interface{}) JSONDict {
	if i, ok := x.(string); ok {
		return JSONDict{"source": []interface{}{i}}
	}
	if i, ok := x.(JSONDict); ok {
		return i
	}
	return JSONDict{}
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

func FixWorkflowStep(doc JSONDict) JSONDict {
	if x, ok := doc["in"]; ok {
		if x_list, ok := x.([]interface{}); ok {
			doc["in"] = FixWorkflowStepInputList(x_list)
		}
		if x_map, ok := x.(JSONDict); ok {
			doc["in"] = FixWorkflowStepInputMap(x_map)
		}
	}
	if x, ok := doc["out"]; ok {
		doc["out"] = FixWorkflowStepOutputList(x.([]interface{}))
	}
	if x, ok := doc["run"]; ok {
		if x_str, ok := x.(string); ok {
			doc["run"] = JSONDict{"path": x_str}
		}
	}
	return doc
}

func FixWorkflowStepList(list []interface{}) interface{} {
	out := []interface{}{}
	for _, i := range list {
		out = append(out, FixWorkflowStep(i.(JSONDict)))
	}
	return out
}

func FixWorkflow(doc JSONDict) JSONDict {
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
