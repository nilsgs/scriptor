# scriptor

Find scripts in a repository, type to filter, and run the one you select.

```sh
scriptor
scriptor --version
```

The colored picker appears inline in your terminal, with up to eight visible
scripts and scrolling for longer lists. Type to fuzzy-filter by repository-relative
path (case-insensitive). Use Up/Down
to select, Enter to run, and Esc or Ctrl+C to cancel. Selection is always explicit,
even with just one result. The picker closes before your script runs, so normal
prompts and terminal input work. The script's exit code becomes scriptor's exit code.
An interactive terminal is required.

## Discovery and execution

From anywhere inside a Git repository, scriptor scans the repository root,
including tracked and untracked non-ignored files. Outside Git (or without Git
installed), it scans the current directory recursively. Paths are sorted.
Symlinks and dependency/generated directories (`node_modules`, `.venv`, `venv`,
`__pycache__`, `vendor`, `dist`) and repository metadata are skipped.

| Scripts | Interpreter |
| --- | --- |
| `.ps1` | `pwsh`, or `powershell.exe` on Windows |
| `.sh` | `sh`, or `bash` for a Bash shebang |
| `.bash` | `bash` |
| `.py` | `python3` / `python`; Windows also supports `py` |
| `.js`, `.mjs`, `.cjs` | `node` |
| `.bat`, `.cmd` | `cmd.exe` (Windows only) |

Install the relevant interpreter and make it available on PATH. Scripts run from
the repository root (or current directory outside Git). PowerShell runs without
profiles; shell scripts are invoked through their listed interpreter and do not
need the executable bit. No script arguments are supplied. Batch paths containing
percent signs are rejected because cmd.exe expands them even inside quotes.

## Install

Requires Go 1.26.1 or newer to build from source.

Windows:

```powershell
.\install.ps1
```

Linux/macOS:

```sh
bash install.sh
```

These build and install to `~/.scriptor/bin` and add it to your user PATH/shell
profile. Restart your terminal afterward. With Task installed, `task install`
builds and copies the binary; add the install directory to PATH yourself.

## Development

```sh
task test
task build
task smoke
task ci
task cross
```

Smoke tests require Smoko and a running Docker daemon. They exercise the actual
interactive picker through a pseudo-terminal, including filtering, script input,
cancellation, and exit codes. Go lives under `src/`; artifacts go into `dist/`.
Builds stamp `VERSION` plus the short Git commit, or `unknown` outside Git.
