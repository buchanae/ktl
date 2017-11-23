package engine

import (
	"github.com/ohsu-comp-bio/ktl/cwl"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"path/filepath"
)

func Render(cmd cwl.CommandLineTool, mapper cwl.FileMapper, env cwl.Environment) (tes.Task, error) {
	cmd_line, err := cmd.Render(mapper, env)
	if err != nil {
		return tes.Task{}, err
	}

	out := tes.Task{}
	exec := tes.Executor{}
	exec.Command = cmd_line
	exec.Image = cmd.GetImageName()
	exec.Workdir = cwl.DOCKER_WORK_DIR
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

	for _, i := range cmd.GetMappedInputs(mapper, env) {
		input := tes.Input{
			Url:  i.StoragePath,
			Path: i.MappedPath,
			Type: tes.FileType_FILE,
		}
		out.Inputs = append(out.Inputs, &input)
	}
	out.Volumes = []string{cwl.DOCKER_WORK_DIR}
	output := tes.Output{
		Url:  mapper.StoragePath("output"),
		Path: cwl.DOCKER_WORK_DIR,
		Type: tes.FileType_DIRECTORY,
	}
	out.Outputs = append(out.Outputs, &output)

	return out, nil
}
