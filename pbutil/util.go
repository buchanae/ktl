package pbutil

import (
	"log"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

type JSONDict map[string]interface{}

func AsMap(src *structpb.Struct) JSONDict {
	out := JSONDict{}
	for k, f := range src.Fields {
		if v, ok := f.Kind.(*structpb.Value_StringValue); ok {
			out[k] = v.StringValue
		} else if v, ok := f.Kind.(*structpb.Value_NumberValue); ok {
			out[k] = v.NumberValue
		} else if v, ok := f.Kind.(*structpb.Value_StructValue); ok {
			out[k] = AsMap(v.StructValue)
		} else if v, ok := f.Kind.(*structpb.Value_ListValue); ok {
			a := make([]interface{}, len(v.ListValue.Values))
			for i := range v.ListValue.Values {
					a[i] = AsValue(v.ListValue.Values[i])
			}
			out[k] = a
		} else if v, ok := f.Kind.(*structpb.Value_BoolValue); ok {
			out[k] = v.BoolValue
		}
	}
	return out
}

func AsValue(src *structpb.Value) interface{} {
	if v, ok := src.Kind.(*structpb.Value_StringValue); ok {
		return v.StringValue
	} else if v, ok := src.Kind.(*structpb.Value_NumberValue); ok {
		return v.NumberValue
	} else if v, ok := src.Kind.(*structpb.Value_StructValue); ok {
		return AsMap(v.StructValue)
	} else if v, ok := src.Kind.(*structpb.Value_ListValue); ok {
		a := make([]interface{}, len(v.ListValue.Values))
		for i := range v.ListValue.Values {
				a[i] = AsValue(v.ListValue.Values[i])
		}
		return a
	} else if v, ok := src.Kind.(*structpb.Value_BoolValue); ok {
		return v.BoolValue
	}
	return nil
}


func (self JSONDict) AsStruct() *structpb.Struct {
	return WrapValue(self).GetStructValue()
}




func WrapValue(value interface{}) *structpb.Value {
	switch v := value.(type) {
	case string:
		return &structpb.Value{Kind: &structpb.Value_StringValue{v}}
	case int:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{float64(v)}}
	case int64:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{float64(v)}}
	case float64:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{float64(v)}}
	case bool:
		return &structpb.Value{Kind: &structpb.Value_BoolValue{v}}
	case *structpb.Value:
		return v
	case []interface{}:
		o := make([]*structpb.Value, len(v))
		for i, k := range v {
			wv := WrapValue(k)
			o[i] = wv
		}
		return &structpb.Value{Kind: &structpb.Value_ListValue{&structpb.ListValue{Values: o}}}
	case []string:
		o := make([]*structpb.Value, len(v))
		for i, k := range v {
			wv := &structpb.Value{Kind: &structpb.Value_StringValue{k}}
			o[i] = wv
		}
		return &structpb.Value{Kind: &structpb.Value_ListValue{&structpb.ListValue{Values: o}}}
	case map[string]interface{}:
		o := &structpb.Struct{Fields: map[string]*structpb.Value{}}
		for k, v := range v {
			wv := WrapValue(v)
			o.Fields[k] = wv
		}
		return &structpb.Value{Kind: &structpb.Value_StructValue{o}}
	case JSONDict:
		o := &structpb.Struct{Fields: map[string]*structpb.Value{}}
		for k, v := range v {
			wv := WrapValue(v)
			o.Fields[k] = wv
		}
		return &structpb.Value{Kind: &structpb.Value_StructValue{o}}
	case map[string]float64:
		o := &structpb.Struct{Fields: map[string]*structpb.Value{}}
		for k, v := range v {
			wv := WrapValue(v)
			o.Fields[k] = wv
		}
		return &structpb.Value{Kind: &structpb.Value_StructValue{o}}
	default:
		log.Printf("unknown data type: %T", value)
	}
	return nil
}
