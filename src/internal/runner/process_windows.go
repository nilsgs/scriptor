package runner

import (
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func prepare(command *exec.Cmd, path string) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".bat" || ext == ".cmd" {
		// Go's standard argv quoting follows CommandLineToArgvW, which cmd.exe does not use.
		command.SysProcAttr = &syscall.SysProcAttr{CmdLine: `"` + command.Path + `" /d /v:off /s /c ""` + path + `""`}
	}
}
