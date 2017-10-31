

PATH := ${PATH}:$(GOPATH)/bin
export PATH


PROTO_INC= -I task-execution-schemas -I googleapis

proto:
	protoc \
	$(PROTO_INC) \
	--go_out=tes\
	Mgoogle/protobuf/struct.proto=github.com/golang/protobuf/ptypes/struct:./tes/ \
	task_execution.proto

download:
	go get github.com/golang/protobuf/protoc-gen-go