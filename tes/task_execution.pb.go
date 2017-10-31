// Code generated by protoc-gen-go. DO NOT EDIT.
// source: task_execution.proto

/*
Package tes is a generated protocol buffer package.

It is generated from these files:
	task_execution.proto

It has these top-level messages:
	Task
	TaskParameter
	Ports
	Executor
	Resources
	TaskLog
	ExecutorLog
	OutputFileLog
	CreateTaskResponse
	GetTaskRequest
	ListTasksRequest
	ListTasksResponse
	CancelTaskRequest
	CancelTaskResponse
	ServiceInfoRequest
	ServiceInfo
*/
package tes

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "google.golang.org/genproto/googleapis/api/annotations"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type FileType int32

const (
	FileType_FILE      FileType = 0
	FileType_DIRECTORY FileType = 1
)

var FileType_name = map[int32]string{
	0: "FILE",
	1: "DIRECTORY",
}
var FileType_value = map[string]int32{
	"FILE":      0,
	"DIRECTORY": 1,
}

func (x FileType) String() string {
	return proto.EnumName(FileType_name, int32(x))
}
func (FileType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// OUTPUT ONLY
//
// Task states.
type State int32

const (
	State_UNKNOWN      State = 0
	State_QUEUED       State = 1
	State_INITIALIZING State = 2
	State_RUNNING      State = 3
	// An implementation *may* have the ability to pause a task,
	// but this is not required.
	State_PAUSED       State = 4
	State_COMPLETE     State = 5
	State_ERROR        State = 6
	State_SYSTEM_ERROR State = 7
	State_CANCELED     State = 8
)

var State_name = map[int32]string{
	0: "UNKNOWN",
	1: "QUEUED",
	2: "INITIALIZING",
	3: "RUNNING",
	4: "PAUSED",
	5: "COMPLETE",
	6: "ERROR",
	7: "SYSTEM_ERROR",
	8: "CANCELED",
}
var State_value = map[string]int32{
	"UNKNOWN":      0,
	"QUEUED":       1,
	"INITIALIZING": 2,
	"RUNNING":      3,
	"PAUSED":       4,
	"COMPLETE":     5,
	"ERROR":        6,
	"SYSTEM_ERROR": 7,
	"CANCELED":     8,
}

func (x State) String() string {
	return proto.EnumName(State_name, int32(x))
}
func (State) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

// TaskView affects the fields returned by the ListTasks endpoint.
//
// Some of the fields in task can be large strings (e.g. logs),
// which can be a burden on the network. In the default BASIC view,
// these heavyweight fields are not included, however, a client may
// request the FULL version to include these fields.
type TaskView int32

const (
	// Task message will include ONLY the fields:
	//   Task.Id
	//   Task.State
	TaskView_MINIMAL TaskView = 0
	// Task message will include all fields EXCEPT:
	//   Task.ExecutorLog.stdout
	//   Task.ExecutorLog.stderr
	//   TaskParameter.Contents in Task.Inputs
	TaskView_BASIC TaskView = 1
	// Task message includes all fields.
	TaskView_FULL TaskView = 2
)

var TaskView_name = map[int32]string{
	0: "MINIMAL",
	1: "BASIC",
	2: "FULL",
}
var TaskView_value = map[string]int32{
	"MINIMAL": 0,
	"BASIC":   1,
	"FULL":    2,
}

func (x TaskView) String() string {
	return proto.EnumName(TaskView_name, int32(x))
}
func (TaskView) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

// Task describes an instance of a task.
type Task struct {
	// OUTPUT ONLY
	//
	// Task identifier assigned by the server.
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	// OUTPUT ONLY
	State State `protobuf:"varint,2,opt,name=state,enum=tes.State" json:"state,omitempty"`
	// OPTIONAL
	Name string `protobuf:"bytes,3,opt,name=name" json:"name,omitempty"`
	// OPTIONAL
	//
	// Describes the project this task is associated with.
	// Commonly used for billing on cloud providers (AWS, Google Cloud, etc).
	Project string `protobuf:"bytes,4,opt,name=project" json:"project,omitempty"`
	// OPTIONAL
	Description string `protobuf:"bytes,5,opt,name=description" json:"description,omitempty"`
	// OPTIONAL
	//
	// Input files.
	// Inputs will be downloaded and mounted into the executor container.
	Inputs []*TaskParameter `protobuf:"bytes,6,rep,name=inputs" json:"inputs,omitempty"`
	// OPTIONAL
	//
	// Output files.
	// Outputs will be uploaded from the executor container to long-term storage.
	Outputs []*TaskParameter `protobuf:"bytes,7,rep,name=outputs" json:"outputs,omitempty"`
	// OPTIONAL
	//
	// Request that the task be run with these resources.
	Resources *Resources `protobuf:"bytes,8,opt,name=resources" json:"resources,omitempty"`
	// REQUIRED
	//
	// A list of executors to be run, sequentially. Execution stops
	// on the first error.
	Executors []*Executor `protobuf:"bytes,9,rep,name=executors" json:"executors,omitempty"`
	// OPTIONAL
	//
	// Declared volumes.
	// Volumes are shared between executors. Volumes for inputs and outputs are
	// inferred and should not be declared here.
	Volumes []string `protobuf:"bytes,10,rep,name=volumes" json:"volumes,omitempty"`
	// OPTIONAL
	//
	// A key-value map of arbitrary tags.
	Tags map[string]string `protobuf:"bytes,11,rep,name=tags" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// OUTPUT ONLY
	//
	// Task logging information.
	// Normally, this will contain only one entry, but in the case where
	// a task fails and is retried, an entry will be appended to this list.
	Logs []*TaskLog `protobuf:"bytes,12,rep,name=logs" json:"logs,omitempty"`
}

func (m *Task) Reset()                    { *m = Task{} }
func (m *Task) String() string            { return proto.CompactTextString(m) }
func (*Task) ProtoMessage()               {}
func (*Task) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Task) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Task) GetState() State {
	if m != nil {
		return m.State
	}
	return State_UNKNOWN
}

func (m *Task) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Task) GetProject() string {
	if m != nil {
		return m.Project
	}
	return ""
}

func (m *Task) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Task) GetInputs() []*TaskParameter {
	if m != nil {
		return m.Inputs
	}
	return nil
}

func (m *Task) GetOutputs() []*TaskParameter {
	if m != nil {
		return m.Outputs
	}
	return nil
}

func (m *Task) GetResources() *Resources {
	if m != nil {
		return m.Resources
	}
	return nil
}

func (m *Task) GetExecutors() []*Executor {
	if m != nil {
		return m.Executors
	}
	return nil
}

func (m *Task) GetVolumes() []string {
	if m != nil {
		return m.Volumes
	}
	return nil
}

func (m *Task) GetTags() map[string]string {
	if m != nil {
		return m.Tags
	}
	return nil
}

func (m *Task) GetLogs() []*TaskLog {
	if m != nil {
		return m.Logs
	}
	return nil
}

// TaskParameter describes input and output files for a Task.
type TaskParameter struct {
	// OPTIONAL
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	// OPTIONAL
	Description string `protobuf:"bytes,2,opt,name=description" json:"description,omitempty"`
	// REQUIRED, unless "contents" is set.
	//
	// URL in long term storage, for example:
	// s3://my-object-store/file1
	// gs://my-bucket/file2
	// file:///path/to/my/file
	// /path/to/my/file
	// etc...
	Url string `protobuf:"bytes,3,opt,name=url" json:"url,omitempty"`
	// REQUIRED
	//
	// Path of the file inside the container.
	// Must be an absolute path.
	Path string `protobuf:"bytes,4,opt,name=path" json:"path,omitempty"`
	// REQUIRED
	//
	// Type of the file, FILE or DIRECTORY
	Type FileType `protobuf:"varint,5,opt,name=type,enum=tes.FileType" json:"type,omitempty"`
	// OPTIONAL
	//
	// File contents literal.
	// Implementations should support a minimum of 128 KiB in this field and may define its own maximum.
	// UTF-8 encoded
	//
	// If contents is not empty, "url" must be ignored.
	Contents string `protobuf:"bytes,6,opt,name=contents" json:"contents,omitempty"`
}

func (m *TaskParameter) Reset()                    { *m = TaskParameter{} }
func (m *TaskParameter) String() string            { return proto.CompactTextString(m) }
func (*TaskParameter) ProtoMessage()               {}
func (*TaskParameter) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *TaskParameter) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *TaskParameter) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *TaskParameter) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

func (m *TaskParameter) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *TaskParameter) GetType() FileType {
	if m != nil {
		return m.Type
	}
	return FileType_FILE
}

func (m *TaskParameter) GetContents() string {
	if m != nil {
		return m.Contents
	}
	return ""
}

// Ports describes the port binding between the container and host.
// For example, a Docker implementation might map this to `docker run -p host:container`.
type Ports struct {
	// REQUIRED
	//
	// Port number opened inside the container.
	Container uint32 `protobuf:"varint,1,opt,name=container" json:"container,omitempty"`
	// OPTIONAL
	//
	// Port number opened on the host.
	// Defaults to 0, which assigns a random port on the host.
	Host uint32 `protobuf:"varint,2,opt,name=host" json:"host,omitempty"`
}

func (m *Ports) Reset()                    { *m = Ports{} }
func (m *Ports) String() string            { return proto.CompactTextString(m) }
func (*Ports) ProtoMessage()               {}
func (*Ports) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Ports) GetContainer() uint32 {
	if m != nil {
		return m.Container
	}
	return 0
}

func (m *Ports) GetHost() uint32 {
	if m != nil {
		return m.Host
	}
	return 0
}

// Executor describes a command to be executed, and its environment.
type Executor struct {
	// REQUIRED
	//
	// Name of the container image, for example:
	// ubuntu
	// quay.io/aptible/ubuntu
	// gcr.io/my-org/my-image
	// etc...
	ImageName string `protobuf:"bytes,1,opt,name=image_name,json=imageName" json:"image_name,omitempty"`
	// REQUIRED
	//
	// A sequence of program arguments to execute, where the first argument
	// is the program to execute (i.e. argv).
	Cmd []string `protobuf:"bytes,2,rep,name=cmd" json:"cmd,omitempty"`
	// OPTIONAL
	//
	// The working directory that the command will be executed in.
	// Defaults to the directory set by the container image.
	Workdir string `protobuf:"bytes,3,opt,name=workdir" json:"workdir,omitempty"`
	// OPTIONAL
	//
	// Path inside the container to a file which will be piped
	// to the executor's stdin. Must be an absolute path.
	Stdin string `protobuf:"bytes,6,opt,name=stdin" json:"stdin,omitempty"`
	// OPTIONAL
	//
	// Path inside the container to a file where the executor's
	// stdout will be written to. Must be an absolute path.
	Stdout string `protobuf:"bytes,4,opt,name=stdout" json:"stdout,omitempty"`
	// OPTIONAL
	//
	// Path inside the container to a file where the executor's
	// stderr will be written to. Must be an absolute path.
	Stderr string `protobuf:"bytes,5,opt,name=stderr" json:"stderr,omitempty"`
	// OPTIONAL
	//
	// A list of port bindings between the container and host.
	// For example, a Docker implementation might map this to `docker run -p host:container`.
	//
	// Port bindings are included in ExecutorLogs, which allows TES clients
	// to discover port bindings and communicate with running tasks/executors.
	Ports []*Ports `protobuf:"bytes,7,rep,name=ports" json:"ports,omitempty"`
	// OPTIONAL
	//
	// Enviromental variables to set within the container.
	Environ map[string]string `protobuf:"bytes,8,rep,name=environ" json:"environ,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *Executor) Reset()                    { *m = Executor{} }
func (m *Executor) String() string            { return proto.CompactTextString(m) }
func (*Executor) ProtoMessage()               {}
func (*Executor) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Executor) GetImageName() string {
	if m != nil {
		return m.ImageName
	}
	return ""
}

func (m *Executor) GetCmd() []string {
	if m != nil {
		return m.Cmd
	}
	return nil
}

func (m *Executor) GetWorkdir() string {
	if m != nil {
		return m.Workdir
	}
	return ""
}

func (m *Executor) GetStdin() string {
	if m != nil {
		return m.Stdin
	}
	return ""
}

func (m *Executor) GetStdout() string {
	if m != nil {
		return m.Stdout
	}
	return ""
}

func (m *Executor) GetStderr() string {
	if m != nil {
		return m.Stderr
	}
	return ""
}

func (m *Executor) GetPorts() []*Ports {
	if m != nil {
		return m.Ports
	}
	return nil
}

func (m *Executor) GetEnviron() map[string]string {
	if m != nil {
		return m.Environ
	}
	return nil
}

// Resources describes the resources requested by a task.
type Resources struct {
	// OPTIONAL
	//
	// Requested number of CPUs
	CpuCores uint32 `protobuf:"varint,1,opt,name=cpu_cores,json=cpuCores" json:"cpu_cores,omitempty"`
	// OPTIONAL
	//
	// Is the task allowed to run on preemptible compute instances (e.g. AWS Spot)?
	Preemptible bool `protobuf:"varint,2,opt,name=preemptible" json:"preemptible,omitempty"`
	// OPTIONAL
	//
	// Requested RAM required in gigabytes (GB)
	RamGb float64 `protobuf:"fixed64,3,opt,name=ram_gb,json=ramGb" json:"ram_gb,omitempty"`
	// OPTIONAL
	//
	// Requested disk size in gigabytes (GB)
	DiskSizeGb float64 `protobuf:"fixed64,4,opt,name=disk_size_gb,json=diskSizeGb" json:"disk_size_gb,omitempty"`
	// OPTIONAL
	//
	// Request that the task be run in these compute zones.
	Zones []string `protobuf:"bytes,5,rep,name=zones" json:"zones,omitempty"`
}

func (m *Resources) Reset()                    { *m = Resources{} }
func (m *Resources) String() string            { return proto.CompactTextString(m) }
func (*Resources) ProtoMessage()               {}
func (*Resources) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *Resources) GetCpuCores() uint32 {
	if m != nil {
		return m.CpuCores
	}
	return 0
}

