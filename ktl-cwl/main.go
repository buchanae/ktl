package main

import (
	"flag"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/ktl/cwl"
	"github.com/ohsu-comp-bio/ktl/tes"
	"io/ioutil"
	"log"
	"os"
	//"path/filepath"
	"strings"
	//"time"
)

func main() {
	var version_flag = flag.Bool("version", false, "version")
	//var tmp_outdir_prefix_flag = flag.String("tmp-outdir-prefix", "./", "Temp output prefix")
	//var tmpdir_prefix_flag = flag.String("tmpdir-prefix", "/tmp", "Tempdir prefix")
	var outdir = flag.String("outdir", "./", "Outdir")
	var tes_flag = flag.Bool("tes", false, "TES Job Output")
	var quiet_flag = flag.Bool("quiet", false, "quiet")
	//var out_path = flag.String("out", false, "outdir")
	flag.Parse()

	if *version_flag {
		fmt.Printf("cwlgo v0.0.1\n")
		return
	}

	if *quiet_flag {
		log.SetOutput(ioutil.Discard)
	}

	//tmp_outdir_prefix, _ := filepath.Abs(*tmp_outdir_prefix_flag)
	//tmpdir_prefix, _ := filepath.Abs(*tmpdir_prefix_flag)

	cwl_path := flag.Arg(0)
	element_id := ""
	if strings.Contains(cwl_path, "#") {
		tmp := strings.Split(cwl_path, "#")
		cwl_path = tmp[0]
		element_id = tmp[1]
	}
	cwl_docs, err := cwl.Parse(cwl_path)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Unable to parse CWL document: %s\n", err))
		if _, ok := err.(cwl.UnsupportedRequirement); ok {
			os.Exit(33)
		}
		os.Exit(1)
	}
	//log.Printf("CWLDoc: %#v", cwl_docs)
	var inputs cwl.JSONDict
	mapper := cwl.URLDockerMapper{*outdir}
	if len(flag.Args()) == 1 {
		inputs = cwl.JSONDict{}
	} else {
		var err error
		inputs, err = cwl.InputParse(flag.Arg(1), mapper)
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Unable to parse Input document: %s\n", err))
			os.Exit(1)
		}
	}

	if cwl_docs.Main == "" {
		if element_id == "" {
			os.Stderr.WriteString(fmt.Sprintf("Need to define element ID\n"))
			os.Exit(1)
		}
		cwl_docs.Main = element_id
	}

	var ok bool
	var cwl_doc cwl.CWLDoc
	cwl_doc, ok = cwl_docs.Elements[cwl_docs.Main]
	if !ok {
		cwl_doc, ok = cwl_docs.Elements["#"+cwl_docs.Main]
	}
	if !ok {
		os.Stderr.WriteString(fmt.Sprintf("Element %s not found\n", cwl_docs.Main))
		os.Exit(1)
	}

	log.Printf("%#v\n", cwl_doc)
	log.Printf("%#v\n", inputs)

	cmd := cwl_doc.CommandLineTool()

	env := cmd.SetDefaults(cwl.Environment{Inputs: inputs})

	if *tes_flag {
		tes_doc, err := tes.Render(cmd, mapper, env)
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Command line render failed %s\n", err))
			os.Exit(1)
		}
		m := jsonpb.Marshaler{}
		m.Indent = " "
		tmes, _ := m.MarshalToString(&tes_doc)
		fmt.Printf("%s\n", tmes)
	} else {
		cmd_line, err := cmd.Render(mapper, env)
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Command line render failed %s\n", err))
			os.Exit(1)
		}
		fmt.Printf("%s\n", strings.Join(cmd_line, " "))
	}

	outputs, _ := cmd.GetOutputMapping(env)

	for _, out := range outputs {
		log.Printf("OUTPUT %s Glob: %s\n", out.Id, out.Glob)
	}

}
