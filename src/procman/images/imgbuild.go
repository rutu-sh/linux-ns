package images

import (
	"fmt"
	"os"
	"os/exec"
	"procman/common"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

func parseImageSpec(img_context string) (ImageSpec, common.ImageBuildErr) {
	_logger := common.GetLogger()

	spec_file := fmt.Sprintf("%v/ImageSpec.yaml", img_context)
	_logger.Info().Msgf("parsing image spec yaml at: %v", spec_file)

	if _, err := os.Stat(spec_file); err != nil {
		_logger.Error().Msgf("file %v not found", spec_file)
		return ImageSpec{}, common.ImageBuildErr{
			Code:    500,
			Message: fmt.Sprintf("file %v not found", spec_file),
		}
	}

	data, err := os.ReadFile(spec_file)
	if err != nil {
		return ImageSpec{}, common.ImageBuildErr{
			Code:    500,
			Message: fmt.Sprintf("error reading file %v: %v", spec_file, err),
		}
	}

	parsed_spec := ImageSpec{}
	err = yaml.Unmarshal(data, &parsed_spec)
	if err != nil {
		_logger.Error().Msgf("error unmarshal: %v", err)
		return ImageSpec{}, common.ImageBuildErr{
			Code:    500,
			Message: fmt.Sprintf("error reading the yaml spec %v: %v", spec_file, err),
		}
	}

	_logger.Info().Msg("image spec yaml parsed successfully")

	return parsed_spec, common.ImageBuildErr{Code: 0, Message: "success"}
}

func runCmd(env []string, command string, args ...string) common.ImageBuildErr {
	_logger := common.GetLogger()

	_logger.Info().Msgf("executing command: %v with args: %v", command, args)

	cmd := exec.Command(command, args...)
	if len(env) > 0 {
		cmd.Env = env
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		_logger.Error().Msgf("error running command: %v", err)
		return common.ImageBuildErr{
			Code:    500,
			Message: "error running the command",
		}
	}

	_logger.Info().Msgf("command %v executed", command)

	return common.ImageBuildErr{Code: 0, Message: "success"}
}

func performCopy(img Image, step Step, img_context string) common.ImageBuildErr {
	_logger := common.GetLogger()

	_logger.Info().Msgf("copying %v to %v", step.Source, step.Destination)

	abs_source := fmt.Sprintf("%v/%v", img_context, step.Source)
	abs_dest := fmt.Sprintf("%v%v", img.ContextTempDir, step.Destination)

	_, err := os.Stat(abs_source)
	if os.IsNotExist(err) {
		return common.ImageBuildErr{Code: 500, Message: "path does not exist"}
	}

	_logger.Info().Msgf("copied %v to %v", step.Source, step.Destination)
	return runCmd([]string{}, "cp", "-r", abs_source, abs_dest)
}

func performRun(img Image, step Step) common.ImageBuildErr {
	_logger := common.GetLogger()

	_logger.Info().Msgf("running command: %v", step.Command)

	pid, _, _ := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
	if pid == 0 {
		// inside child
		childEnv := []string{"PATH=/bin:/sbin:/usr/bin:/usr/sbin"}
		if err := syscall.Chdir(img.ContextTempDir); err != nil {
			return common.ImageBuildErr{Code: 500, Message: fmt.Sprintf("error changing dir: %v", err)}
		}
		if err := syscall.Chroot(img.ContextTempDir); err != nil {
			return common.ImageBuildErr{Code: 500, Message: fmt.Sprintf("error changing root: %v", err)}
		}
		return runCmd(childEnv, step.Command[0], step.Command[1:]...)

	} else {
		// inside parent
		var ws syscall.WaitStatus
		_, err := syscall.Wait4(int(pid), &ws, 0, nil)
		if err != nil {
			return common.ImageBuildErr{Code: 500, Message: fmt.Sprintf("error waiting for child process: %v", err)}
		}
		if ws.Exited() {
			_logger.Info().Msgf("child process exited with status: %v", ws.ExitStatus())
			if ws.ExitStatus() != 0 {
				return common.ImageBuildErr{Code: 500, Message: "something happend to the child proc"}
			} else {
				_logger.Info().Msg("command executed successfully")
				return common.ImageBuildErr{Code: 0, Message: "success"}
			}
		} else if ws.Signaled() {
			_logger.Info().Msgf("child process killed by signal: %v", ws.Signal())
			return common.ImageBuildErr{Code: 500, Message: "something happend to the child proc"}
		} else {
			return common.ImageBuildErr{Code: 500, Message: "something happend to the child proc"}
		}
	}
}

func performSteps(img Image, spec ImageSpec, img_context string) common.ImageBuildErr {
	for _, step := range spec.Steps {
		if step.Type == "copy" {
			if err := performCopy(img, step, img_context); err.Code != 0 {
				return err
			}
			continue
		}
		if step.Type == "run" {
			if err := performRun(img, step); err.Code != 0 {
				return err
			}
			continue
		}
	}

	if err := os.MkdirAll(fmt.Sprintf("%v/etc/procman", img.ContextTempDir), 0755); err != nil {
		return common.ImageBuildErr{Code: 500, Message: fmt.Sprintf("error creating conf: %v", err)}
	}

	yamlData, err := yaml.Marshal(&spec.Job)
	if err != nil {
		return common.ImageBuildErr{Code: 500, Message: fmt.Sprintf("error creating conf: %v", err)}
	}

	filepath := fmt.Sprintf("%v/etc/procman/job.yaml", img.ContextTempDir)
	err = os.WriteFile(filepath, yamlData, 0644)
	if err != nil {
		return common.ImageBuildErr{Code: 500, Message: fmt.Sprintf("error creating conf: %v", err)}
	}

	return common.ImageBuildErr{Code: 0, Message: "success"}
}

func buildAlpineBase(img Image, spec ImageSpec) common.ImageBuildErr {
	_logger := common.GetLogger()

	_logger.Info().Msg("building alpine base")
	base := strings.Split(spec.Base, ":")
	// imgbase_src := base[0]
	imgbase_ver := base[1]

	arch := "x86_64"

	commands := [][]string{
		{"wget", "-q", "-O", fmt.Sprintf("%v/rootfs.tar.gz", img.ContextTempDir), fmt.Sprintf("http://dl-cdn.alpinelinux.org/alpine/v%v/releases/%v/alpine-minirootfs-%v.0-%v.tar.gz", imgbase_ver, arch, imgbase_ver, arch)},
		{"sh", "-c", fmt.Sprintf("cd %v && tar -xf rootfs.tar.gz && rm rootfs.tar.gz", img.ContextTempDir)},
		{"chmod", "755", img.ContextTempDir},
		{"find", img.ContextTempDir, "-type", "d", "-exec", "chmod", "755", "{}", ";"},
	}

	for _, cmd := range commands {
		if err := runCmd([]string{}, cmd[0], cmd[1:]...); err.Code != 0 {
			return err
		}
	}

	_logger.Info().Msg("base alpine build succeeded")
	return common.ImageBuildErr{Code: 0, Message: "success"}
}

func packageImage(img Image) common.ImageBuildErr {
	_logger := common.GetLogger()
	if err := os.MkdirAll(img.ImgPath, 0755); err != nil {
		_logger.Error().Msgf("error creating the imgpath dir: %v", err)
		return common.ImageBuildErr{Code: 500, Message: fmt.Sprintf("error creating imgpath: %v", err)}
	}
	cmd := []string{"tar", "-czf", img.ImgPath + "/img.tar.gz", img.ContextTempDir}
	return runCmd([]string{}, cmd[0], cmd[1:]...)
}

func deleteImageContext(img Image) common.ImageBuildErr {
	_logger := common.GetLogger()
	_logger.Info().Msg("deleting image context dir: " + img.ContextTempDir)
	cmd := []string{"rm", "-rf", img.ContextTempDir}
	return runCmd([]string{}, cmd[0], cmd[1:]...)
}

func writeImageMetadata(img Image) common.ImageBuildErr {
	_logger := common.GetLogger()
	_logger.Info().Msgf("writing metadata for image: %v", img)

	yamlData, err := yaml.Marshal(&img)
	if err != nil {
		return common.ImageBuildErr{Code: 500, Message: fmt.Sprintf("error writing image metadata: %v", err)}
	}

	filepath := fmt.Sprintf("%v/img.yaml", img.ImgPath)
	err = os.WriteFile(filepath, yamlData, 0644)
	if err != nil {
		return common.ImageBuildErr{Code: 500, Message: fmt.Sprintf("error creating conf: %v", err)}
	}

	_logger.Info().Msgf("successfully wrote metadata for image: %v", img)
	return common.ImageBuildErr{Code: 0, Message: "success"}
}

func buildImage(img Image, img_context string) (Image, common.ImageBuildErr) {
	_logger := common.GetLogger()
	_logger.Info().Msgf("starting image build: %v", img)

	spec, err := parseImageSpec(img_context)
	if err.Code != 0 {
		return Image{}, err
	}

	err = buildAlpineBase(img, spec)
	if err.Code != 0 {
		return Image{}, err
	}

	if err = performSteps(img, spec, img_context); err.Code != 0 {
		return Image{}, err
	}

	if err = packageImage(img); err.Code != 0 {
		return Image{}, err
	}

	if err = deleteImageContext(img); err.Code != 0 {
		return Image{}, err
	}

	img.Created = time.Now().UTC().Format("2006-01-02 15:04:05")

	if err = writeImageMetadata(img); err.Code != 0 {
		return Image{}, err
	}

	return img, common.ImageBuildErr{Code: 0, Message: "success"}

}