func (m *Resources) GetPreemptible() bool {
	if m != nil {
		return m.Preemptible
	}
	return false
}

func (m *Resources) GetRamGb() float64 {
	if m != nil {
		return m.RamGb
	}
	return 0
}

func (m *Resources) GetDiskSizeGb() float64 {
	if m != nil {
		return m.DiskSizeGb
	}
	return 0
}

func (m *Resources) GetZones() []string {
	if m != nil {
		return m.Zones
	}
	return nil
}

// OUTPUT ONLY
//
// TaskLog describes logging information related to a Task.
type TaskLog struct {
	// REQUIRED
	//
	// Logs for each executor
	Logs []*ExecutorLog `protobuf:"bytes,1,rep,name=logs" json:"logs,omitempty"`
	// OPTIONAL
	//
	// Arbitrary logging metadata included by the implementation.
	Metadata map[string]string `protobuf:"bytes,2,rep,name=metadata" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// OPTIONAL
	//
	// When the task started, in RFC 3339 format.
	StartTime string `protobuf:"bytes,3,opt,name=start_time,json=startTime" json:"start_time,omitempty"`
	// OPTIONAL
	//
	// When the task ended, in RFC 3339 format.
	EndTime string `protobuf:"bytes,4,opt,name=end_time,json=endTime" json:"end_time,omitempty"`
	// REQUIRED
	//
	// Information about all output files. Directory outputs are
	// flattened into separate items.
	Outputs []*OutputFileLog `protobuf:"bytes,5,rep,name=outputs" json:"outputs,omitempty"`
}

func (m *TaskLog) Reset()                    { *m = TaskLog{} }
func (m *TaskLog) String() string            { return proto.CompactTextString(m) }
func (*TaskLog) ProtoMessage()               {}
func (*TaskLog) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *TaskLog) GetLogs() []*ExecutorLog {
	if m != nil {
		return m.Logs
	}
	return nil
}

func (m *TaskLog) GetMetadata() map[string]string {
	if m != nil {
		return m.Metadata
	}
	return nil
}

func (m *TaskLog) GetStartTime() string {
	if m != nil {
		return m.StartTime
	}
	return ""
}

func (m *TaskLog) GetEndTime() string {
	if m != nil {
		return m.EndTime
	}
	return ""
}

func (m *TaskLog) GetOutputs() []*OutputFileLog {
	if m != nil {
		return m.Outputs
	}
	return nil
}

// OUTPUT ONLY
//
// ExecutorLog describes logging information related to an Executor.
type ExecutorLog struct {
	// OPTIONAL
	//
	// Time the executor started, in RFC 3339 format.
	StartTime string `protobuf:"bytes,2,opt,name=start_time,json=startTime" json:"start_time,omitempty"`
	// OPTIONAL
	//
	// Time the executor ended, in RFC 3339 format.
	EndTime string `protobuf:"bytes,3,opt,name=end_time,json=endTime" json:"end_time,omitempty"`
	// OPTIONAL
	//
	// Stdout tail.
	// This is not guaranteed to be the entire log.
	// Implementations determine the maximum size.
	Stdout string `protobuf:"bytes,4,opt,name=stdout" json:"stdout,omitempty"`
	// OPTIONAL
	//
	// Stderr tail.
	// This is not guaranteed to be the entire log.
	// Implementations determine the maximum size.
	Stderr string `protobuf:"bytes,5,opt,name=stderr" json:"stderr,omitempty"`
	// REQUIRED
	//
	// Exit code.
	ExitCode int32 `protobuf:"varint,6,opt,name=exit_code,json=exitCode" json:"exit_code,omitempty"`
	// OPTIONAL
	//
	// IP address of the host.
	HostIp string `protobuf:"bytes,7,opt,name=host_ip,json=hostIp" json:"host_ip,omitempty"`
	// OPTIONAL
	//
	// Ports bound between the Executor's container and host.
	//
	// TES clients can use these logs to discover port bindings
	// and communicate with running tasks/executors.
	Ports []*Ports `protobuf:"bytes,8,rep,name=ports" json:"ports,omitempty"`
}

func (m *ExecutorLog) Reset()                    { *m = ExecutorLog{} }
func (m *ExecutorLog) String() string            { return proto.CompactTextString(m) }
func (*ExecutorLog) ProtoMessage()               {}
func (*ExecutorLog) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *ExecutorLog) GetStartTime() string {
	if m != nil {
		return m.StartTime
	}
	return ""
}

func (m *ExecutorLog) GetEndTime() string {
	if m != nil {
		return m.EndTime
	}
	return ""
}

func (m *ExecutorLog) GetStdout() string {
	if m != nil {
		return m.Stdout
	}
	return ""
}

func (m *ExecutorLog) GetStderr() string {
	if m != nil {
		return m.Stderr
	}
	return ""
}

func (m *ExecutorLog) GetExitCode() int32 {
	if m != nil {
		return m.ExitCode
	}
	return 0
}

func (m *ExecutorLog) GetHostIp() string {
	if m != nil {
		return m.HostIp
	}
	return ""
}

func (m *ExecutorLog) GetPorts() []*Ports {
	if m != nil {
		return m.Ports
	}
	return nil
}

// OUTPUT ONLY
//
// OutputFileLog describes a single output file. This describes
// file details after the task has completed successfully,
// for logging purposes.
type OutputFileLog struct {
	// REQUIRED
	//
	// URL of the file in storage, e.g. s3://bucket/file.txt
	Url string `protobuf:"bytes,1,opt,name=url" json:"url,omitempty"`
	// REQUIRED
	//
	// Path of the file inside the container. Must be an absolute path.
	Path string `protobuf:"bytes,2,opt,name=path" json:"path,omitempty"`
	// REQUIRED
	//
	// Size of the file in bytes.
	SizeBytes int64 `protobuf:"varint,3,opt,name=size_bytes,json=sizeBytes" json:"size_bytes,omitempty"`
}

func (m *OutputFileLog) Reset()                    { *m = OutputFileLog{} }
func (m *OutputFileLog) String() string            { return proto.CompactTextString(m) }
func (*OutputFileLog) ProtoMessage()               {}
func (*OutputFileLog) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *OutputFileLog) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

func (m *OutputFileLog) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *OutputFileLog) GetSizeBytes() int64 {
	if m != nil {
		return m.SizeBytes
	}
	return 0
}

// OUTPUT ONLY
//
// CreateTaskResponse describes a response from the CreateTask endpoint.
type CreateTaskResponse struct {
	// REQUIRED
	//
	// Task identifier assigned by the server.
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *CreateTaskResponse) Reset()                    { *m = CreateTaskResponse{} }
func (m *CreateTaskResponse) String() string            { return proto.CompactTextString(m) }
func (*CreateTaskResponse) ProtoMessage()               {}
func (*CreateTaskResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *CreateTaskResponse) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

// GetTaskRequest describes a request to the GetTask endpoint.
type GetTaskRequest struct {
	// REQUIRED
	//
	// Task identifier.
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	// OPTIONAL
	//
	// Affects the fields included in the returned Task messages.
	// See TaskView below.
	View TaskView `protobuf:"varint,2,opt,name=view,enum=tes.TaskView" json:"view,omitempty"`
}

func (m *GetTaskRequest) Reset()                    { *m = GetTaskRequest{} }
func (m *GetTaskRequest) String() string            { return proto.CompactTextString(m) }
func (*GetTaskRequest) ProtoMessage()               {}
func (*GetTaskRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *GetTaskRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *GetTaskRequest) GetView() TaskView {
	if m != nil {
		return m.View
	}
	return TaskView_MINIMAL
}

// ListTasksRequest describes a request to the ListTasks service endpoint.
type ListTasksRequest struct {
	// OPTIONAL
	//
	// Filter the task list to include tasks in this project.
	Project string `protobuf:"bytes,1,opt,name=project" json:"project,omitempty"`
	// OPTIONAL
	//
	// Filter the list to include tasks where the name matches this prefix.
	// If unspecified, no task name filtering is done.
	NamePrefix string `protobuf:"bytes,2,opt,name=name_prefix,json=namePrefix" json:"name_prefix,omitempty"`
	// OPTIONAL
	//
	// Number of tasks to return in one page.
	// Must be less than 2048. Defaults to 256.
	PageSize uint32 `protobuf:"varint,3,opt,name=page_size,json=pageSize" json:"page_size,omitempty"`
	// OPTIONAL
	//
	// Page token is used to retrieve the next page of results.
	// If unspecified, returns the first page of results.
	// See ListTasksResponse.next_page_token
	PageToken string `protobuf:"bytes,4,opt,name=page_token,json=pageToken" json:"page_token,omitempty"`
	// OPTIONAL
	//
	// Affects the fields included in the returned Task messages.
	// See TaskView below.
	View TaskView `protobuf:"varint,5,opt,name=view,enum=tes.TaskView" json:"view,omitempty"`
}

func (m *ListTasksRequest) Reset()                    { *m = ListTasksRequest{} }
func (m *ListTasksRequest) String() string            { return proto.CompactTextString(m) }
func (*ListTasksRequest) ProtoMessage()               {}
func (*ListTasksRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *ListTasksRequest) GetProject() string {
	if m != nil {
		return m.Project
	}
	return ""
}

func (m *ListTasksRequest) GetNamePrefix() string {
	if m != nil {
		return m.NamePrefix
	}
	return ""
}

func (m *ListTasksRequest) GetPageSize() uint32 {
	if m != nil {
		return m.PageSize
	}
	return 0
}

func (m *ListTasksRequest) GetPageToken() string {
	if m != nil {
		return m.PageToken
	}
	return ""
}

func (m *ListTasksRequest) GetView() TaskView {
	if m != nil {
		return m.View
	}
	return TaskView_MINIMAL
}

// OUTPUT ONLY
//
// ListTasksResponse describes a response from the ListTasks endpoint.
type ListTasksResponse struct {
	// REQUIRED
	//
	// List of tasks.
	Tasks []*Task `protobuf:"bytes,1,rep,name=tasks" json:"tasks,omitempty"`
	// OPTIONAL
	//
	// Token used to return the next page of results.
	// See TaskListRequest.next_page_token
	NextPageToken string `protobuf:"bytes,2,opt,name=next_page_token,json=nextPageToken" json:"next_page_token,omitempty"`
}

func (m *ListTasksResponse) Reset()                    { *m = ListTasksResponse{} }
func (m *ListTasksResponse) String() string            { return proto.CompactTextString(m) }
func (*ListTasksResponse) ProtoMessage()               {}
func (*ListTasksResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *ListTasksResponse) GetTasks() []*Task {
	if m != nil {
		return m.Tasks
	}
	return nil
}

func (m *ListTasksResponse) GetNextPageToken() string {
	if m != nil {
		return m.NextPageToken
	}
	return ""
}

// CancelTaskRequest describes a request to the CancelTask endpoint.
type CancelTaskRequest struct {
	// REQUIRED
	//
	// Task identifier.
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *CancelTaskRequest) Reset()                    { *m = CancelTaskRequest{} }
func (m *CancelTaskRequest) String() string            { return proto.CompactTextString(m) }
func (*CancelTaskRequest) ProtoMessage()               {}
func (*CancelTaskRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

func (m *CancelTaskRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

// OUTPUT ONLY
//
// CancelTaskResponse describes a response from the CancelTask endpoint.
type CancelTaskResponse struct {
}

func (m *CancelTaskResponse) Reset()                    { *m = CancelTaskResponse{} }
func (m *CancelTaskResponse) String() string            { return proto.CompactTextString(m) }
func (*CancelTaskResponse) ProtoMessage()               {}
func (*CancelTaskResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

// ServiceInfoRequest describes a request to the ServiceInfo endpoint.
type ServiceInfoRequest struct {
}

func (m *ServiceInfoRequest) Reset()                    { *m = ServiceInfoRequest{} }
func (m *ServiceInfoRequest) String() string            { return proto.CompactTextString(m) }
func (*ServiceInfoRequest) ProtoMessage()               {}
func (*ServiceInfoRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

// OUTPUT ONLY
//
// ServiceInfo describes information about the service,
// such as storage details, resource availability,
// and other documentation.
type ServiceInfo struct {
	// Returns the name of the service, e.g. "ohsu-compbio-funnel".
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	// Returns a documentation string, e.g. "Hey, we're OHSU Comp. Bio!".
	Doc string `protobuf:"bytes,2,opt,name=doc" json:"doc,omitempty"`
	// Lists some, but not necessarily all, storage locations supported by the service.
	//
	// Must be in a valid URL format.
	// e.g.
	// file:///path/to/local/funnel-storage
	// s3://ohsu-compbio-funnel/storage
	// etc.
	Storage []string `protobuf:"bytes,3,rep,name=storage" json:"storage,omitempty"`
}

func (m *ServiceInfo) Reset()                    { *m = ServiceInfo{} }
func (m *ServiceInfo) String() string            { return proto.CompactTextString(m) }
func (*ServiceInfo) ProtoMessage()               {}
func (*ServiceInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

func (m *ServiceInfo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ServiceInfo) GetDoc() string {
	if m != nil {
		return m.Doc
	}
	return ""
}

func (m *ServiceInfo) GetStorage() []string {
	if m != nil {
		return m.Storage
	}
	return nil
}

func init() {
	proto.RegisterType((*Task)(nil), "tes.Task")
	proto.RegisterType((*TaskParameter)(nil), "tes.TaskParameter")
	proto.RegisterType((*Ports)(nil), "tes.Ports")
	proto.RegisterType((*Executor)(nil), "tes.Executor")
	proto.RegisterType((*Resources)(nil), "tes.Resources")
	proto.RegisterType((*TaskLog)(nil), "tes.TaskLog")
	proto.RegisterType((*ExecutorLog)(nil), "tes.ExecutorLog")
	proto.RegisterType((*OutputFileLog)(nil), "tes.OutputFileLog")
	proto.RegisterType((*CreateTaskResponse)(nil), "tes.CreateTaskResponse")
	proto.RegisterType((*GetTaskRequest)(nil), "tes.GetTaskRequest")
	proto.RegisterType((*ListTasksRequest)(nil), "tes.ListTasksRequest")
	proto.RegisterType((*ListTasksResponse)(nil), "tes.ListTasksResponse")
	proto.RegisterType((*CancelTaskRequest)(nil), "tes.CancelTaskRequest")
	proto.RegisterType((*CancelTaskResponse)(nil), "tes.CancelTaskResponse")
	proto.RegisterType((*ServiceInfoRequest)(nil), "tes.ServiceInfoRequest")
	proto.RegisterType((*ServiceInfo)(nil), "tes.ServiceInfo")
	proto.RegisterEnum("tes.FileType", FileType_name, FileType_value)
	proto.RegisterEnum("tes.State", State_name, State_value)
	proto.RegisterEnum("tes.TaskView", TaskView_name, TaskView_value)
}

func init() { proto.RegisterFile("task_execution.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 1371 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x56, 0xcf, 0x73, 0xdb, 0xc4,
	0x17, 0xaf, 0xfc, 0x23, 0x96, 0x9e, 0xe3, 0x7c, 0xd5, 0xfd, 0xa6, 0xa9, 0x70, 0x5b, 0xea, 0xaa,
	0x1d, 0xc8, 0x84, 0x92, 0x0c, 0x81, 0xe1, 0x47, 0x38, 0xa5, 0x8e, 0x9a, 0xf1, 0x60, 0x3b, 0x41,
	0x76, 0x60, 0x0a, 0x9d, 0xf1, 0x28, 0xd2, 0xd6, 0x15, 0xb1, 0xb5, 0x42, 0xbb, 0x4e, 0x9b, 0x32,
	0x5c, 0x18, 0x8e, 0xdc, 0x38, 0x71, 0xe5, 0xc8, 0x30, 0xc3, 0x9f, 0xc2, 0x81, 0x23, 0x57, 0xfe,
	0x10, 0xe6, 0xad, 0x56, 0xb2, 0x1c, 0xd3, 0xce, 0xf4, 0xb6, 0xef, 0xf3, 0x3e, 0xfb, 0xf6, 0xfd,
	0xdc, 0x5d, 0x58, 0x17, 0x1e, 0x3f, 0x1b, 0xd1, 0xe7, 0xd4, 0x9f, 0x89, 0x90, 0x45, 0xdb, 0x71,
	0xc2, 0x04, 0x23, 0x65, 0x41, 0x79, 0xf3, 0xe6, 0x98, 0xb1, 0xf1, 0x84, 0xee, 0x78, 0x71, 0xb8,
	0xe3, 0x45, 0x11, 0x13, 0x1e, 0x32, 0x78, 0x4a, 0xb1, 0xff, 0x2e, 0x43, 0x65, 0xe8, 0xf1, 0x33,
	0xb2, 0x06, 0xa5, 0x30, 0xb0, 0xb4, 0x96, 0xb6, 0x69, 0xb8, 0xa5, 0x30, 0x20, 0x2d, 0xa8, 0x72,
	0xe1, 0x09, 0x6a, 0x95, 0x5a, 0xda, 0xe6, 0xda, 0x2e, 0x6c, 0x0b, 0xca, 0xb7, 0x07, 0x88, 0xb8,
	0xa9, 0x82, 0x10, 0xa8, 0x44, 0xde, 0x94, 0x5a, 0x65, 0xb9, 0x47, 0xae, 0x89, 0x05, 0xb5, 0x38,
	0x61, 0xdf, 0x50, 0x5f, 0x58, 0x15, 0x09, 0x67, 0x22, 0x69, 0x41, 0x3d, 0xa0, 0xdc, 0x4f, 0xc2,
	0x18, 0x8f, 0xb7, 0xaa, 0x52, 0x5b, 0x84, 0xc8, 0x16, 0xac, 0x84, 0x51, 0x3c, 0x13, 0xdc, 0x5a,
	0x69, 0x95, 0x37, 0xeb, 0xbb, 0x44, 0x1e, 0x89, 0xce, 0x1d, 0x7b, 0x89, 0x37, 0xa5, 0x82, 0x26,
	0xae, 0x62, 0x90, 0xfb, 0x50, 0x63, 0x33, 0x21, 0xc9, 0xb5, 0x97, 0x92, 0x33, 0x0a, 0xb9, 0x0f,
	0x46, 0x42, 0x39, 0x9b, 0x25, 0x3e, 0xe5, 0x96, 0xde, 0xd2, 0x36, 0xeb, 0xbb, 0x6b, 0x92, 0xef,
	0x66, 0xa8, 0x3b, 0x27, 0x90, 0x77, 0xc0, 0x48, 0x13, 0xc9, 0x12, 0x6e, 0x19, 0xd2, 0x7a, 0x43,
	0xb2, 0x1d, 0x85, 0xba, 0x73, 0x3d, 0x06, 0x7c, 0xce, 0x26, 0xb3, 0x29, 0xe5, 0x16, 0xb4, 0xca,
	0x18, 0xb0, 0x12, 0xc9, 0xdb, 0x50, 0x11, 0xde, 0x98, 0x5b, 0x75, 0x69, 0xe1, 0xff, 0xb9, 0x7f,
	0xdb, 0x43, 0x6f, 0xcc, 0x9d, 0x48, 0x24, 0x17, 0xae, 0x24, 0x90, 0x16, 0x54, 0x26, 0x6c, 0xcc,
	0xad, 0x55, 0x49, 0x5c, 0xcd, 0x89, 0x5d, 0x36, 0x76, 0xa5, 0xa6, 0xf9, 0x11, 0x18, 0xf9, 0x26,
	0x62, 0x42, 0xf9, 0x8c, 0x5e, 0xa8, 0x4a, 0xe1, 0x92, 0xac, 0x43, 0xf5, 0xdc, 0x9b, 0xcc, 0xd2,
	0x52, 0x19, 0x6e, 0x2a, 0xec, 0x95, 0x3e, 0xd6, 0xec, 0xdf, 0x34, 0x68, 0x2c, 0xe4, 0x24, 0x2f,
	0x9a, 0x56, 0x28, 0xda, 0xa5, 0xd2, 0x94, 0x96, 0x4b, 0x63, 0x42, 0x79, 0x96, 0x4c, 0x54, 0xa5,
	0x71, 0x89, 0x76, 0x62, 0x4f, 0x3c, 0x55, 0x55, 0x96, 0x6b, 0x72, 0x07, 0x2a, 0xe2, 0x22, 0xa6,
	0xb2, 0xb6, 0x6b, 0x2a, 0x67, 0x0f, 0xc3, 0x09, 0x1d, 0x5e, 0xc4, 0xd4, 0x95, 0x2a, 0xd2, 0x04,
	0xdd, 0x67, 0x91, 0xa0, 0x91, 0xac, 0x32, 0x6e, 0xcd, 0x65, 0xfb, 0x13, 0xa8, 0x1e, 0xb3, 0x44,
	0x70, 0x72, 0x13, 0x0c, 0x04, 0xbd, 0x30, 0xa2, 0x89, 0x74, 0xb4, 0xe1, 0xce, 0x01, 0x3c, 0xf9,
	0x29, 0xe3, 0x42, 0xba, 0xd9, 0x70, 0xe5, 0xda, 0xfe, 0xa3, 0x04, 0x7a, 0x56, 0x1d, 0x72, 0x0b,
	0x20, 0x9c, 0x7a, 0x63, 0x3a, 0x2a, 0x04, 0x6a, 0x48, 0xa4, 0x8f, 0xd1, 0x9a, 0x50, 0xf6, 0xa7,
	0x81, 0x55, 0x92, 0xd5, 0xc2, 0x25, 0xd6, 0xf0, 0x19, 0x4b, 0xce, 0x82, 0x30, 0x51, 0x11, 0x66,
	0x22, 0x66, 0x96, 0x8b, 0x20, 0x8c, 0x94, 0xaf, 0xa9, 0x40, 0x36, 0x60, 0x85, 0x8b, 0x80, 0xcd,
	0xb2, 0x1e, 0x57, 0x92, 0xc2, 0x69, 0x92, 0xa8, 0xee, 0x56, 0x12, 0x8e, 0x52, 0x8c, 0x81, 0xa9,
	0x56, 0x4d, 0x47, 0x49, 0x86, 0xea, 0xa6, 0x0a, 0xf2, 0x01, 0xd4, 0x68, 0x74, 0x1e, 0x26, 0x2c,
	0xb2, 0x74, 0xc9, 0x69, 0x2e, 0x34, 0xdc, 0xb6, 0x93, 0x2a, 0xd3, 0xae, 0xc9, 0xa8, 0xcd, 0x3d,
	0x58, 0x2d, 0x2a, 0x5e, 0xab, 0x33, 0x7e, 0xd1, 0xc0, 0xc8, 0xbb, 0x9f, 0xdc, 0x00, 0xc3, 0x8f,
	0x67, 0x23, 0x9f, 0x25, 0x94, 0xab, 0x8c, 0xeb, 0x7e, 0x3c, 0x6b, 0xa3, 0x8c, 0xed, 0x11, 0x27,
	0x94, 0x4e, 0x63, 0x11, 0x9e, 0x4e, 0x52, 0x53, 0xba, 0x5b, 0x84, 0xc8, 0x35, 0x58, 0x49, 0xbc,
	0xe9, 0x68, 0x7c, 0x2a, 0xf3, 0xa7, 0xb9, 0xd5, 0xc4, 0x9b, 0x1e, 0x9e, 0x92, 0x16, 0xac, 0x06,
	0x21, 0x3f, 0x1b, 0xf1, 0xf0, 0x05, 0x45, 0x65, 0x45, 0x2a, 0x01, 0xb1, 0x41, 0xf8, 0x82, 0x1e,
	0x9e, 0xa2, 0x7f, 0x2f, 0x58, 0x44, 0xb9, 0x55, 0x95, 0xd5, 0x48, 0x05, 0xfb, 0xa7, 0x12, 0xd4,
	0xd4, 0x00, 0x90, 0x7b, 0x6a, 0x38, 0x34, 0x99, 0x16, 0x73, 0x21, 0x2d, 0xf9, 0x80, 0x90, 0x0f,
	0x41, 0x9f, 0x52, 0xe1, 0x05, 0x9e, 0xf0, 0x64, 0x61, 0xb3, 0x04, 0x2a, 0x2b, 0xdb, 0x3d, 0xa5,
	0x4c, 0x13, 0x98, 0x73, 0xb1, 0x55, 0xb8, 0xf0, 0x12, 0x31, 0x12, 0x61, 0x7e, 0x91, 0x19, 0x12,
	0x19, 0x86, 0x53, 0x4a, 0xde, 0x00, 0x9d, 0x46, 0x41, 0xaa, 0x54, 0xd7, 0x19, 0x8d, 0x02, 0xa9,
	0x2a, 0x5c, 0x40, 0xd5, 0xc2, 0x05, 0x74, 0x24, 0x31, 0x6c, 0x7a, 0x74, 0x2e, 0xa3, 0x34, 0x3f,
	0x85, 0xc6, 0x82, 0x0b, 0xaf, 0x55, 0xaa, 0x3f, 0x35, 0xa8, 0x17, 0x42, 0xbe, 0xe4, 0x74, 0xe9,
	0x55, 0x4e, 0x97, 0x17, 0x9d, 0x7e, 0xdd, 0xc6, 0xbd, 0x81, 0x37, 0x61, 0x28, 0x46, 0x3e, 0x0b,
	0xa8, 0x1c, 0x81, 0xaa, 0xab, 0x23, 0xd0, 0x66, 0x01, 0x25, 0xd7, 0xa1, 0x86, 0xb3, 0x37, 0x0a,
	0x63, 0xab, 0x96, 0xee, 0x42, 0xb1, 0x13, 0xcf, 0xdb, 0x5d, 0x7f, 0x49, 0xbb, 0xdb, 0x43, 0x68,
	0x2c, 0x24, 0x2a, 0xbb, 0x5f, 0xb4, 0xe5, 0xfb, 0xa5, 0x54, 0xb8, 0x5f, 0x30, 0x70, 0x6c, 0xa5,
	0xd3, 0x0b, 0x41, 0xb9, 0x8c, 0xad, 0xec, 0x1a, 0x88, 0x3c, 0x40, 0xc0, 0xbe, 0x07, 0xa4, 0x9d,
	0x50, 0x4f, 0x50, 0xac, 0xba, 0x4b, 0x79, 0xcc, 0x22, 0x4e, 0x2f, 0xbf, 0x6b, 0x76, 0x1b, 0xd6,
	0x0e, 0xa9, 0x48, 0x29, 0xdf, 0xce, 0x28, 0x17, 0x4b, 0x2f, 0xdf, 0x1d, 0xa8, 0x9c, 0x87, 0xf4,
	0x99, 0x7a, 0xf8, 0x1a, 0x79, 0x23, 0x7d, 0x11, 0xd2, 0x67, 0xae, 0x54, 0xd9, 0xbf, 0x6b, 0x60,
	0x76, 0x43, 0x2e, 0xcd, 0xf0, 0xcc, 0x4e, 0xe1, 0xed, 0xd3, 0x16, 0xdf, 0xbe, 0xdb, 0x50, 0xc7,
	0xbb, 0x68, 0x14, 0x27, 0xf4, 0x49, 0xf8, 0x5c, 0xc5, 0x04, 0x08, 0x1d, 0x4b, 0x04, 0x13, 0x1d,
	0xe3, 0x8d, 0x85, 0xc1, 0xc8, 0xc0, 0x1a, 0xae, 0x8e, 0x00, 0x8e, 0x09, 0x86, 0x2d, 0x95, 0x82,
	0x9d, 0xd1, 0x48, 0x55, 0x4e, 0xd2, 0x87, 0x08, 0xe4, 0xee, 0x56, 0x5f, 0xee, 0xee, 0x63, 0xb8,
	0x5a, 0xf0, 0x56, 0x25, 0xe6, 0x36, 0x54, 0xf1, 0xd3, 0x90, 0x8d, 0x96, 0x91, 0x6f, 0x74, 0x53,
	0x9c, 0xbc, 0x05, 0xff, 0x8b, 0xe8, 0x73, 0x31, 0x2a, 0x1c, 0x9e, 0x7a, 0xde, 0x40, 0xf8, 0x38,
	0x73, 0xc0, 0xbe, 0x0b, 0x57, 0xdb, 0x5e, 0xe4, 0xd3, 0xc9, 0x2b, 0x92, 0x6a, 0xaf, 0x03, 0x29,
	0x92, 0x52, 0x1f, 0x10, 0x1d, 0xd0, 0xe4, 0x3c, 0xf4, 0x69, 0x27, 0x7a, 0xc2, 0xd4, 0x5e, 0xbb,
	0x07, 0xf5, 0x02, 0xfa, 0x9f, 0x4f, 0x96, 0x09, 0xe5, 0x80, 0xf9, 0xca, 0x1f, 0x5c, 0x62, 0xf6,
	0xb9, 0x60, 0x89, 0x37, 0xc6, 0x04, 0xca, 0x87, 0x58, 0x89, 0x5b, 0x77, 0x41, 0xcf, 0x5e, 0x21,
	0xa2, 0x43, 0xe5, 0x61, 0xa7, 0xeb, 0x98, 0x57, 0x48, 0x03, 0x8c, 0x83, 0x8e, 0xeb, 0xb4, 0x87,
	0x47, 0xee, 0x23, 0x53, 0xdb, 0xfa, 0x51, 0x83, 0xaa, 0xfc, 0xdd, 0x90, 0x3a, 0xd4, 0x4e, 0xfa,
	0x9f, 0xf5, 0x8f, 0xbe, 0xec, 0x9b, 0x57, 0x08, 0xc0, 0xca, 0xe7, 0x27, 0xce, 0x89, 0x73, 0x60,
	0x6a, 0xc4, 0x84, 0xd5, 0x4e, 0xbf, 0x33, 0xec, 0xec, 0x77, 0x3b, 0x5f, 0x75, 0xfa, 0x87, 0x66,
	0x09, 0xa9, 0xee, 0x49, 0xbf, 0x8f, 0x42, 0x19, 0xa9, 0xc7, 0xfb, 0x27, 0x03, 0xe7, 0xc0, 0xac,
	0x90, 0x55, 0xd0, 0xdb, 0x47, 0xbd, 0xe3, 0xae, 0x33, 0x74, 0xcc, 0x2a, 0x31, 0xa0, 0xea, 0xb8,
	0xee, 0x91, 0x6b, 0xae, 0xa0, 0x8d, 0xc1, 0xa3, 0xc1, 0xd0, 0xe9, 0x8d, 0x52, 0xa4, 0x26, 0xa9,
	0xfb, 0xfd, 0xb6, 0xd3, 0x75, 0x0e, 0x4c, 0x7d, 0xeb, 0x3e, 0xe8, 0x59, 0xed, 0xd0, 0x7a, 0xaf,
	0xd3, 0xef, 0xf4, 0xf6, 0xbb, 0xe6, 0x15, 0xb4, 0xf1, 0x60, 0x7f, 0xd0, 0x69, 0x9b, 0x9a, 0x8c,
	0xe1, 0xa4, 0xdb, 0x35, 0x4b, 0xbb, 0xbf, 0x96, 0xa1, 0x8e, 0x74, 0x95, 0x2d, 0xf2, 0xb5, 0xec,
	0xed, 0x62, 0xee, 0xae, 0xa7, 0xdf, 0xb6, 0xa5, 0x1c, 0x37, 0xcd, 0xcb, 0x0a, 0xfb, 0xcd, 0x1f,
	0xfe, 0xfa, 0xe7, 0xe7, 0x92, 0x45, 0x36, 0x76, 0xce, 0xdf, 0xdb, 0x91, 0x1d, 0xb0, 0xc3, 0x53,
	0xf5, 0xbb, 0x21, 0x9a, 0x7a, 0x08, 0x30, 0x1f, 0x2f, 0x32, 0x6f, 0x97, 0x66, 0x7a, 0xc6, 0xf2,
	0xe8, 0xd9, 0xeb, 0xd2, 0xe2, 0x9a, 0x6d, 0xe4, 0x16, 0xf7, 0xb4, 0x2d, 0xd2, 0x03, 0x23, 0x6f,
	0x46, 0x72, 0x4d, 0xee, 0xbd, 0x3c, 0x4a, 0xcd, 0x8d, 0xcb, 0xb0, 0xb2, 0x78, 0x55, 0x5a, 0xac,
	0x93, 0xb9, 0x45, 0xb2, 0x0f, 0x35, 0x35, 0xcf, 0x24, 0xfd, 0x63, 0x2d, 0x4e, 0x77, 0x73, 0xee,
	0xa8, 0xbd, 0x21, 0x77, 0x9b, 0x64, 0x6d, 0x1e, 0xe1, 0x77, 0x61, 0xf0, 0x3d, 0x79, 0x0c, 0x30,
	0xef, 0x4d, 0x92, 0x9e, 0xbd, 0xd4, 0xd1, 0x59, 0x98, 0xcb, 0x4d, 0x7c, 0x4b, 0x9a, 0xbd, 0x6e,
	0x5f, 0x5b, 0x34, 0xbb, 0xe7, 0x4b, 0xea, 0xe9, 0x8a, 0xfc, 0x68, 0xbf, 0xff, 0x6f, 0x00, 0x00,
	0x00, 0xff, 0xff, 0xc1, 0xf6, 0xe4, 0x85, 0xa3, 0x0b, 0x00, 0x00,
}
