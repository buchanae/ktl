package cwl

import (
	structpb "github.com/golang/protobuf/ptypes/struct"
)

func AsMap(src *structpb.Struct) JSONDict {
	out := JSONDict{}
	for k, f := range src.Fields {
		if v, ok := f.Kind.(*structpb.Value_StringValue); ok {
			out[k] = v.StringValue
		} else if v, ok := f.Kind.(*structpb.Value_NumberValue); ok {
			out[k] = v.NumberValue
		} else if v, ok := f.Kind.(*structpb.Value_StructValue); ok {
			out[k] = AsMap(v.StructValue)
		} else if v, ok := f.Kind.(*structpb.Value_BoolValue); ok {
			out[k] = v.BoolValue
		}
	}
	return out
}
