package discovery

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func write(t *testing.T, root, path, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func paths(scripts []Script) []string {
	result := []string{}
	for _, script := range scripts {
		result = append(result, script.Path)
	}
	return result
}

func TestFilesystemDiscovery(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{"b/build.PS1", "a/build.sh", "a/test.py", "a/run.mjs", "a/run.cjs", "a/run.js", "a/run.bash", "a/run.bat", "a/run.cmd", "readme.md", "node_modules/dependency.js", "dist/generated.sh", ".venv/tool.py"} {
		write(t, root, path, "")
	}
	scripts, err := Find(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a/build.sh", "a/run.bash", "a/run.bat", "a/run.cjs", "a/run.cmd", "a/run.js", "a/run.mjs", "a/test.py", "b/build.PS1"}
	if !reflect.DeepEqual(paths(scripts), want) {
		t.Fatalf("got %v, want %v", paths(scripts), want)
	}
	gotRoot, err := Root(root)
	if err != nil || gotRoot != root {
		t.Fatalf("root: %q, %v", gotRoot, err)
	}
}

func TestGitDiscovery(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	root := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		command := exec.Command("git", append([]string{"-C", root}, args...)...)
		if out, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git: %s: %v", out, err)
		}
	}
	git("init", "-q")
	write(t, root, "tracked.ps1", "")
	write(t, root, "deleted.sh", "")
	git("add", "tracked.ps1", "deleted.sh")
	write(t, root, ".gitignore", "*.ps1\nignored/\n")
	write(t, root, "ignored.ps1", "")
	write(t, root, "ignored/no.sh", "")
	write(t, root, "nested/untracked.sh", "")
	if err := os.Remove(filepath.Join(root, "deleted.sh")); err != nil {
		t.Fatal(err)
	}
	resolved, err := Root(filepath.Join(root, "nested"))
	if err != nil || resolved != root {
		t.Fatalf("root: %q, %v", resolved, err)
	}
	scripts, err := Find(resolved)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"nested/untracked.sh", "tracked.ps1"}
	if !reflect.DeepEqual(paths(scripts), want) {
		t.Fatalf("got %v, want %v", paths(scripts), want)
	}
}

func TestSymlinksAreNotFollowed(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	write(t, outside, "outside.sh", "")
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.Symlink(filepath.Join(outside, "outside.sh"), filepath.Join(root, "linked.sh")); err != nil {
		t.Fatal(err)
	}
	scripts, err := Find(root)
	if err != nil || len(scripts) != 0 {
		t.Fatalf("got %v, %v", scripts, err)
	}
}
