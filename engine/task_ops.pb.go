// Code generated by protoc-gen-go. DO NOT EDIT.
// source: task_ops.proto

/*
Package engine is a generated protocol buffer package.

It is generated from these files:
	task_ops.proto

It has these top-level messages:
	OutputGlob
	PostProcessStep
	PostProcess
*/
package engine

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type OutputGlob struct {
	ParamName string `protobuf:"bytes,1,opt,name=param_name,json=paramName" json:"param_name,omitempty"`
	Glob      string `protobuf:"bytes,2,opt,name=glob" json:"glob,omitempty"`
}

func (m *OutputGlob) Reset()                    { *m = OutputGlob{} }
func (m *OutputGlob) String() string            { return proto.CompactTextString(m) }
func (*OutputGlob) ProtoMessage()               {}
func (*OutputGlob) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *OutputGlob) GetParamName() string {
	if m != nil {
		return m.ParamName
	}
	return ""
}

func (m *OutputGlob) GetGlob() string {
	if m != nil {
		return m.Glob
	}
	return ""
}

type PostProcessStep struct {
	// Types that are valid to be assigned to Step:
	//	*PostProcessStep_GlobOutput
	Step isPostProcessStep_Step `protobuf_oneof:"step"`
}

func (m *PostProcessStep) Reset()                    { *m = PostProcessStep{} }
func (m *PostProcessStep) String() string            { return proto.CompactTextString(m) }
func (*PostProcessStep) ProtoMessage()               {}
func (*PostProcessStep) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type isPostProcessStep_Step interface {
	isPostProcessStep_Step()
}

type PostProcessStep_GlobOutput struct {
	GlobOutput *OutputGlob `protobuf:"bytes,1,opt,name=glob_output,json=globOutput,oneof"`
}

func (*PostProcessStep_GlobOutput) isPostProcessStep_Step() {}

func (m *PostProcessStep) GetStep() isPostProcessStep_Step {
	if m != nil {
		return m.Step
	}
	return nil
}

func (m *PostProcessStep) GetGlobOutput() *OutputGlob {
	if x, ok := m.GetStep().(*PostProcessStep_GlobOutput); ok {
		return x.GlobOutput
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*PostProcessStep) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _PostProcessStep_OneofMarshaler, _PostProcessStep_OneofUnmarshaler, _PostProcessStep_OneofSizer, []interface{}{
		(*PostProcessStep_GlobOutput)(nil),
	}
}

func _PostProcessStep_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*PostProcessStep)
	// step
	switch x := m.Step.(type) {
	case *PostProcessStep_GlobOutput:
		b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.GlobOutput); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("PostProcessStep.Step has unexpected type %T", x)
	}
	return nil
}

func _PostProcessStep_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*PostProcessStep)
	switch tag {
	case 1: // step.glob_output
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(OutputGlob)
		err := b.DecodeMessage(msg)
		m.Step = &PostProcessStep_GlobOutput{msg}
		return true, err
	default:
		return false, nil
	}
}

func _PostProcessStep_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*PostProcessStep)
	// step
	switch x := m.Step.(type) {
	case *PostProcessStep_GlobOutput:
		s := proto.Size(x.GlobOutput)
		n += proto.SizeVarint(1<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

type PostProcess struct {
	Steps []*PostProcessStep `protobuf:"bytes,1,rep,name=steps" json:"steps,omitempty"`
}

func (m *PostProcess) Reset()                    { *m = PostProcess{} }
func (m *PostProcess) String() string            { return proto.CompactTextString(m) }
func (*PostProcess) ProtoMessage()               {}
func (*PostProcess) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *PostProcess) GetSteps() []*PostProcessStep {
	if m != nil {
		return m.Steps
	}
	return nil
}

func init() {
	proto.RegisterType((*OutputGlob)(nil), "engine.OutputGlob")
	proto.RegisterType((*PostProcessStep)(nil), "engine.PostProcessStep")
	proto.RegisterType((*PostProcess)(nil), "engine.PostProcess")
}

func init() { proto.RegisterFile("task_ops.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 193 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x8f, 0x41, 0xab, 0x82, 0x50,
	0x10, 0x85, 0x9f, 0xef, 0xf9, 0x04, 0xe7, 0xc2, 0x7b, 0x70, 0x37, 0xb9, 0x09, 0xc4, 0x95, 0x9b,
	0x5c, 0x18, 0xed, 0x82, 0xa0, 0x4d, 0xad, 0x4a, 0xec, 0x07, 0xc8, 0x35, 0x06, 0x89, 0xd4, 0xb9,
	0x38, 0xe3, 0xff, 0x0f, 0xaf, 0x44, 0xd1, 0x6e, 0x38, 0xe7, 0xf0, 0x7d, 0x0c, 0xfc, 0x89, 0xe1,
	0x7b, 0x45, 0x96, 0x33, 0x3b, 0x90, 0x90, 0x0e, 0xb0, 0x6f, 0x6e, 0x3d, 0x26, 0x3b, 0x80, 0xf3,
	0x28, 0x76, 0x94, 0x43, 0x4b, 0xb5, 0x5e, 0x02, 0x58, 0x33, 0x98, 0xae, 0xea, 0x4d, 0x87, 0x91,
	0x17, 0x7b, 0x69, 0x58, 0x86, 0x2e, 0x39, 0x99, 0x0e, 0xb5, 0x06, 0xbf, 0x69, 0xa9, 0x8e, 0xbe,
	0x5d, 0xe1, 0xee, 0xa4, 0x80, 0xff, 0x82, 0x58, 0x8a, 0x81, 0xae, 0xc8, 0x7c, 0x11, 0xb4, 0x7a,
	0x03, 0x6a, 0xaa, 0x2a, 0x72, 0x60, 0x87, 0x51, 0xb9, 0xce, 0x66, 0x63, 0xf6, 0xd2, 0x1d, 0xbf,
	0x4a, 0x98, 0x86, 0x73, 0xb2, 0x0f, 0xc0, 0x67, 0x41, 0x9b, 0x6c, 0x41, 0xbd, 0x11, 0xf5, 0x0a,
	0x7e, 0xa7, 0x98, 0x23, 0x2f, 0xfe, 0x49, 0x55, 0xbe, 0x78, 0x72, 0x3e, 0xac, 0xe5, 0xbc, 0xaa,
	0x03, 0xf7, 0xdf, 0xfa, 0x11, 0x00, 0x00, 0xff, 0xff, 0xac, 0x01, 0x12, 0xa2, 0xf1, 0x00, 0x00,
	0x00,
}
