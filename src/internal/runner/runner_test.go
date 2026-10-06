package runner

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"scriptor/internal/discovery"
)

func TestInterpreterResolution(t *testing.T) {
	for _, test := range []struct{ path, platform, available, want string }{
		{"build.ps1", "windows", "powershell.exe", "powershell.exe"},
		{"build.ps1", "linux", "pwsh", "pwsh"},
		{"test.py", "windows", "py", "py"},
		{"test.py", "linux", "python3", "python3"},
		{"build.sh", "linux", "sh", "sh"},
		{"build.bash", "linux", "bash", "bash"},
		{"run.mjs", "linux", "node", "node"},
		{"build.cmd", "windows", "cmd.exe", "cmd.exe"},
	} {
		t.Run(test.path+test.platform, func(t *testing.T) {
			got, _, err := interpreter(test.path, test.platform, func(name string) (string, error) {
				if name == test.available {
					return name, nil
				}
				return "", exec.ErrNotFound
			})
			if err != nil || got != test.want {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
	for _, path := range []string{"script.ps1", "script.cmd"} {
		_, _, err := interpreter(path, "linux", func(string) (string, error) { return "", exec.ErrNotFound })
		if err == nil {
			t.Fatal("expected interpreter/platform error")
		}
	}
}

func TestExecutionWorkingDirectoryInputAndExitCode(t *testing.T) {
	root := filepath.Join(t.TempDir(), "repo with spaces & symbols")
	if err := os.MkdirAll(filepath.Join(root, "scripts"), 0755); err != nil {
		t.Fatal(err)
	}
	path, content := "scripts/test script.sh", "read answer\nprintf 'answer=%s\\n' \"$answer\"\nprintf done > marker.txt\nexit 7\n"
	if runtime.GOOS == "windows" {
		path, content = "scripts/test script.ps1", "$answer = Read-Host\nWrite-Output \"answer=$answer\"\nSet-Content marker.txt done\nexit 7\n"
	}
	if _, _, err := interpreter(filepath.Join(root, path), runtime.GOOS, exec.LookPath); err != nil {
		t.Skip(err)
	}
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(path)), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	err := Run(root, discovery.Script{Path: path}, strings.NewReader("hello\n"), &stdout, &stderr)
	var exit *ExitError
	if !errors.As(err, &exit) || exit.Code != 7 {
		t.Fatalf("exit: %v; stderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "answer=hello") {
		t.Fatalf("stdout: %s", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(root, "marker.txt")); err != nil {
		t.Fatalf("script did not run from root: %v", err)
	}
}

func TestWindowsBatchQuoting(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows batch test")
	}
	root := filepath.Join(t.TempDir(), "repo with spaces & symbols (test)!")
	if err := os.Mkdir(root, 0755); err != nil {
		t.Fatal(err)
	}
	path := "hello & world.cmd"
	if err := os.WriteFile(filepath.Join(root, path), []byte("@echo off\r\necho batch-ok\r\nexit /b 9\r\n"), 0644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	err := Run(root, discovery.Script{Path: path}, nil, &stdout, &stdout)
	var exit *ExitError
	if !errors.As(err, &exit) || exit.Code != 9 || !strings.Contains(stdout.String(), "batch-ok") {
		t.Fatalf("got %v, %s", err, stdout.String())
	}
}

func TestShellShebang(t *testing.T) {
	for _, test := range []struct{ header, want string }{
		{"#!/usr/bin/env bash\n", "bash"},
		{"#!/bin/bash -e\n", "bash"},
		{"#!/usr/bin/env -S bash -eu\n", "bash"},
		{"#!/bin/sh\n", "sh"},
		{"echo hello\n", "sh"},
	} {
		t.Run(test.header, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "script.sh")
			if err := os.WriteFile(path, []byte(test.header), 0644); err != nil {
				t.Fatal(err)
			}
			got, _, err := interpreter(path, "linux", func(name string) (string, error) { return name, nil })
			if err != nil || got != test.want {
				t.Fatalf("got %s, %v", got, err)
			}
		})
	}
}
