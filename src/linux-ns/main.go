package main

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func get_logger() zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	return log.Logger
}

func get_ns_env() []string {
	env := []string{}
	env = append(env, "PATH=/bin:/sbin:/usr/bin:/usr/sbin")
	env = append(env, "PS1=[namespace] # ")
	env = append(env, "HOME=/home")
	// env = append(env, "TERM=xterm-256color")
	// env = append(env, "LANG=en_US.UTF-8")
	// env = append(env, "LANGUAGE=en_US:en")
	// env = append(env, "LC_ALL=en_US.UTF-8")
	return env
}

func configure_ns_root(ns_root string) (error, bool) {
	_logger := get_logger()
	_logger.Info().Msg("configuring the root for new namespace")

	cmd := exec.Command("/bin/bash", "../../ns-deps/create_dep.sh", "create_alpine_dep", ns_root)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		_logger.Error().Msgf("error setting up dependencies: %v", err)
		os.Exit(1)
	}

	return nil, true
}

func create_new_ns(ns_root string) (error, bool) {
	_logger := get_logger()
	_logger.Info().Msg("creating a new namespace")

	configure_ns_root(ns_root)

	_logger.Info().Msg("setting up new namespace")
	if err := syscall.Unshare(syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_FS); err != nil {
		_logger.Error().Msgf("error unsharing: %v", err)
		os.Exit(1)
	}
	_logger.Info().Msg("new namespace setup complete")

	_logger.Info().Msg("changing dir to namespace root")
	if err := os.Chdir(ns_root); err != nil {
		_logger.Error().Msgf("error changing directory: %v", err)
		os.Exit(1)
	}
	_logger.Info().Msg("in the namespace root directory")

	_logger.Info().Msg("changing namespace root")
	if err := syscall.Chroot(ns_root); err != nil {
		_logger.Error().Msgf("error changing root: %v", err)
	}
	_logger.Info().Msg("changed namespace root")

	procAttr := &syscall.ProcAttr{
		Env: get_ns_env(),
		Files: []uintptr{
			uintptr(syscall.Stdin),
			uintptr(syscall.Stdout),
			uintptr(syscall.Stderr),
		},
	}
	args := []string{"sh", "-i"}
	pid, err := syscall.ForkExec("/bin/sh", args, procAttr)
	if err != nil {
		_logger.Error().Msgf("error forking: %v", err)
		return err, false
	}
	var ws syscall.WaitStatus
	_, err = syscall.Wait4(pid, &ws, 0, nil)
	if err != nil {
		_logger.Error().Msgf("error waiting for process: %v", err)
		return err, false
	}

	_logger.Info().Msgf("forked process with pid: %v", pid)

	return nil, true

}

func main() {
	ns_root := "/users/rutu_sh/ns-roots/ns-root-4"
	create_new_ns(ns_root)
}
