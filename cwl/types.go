package cwl

type JSONDict map[string]interface{}

type CWLDoc interface {
	CommandLineTool() CommandLineTool
}

type CWLParser struct {
	Path string
	//Schemas  map[string]Schema
	Elements map[string]CWLDoc
}

type CWLGraph struct {
	Elements map[string]CWLDoc
	Main     string
}

type UnsupportedRequirement struct {
	Message string
}

func (e UnsupportedRequirement) Error() string {
	return e.Message
}
