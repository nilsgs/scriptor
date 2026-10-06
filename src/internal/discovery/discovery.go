package discovery

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Script paths are relative to the scan root and use forward slashes.
type Script struct{ Path string }

var extensions = map[string]bool{
	".ps1": true, ".sh": true, ".bash": true, ".py": true,
	".js": true, ".mjs": true, ".cjs": true, ".bat": true, ".cmd": true,
}

var skipped = map[string]bool{
	".git": true, ".hg": true, ".svn": true, "node_modules": true,
	".venv": true, "venv": true, "__pycache__": true, "dist": true, "vendor": true,
}

func Root(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", abs)
	}
	output, err := exec.Command("git", "-C", abs, "rev-parse", "--show-toplevel").Output()
	if err == nil {
		return filepath.Clean(strings.TrimRight(string(output), "\r\n")), nil
	}
	return abs, nil
}

func Find(root string) ([]Script, error) {
	var paths []string
	// Git owns ignore semantics, including tracked scripts which happen to match .gitignore.
	if err := exec.Command("git", "-C", root, "rev-parse", "--show-toplevel").Run(); err == nil {
		output, err := exec.Command("git", "-C", root, "ls-files", "--cached", "--others", "--exclude-standard", "-z").Output()
		if err != nil {
			return nil, fmt.Errorf("list repository files: %w", err)
		}
		for _, path := range bytes.Split(output, []byte{0}) {
			if len(path) > 0 {
				paths = append(paths, string(path))
			}
		}
	} else {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path == root {
				return nil
			}
			if entry.IsDir() && skipped[entry.Name()] {
				return filepath.SkipDir
			}
			if entry.Type().IsRegular() {
				rel, err := filepath.Rel(root, path)
				if err != nil {
					return err
				}
				paths = append(paths, filepath.ToSlash(rel))
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("scan scripts: %w", err)
		}
	}
	result := []Script{}
	seen := map[string]bool{}
	for _, path := range paths {
		if !extensions[strings.ToLower(filepath.Ext(path))] || seen[path] {
			continue
		}
		parts := strings.Split(path, "/")
		excluded := false
		for _, part := range parts[:len(parts)-1] {
			if skipped[part] {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}
		// Reject symlinks (including ancestor links) and deleted tracked files.
		full := root
		valid := true
		for i, part := range parts {
			full = filepath.Join(full, part)
			info, err := os.Lstat(full)
			if os.IsNotExist(err) {
				valid = false
				break
			}
			if err != nil {
				return nil, err
			}
			if info.Mode()&os.ModeSymlink != 0 || (i == len(parts)-1 && !info.Mode().IsRegular()) {
				valid = false
				break
			}
		}
		if valid {
			result = append(result, Script{Path: path})
			seen[path] = true
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, nil
}
