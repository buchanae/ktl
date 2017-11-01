package tes

import (
	"github.com/ohsu-comp-bio/ktl/cwl"
	"path/filepath"
)

func Render(cmd cwl.CommandLineTool, mapper cwl.FileMapper, env cwl.Environment) (Task, error) {
	cmd_line, err := cmd.Render(mapper, env)
	if err != nil {
		return Task{}, err
	}

	out := Task{}
	exec := Executor{}
	exec.Cmd = cmd_line
	exec.ImageName = cmd.GetImageName()
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
	out.Executors = []*Executor{&exec}

	for _, i := range cmd.GetMappedInputs(mapper, env) {
		input := TaskParameter{
			Url:  i.StoragePath,
			Path: i.MappedPath,
			Type: FileType_FILE,
		}
		out.Inputs = append(out.Inputs, &input)
	}
	out.Volumes = []string{cwl.DOCKER_WORK_DIR}
	output := TaskParameter{
		Url:  mapper.Volume2Storage(cwl.DOCKER_WORK_DIR),
		Path: cwl.DOCKER_WORK_DIR,
		Type: FileType_DIRECTORY,
	}
	out.Outputs = append(out.Outputs, &output)

	return out, nil
}
