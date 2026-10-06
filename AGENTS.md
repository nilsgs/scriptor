# scriptor Agent Notes

## Layout

- `src/main.go`: entry point.
- `src/cmd/`: Cobra command and minimal argument handling.
- `src/internal/discovery/`: repository root and script discovery.
- `src/internal/picker/`: searchable terminal list.
- `src/internal/runner/`: interpreter selection and child processes.
- `specs/`: Smoko specs, including a pseudo-terminal test helper.
- `.smokorc`: owns the Docker test image build.

## Build And Test

Use `task test`, `task build`, `task smoke`, and `task ci`.
Use `task cross` to validate the six OS/architecture targets.
For focused debugging, run `go test ./... -v -count=1` from `src/`.

## Constraints

- The only CLI option is `--version`; no help, completion, or other subcommands.
- Filtering happens inside the picker. Never automatically execute a sole result.
- Keep the colored picker inline and compact; do not use the alternate screen.
- Search the enclosing Git repository or the current directory outside Git.
- Honor Git ignore rules and skip dependency/generated directories.
- Do not follow symlinks.
- Run scripts from the scan root with inherited standard streams.
- Restore the terminal before execution and preserve the script's exit code.
- No configuration or persistent state is required.

## Documentation And Versioning

Keep README concise. Builds stamp root `VERSION` and the short Git commit,
falling back to `unknown`. Build artifacts belong in `dist/`.
