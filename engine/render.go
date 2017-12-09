package engine

import (
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/ktl/cwl"
	"log"
	"path/filepath"
)

func Render(cmd cwl.CommandLineTool, mapper cwl.FileMapper, env cwl.Environment) (tes.Task, PostProcess, error) {
	cmd_line, err := cmd.Render(mapper, env)
	if err != nil {
		return tes.Task{}, PostProcess{}, err
	}

	out := tes.Task{}
	exec := tes.Executor{}
	exec.Command = cmd_line
	exec.Image = cmd.GetImageName()
	if exec.Image == "" {
		exec.Image = env.DefaultImage
	}
	exec.Workdir = cwl.DOCKER_WORK_DIR
	if cmd.Stdin != "" {
		s, err := cmd.RenderStdinPath(mapper, env)
		if err != nil {
			return out, PostProcess{}, err
		}
		exec.Stdin = s
	}
	if cmd.Stdout != "" {
		exec.Stdout = filepath.Join(cwl.DOCKER_WORK_DIR, cmd.Stdout)
	} else {
		exec.Stdout = cwl.DOCKER_LOG_DIR + "/STDOUT"
	}
	if cmd.Stderr != "" {
		exec.Stderr = filepath.Join(cwl.DOCKER_WORK_DIR, cmd.Stderr)
	} else {
		exec.Stderr = cwl.DOCKER_LOG_DIR + "/STDERR"
	}
	out.Executors = []*tes.Executor{&exec}

	for _, i := range cwl.GetFileInputs(env.Inputs) {
		input := tes.Input{
			Url:  i.StoragePath,
			Path: i.MappedPath,
			Type: tes.FileType_FILE,
		}
		log.Printf("Found Input: %s", i)
		out.Inputs = append(out.Inputs, &input)
	}
	out.Volumes = []string{cwl.DOCKER_WORK_DIR}
	output := tes.Output{
		Url:  mapper.StoragePath("output"),
		Path: cwl.DOCKER_WORK_DIR,
		Type: tes.FileType_DIRECTORY,
	}
	out.Outputs = append(out.Outputs, &output)

	post := PostProcess{Steps: []*PostProcessStep{}}
	for _, i := range cmd.Outputs {
		g := OutputGlob{ParamName: i.Id, Glob: i.OutputBinding.Glob}
		post.Steps = append(post.Steps, &PostProcessStep{Step: &PostProcessStep_GlobOutput{&g}})

	}
	return out, post, nil
}
