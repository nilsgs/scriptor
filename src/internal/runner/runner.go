package runner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"scriptor/internal/discovery"
)

type ExitError struct{ Code int }

func (e *ExitError) Error() string { return fmt.Sprintf("script exited with code %d", e.Code) }

func interpreter(path, platform string, lookup func(string) (string, error)) (string, []string, error) {
	var candidates, args []string
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ps1":
		candidates = []string{"pwsh"}
		if platform == "windows" {
			candidates = append(candidates, "powershell.exe")
		}
		args = []string{"-NoProfile", "-File", path}
	case ".sh":
		candidates, args = []string{"sh"}, []string{path}
		if usesBash(path) {
			candidates = []string{"bash"}
		}
	case ".bash":
		candidates, args = []string{"bash"}, []string{path}
	case ".py":
		candidates, args = []string{"python3", "python"}, []string{path}
		if platform == "windows" {
			candidates = []string{"python", "py", "python3"}
		}
	case ".js", ".mjs", ".cjs":
		candidates, args = []string{"node"}, []string{path}
	case ".bat", ".cmd":
		if platform != "windows" {
			return "", nil, fmt.Errorf("Windows batch scripts require Windows")
		}
		// cmd expands percent variables even inside quotes. Reject those paths rather than execute a different command.
		if strings.ContainsAny(path, "%\r\n\"") {
			return "", nil, fmt.Errorf("batch script path contains unsupported command characters")
		}
		candidates, args = []string{"cmd.exe"}, []string{"/d", "/v:off", "/s", "/c", `""` + path + `""`}
	default:
		return "", nil, fmt.Errorf("unsupported script: %s", path)
	}
	for _, name := range candidates {
		if executable, err := lookup(name); err == nil {
			return executable, args, nil
		}
	}
	return "", nil, fmt.Errorf("missing interpreter: install %s", strings.Join(candidates, " or "))
}

func usesBash(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	header, _ := bufio.NewReader(io.LimitReader(file, 256)).ReadString('\n')
	if !strings.HasPrefix(header, "#!") {
		return false
	}
	fields := strings.Fields(strings.TrimPrefix(header, "#!"))
	if len(fields) == 0 {
		return false
	}
	if filepath.Base(fields[0]) == "bash" {
		return true
	}
	if filepath.Base(fields[0]) == "env" {
		for _, field := range fields[1:] {
			if field == "-S" {
				continue
			}
			return field == "bash"
		}
	}
	return false
}

func Run(root string, script discovery.Script, input io.Reader, output, stderr io.Writer) error {
	path := filepath.Join(root, filepath.FromSlash(script.Path))
	executable, args, err := interpreter(path, runtime.GOOS, exec.LookPath)
	if err != nil {
		return err
	}
	command := exec.Command(executable, args...)
	command.Dir = root
	command.Stdin, command.Stdout, command.Stderr = input, output, stderr
	prepare(command, path)
	if err := command.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return &ExitError{Code: max(1, exit.ExitCode())}
		}
		return fmt.Errorf("start script: %w", err)
	}
	return nil
}
