

PATH := ${PATH}:$(GOPATH)/bin
export PATH


PROTO_INC= -I task-execution-schemas -I googleapis

tes:
	protoc \
	$(PROTO_INC) \
	--go_out=tes\
	Mgoogle/protobuf/struct.proto=github.com/golang/protobuf/ptypes/struct:./tes/ \
	task_execution.proto

download:
	go get github.com/golang/protobuf/protoc-gen-go

cwl: cwl_proto
	protoc --go_out=Mgoogle/protobuf/struct.proto=github.com/golang/protobuf/ptypes/struct:./ cwl/cwl.proto

cwl_proto: cwl.avsc
	./tools/cwl-avro-to-lite.py cwl.avsc > cwl/cwl.proto

cwl.avsc:
	python -mschema_salad --print-avro ./common-workflow-language/v1.0/CommonWorkflowLanguage.yml > cwl.avsc

common-workflow-language:
	git clone https://github.com/common-workflow-language/common-workflow-language.git

dag: blank
	protoc \
	-I dag -I googleapis -I task-execution-schemas \
	--go_out=dag\
	Mgoogle/protobuf/struct.proto=github.com/golang/protobuf/ptypes/struct:./dag/ \
	dag/dag.proto

engine: blank
	protoc \
	-I engine -I googleapis \
	--go_out=engine\
	Mgoogle/protobuf/struct.proto=github.com/golang/protobuf/ptypes/struct:./engine/ \
	engine/task_ops.proto

blank:
