package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"scriptor/internal/discovery"
	"scriptor/internal/picker"
	"scriptor/internal/runner"
)

var version = "dev"
var commit = "none"

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use: "scriptor", Short: "Find, filter, and run repository scripts",
		DisableFlagParsing: true, SilenceUsage: true, SilenceErrors: true,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 || (len(args) == 1 && args[0] == "--version") {
				return nil
			}
			return fmt.Errorf("usage: scriptor [--version]")
		},
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) == 1 {
				_, err := fmt.Fprintln(command.OutOrStdout(), command.Version)
				return err
			}
			return run(command, args)
		},
	}
	root.Version = version + "+" + commit
	root.SetVersionTemplate("{{.Version}}\n")
	root.CompletionOptions.DisableDefaultCmd = true
	return root
}

func run(command *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root, err := discovery.Root(cwd)
	if err != nil {
		return err
	}
	scripts, err := discovery.Find(root)
	if err != nil {
		return err
	}
	if len(scripts) == 0 {
		return fmt.Errorf("no scripts found in %s", root)
	}
	input, ok := command.InOrStdin().(*os.File)
	if !ok || !term.IsTerminal(int(input.Fd())) {
		return fmt.Errorf("script selection requires an interactive terminal")
	}
	selected, err := picker.Select(scripts, input, command.OutOrStdout())
	if err != nil {
		return err
	}
	if selected == nil {
		return nil
	}
	fmt.Fprintln(command.OutOrStdout(), "Running", selected.Path)
	return runner.Run(root, *selected, input, command.OutOrStdout(), command.ErrOrStderr())
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		var exit *runner.ExitError
		if errors.As(err, &exit) {
			os.Exit(exit.Code)
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
