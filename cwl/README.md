
Building CWL proto
------------------




```
git clone https://github.com/common-workflow-language/common-workflow-language.git
virtualenv venv
. venv/bin/activate
pip install schema-salad
python -mschema_salad --print-avro ./common-workflow-language/v1.0/CommonWorkflowLanguage.yml > cwl.avsc
../tools/cwl-avro-to-lite.py cwl.avsc > cwl.proto
```


Build Proto
```
go get github.com/golang/protobuf/protoc-gen-go
cd src/github.com/ohsu-comp-bio/ktl/cwl && protoc \
--go_out=Mgoogle/protobuf/struct.proto=github.com/golang/protobuf/ptypes/struct:./ \
cwl.proto
```
