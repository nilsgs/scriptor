//go:build !windows

package runner

import "os/exec"

func prepare(command *exec.Cmd, path string) {}
