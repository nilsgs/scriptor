package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionAndMinimalInterface(t *testing.T) {
	for _, args := range [][]string{{"--version"}, {"--help"}, {"help"}, {"completion"}, {"--list"}, {"--version", "extra"}} {
		command := NewRootCmd()
		var output bytes.Buffer
		command.SetOut(&output)
		command.SetErr(&output)
		command.SetArgs(args)
		err := command.Execute()
		if args[0] == "--version" && len(args) == 1 {
			if err != nil || output.String() != version+"+"+commit+"\n" {
				t.Fatalf("version: %q, %v", output.String(), err)
			}
		} else if err == nil || !strings.Contains(err.Error(), "usage: scriptor [--version]") {
			t.Fatalf("args %v: %v", args, err)
		}
	}
}

func TestEmptyDirectory(t *testing.T) {
	t.Chdir(t.TempDir())
	command := NewRootCmd()
	command.SetArgs([]string{})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "no scripts found") {
		t.Fatalf("error: %v", err)
	}
}
