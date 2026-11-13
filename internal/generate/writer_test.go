package generate

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// clonedRepo mimics a freshly cloned GitHub repo.
func clonedRepo(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "acme-notes")
	write(t, filepath.Join(dir, "README.md"), "# Acme\n")
	write(t, filepath.Join(dir, ".gitignore"), "*.log\n")
	git(t, filepath.Dir(dir), "init", "-q", dir)
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "init")
	return dir
}

func TestHereIntoClonedRepo(t *testing.T) {
	repo := clonedRepo(t)
	head := git(t, repo, "rev-parse", "HEAD")
	opts := Options{Package: "com.acme.notes", Project: "AcmeNotes", Template: "single", OutDir: repo, Here: true}
	var stdout, stderr bytes.Buffer
	if err := Project(opts, &stdout, &stderr); err != nil {
		t.Fatalf("Project: %v\n%s", err, stderr.String())
	}

	if _, err := os.Stat(filepath.Join(repo, "settings.gradle.kts")); err != nil {
		t.Error("project not generated at the repo root")
	}
	if got := read(t, filepath.Join(repo, "README.md")); got != "# Acme\n" {
		t.Errorf("user's README was changed: %q", got)
	}
	gitignore := read(t, filepath.Join(repo, ".gitignore"))
	if !strings.HasPrefix(gitignore, "*.log\n") || !strings.Contains(gitignore, ".gradle") {
		t.Errorf("template .gitignore not merged into the user's:\n%s", gitignore)
	}
	if got := git(t, repo, "rev-parse", "HEAD"); got != head {
		t.Error("must not commit into the user's existing repo")
	}
	if _, err := os.Stat(filepath.Join(repo, "app", ".git")); err == nil {
		t.Error("nested repository created")
	}
	if !strings.Contains(stdout.String(), "inside an existing git repository") {
		t.Errorf("missing note about the existing repo:\n%s", stdout.String())
	}
}

func TestSubfolderInsideRepoIsNotNested(t *testing.T) {
	repo := clonedRepo(t)
	opts := Options{Package: "com.acme.notes", Project: "AcmeNotes", Template: "empty", OutDir: repo}
	var stdout, stderr bytes.Buffer
	if err := Project(opts, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(repo, "AcmeNotes", ".git")); err == nil {
		t.Error("nested repository created inside an existing repo")
	}
}

func TestHereRefusesExistingGradleProject(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "settings.gradle.kts"), "")
	opts := Options{Package: "com.acme.notes", Project: "AcmeNotes", Template: "single", OutDir: dir, Here: true, NoGit: true}
	var stdout, stderr bytes.Buffer
	err := Project(opts, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "already contains a Gradle project") {
		t.Fatalf("expected refusal, got %v", err)
	}
}

func TestHereRefusesOverwriteAndWritesNothing(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "gradlew"), "mine")
	opts := Options{Package: "com.acme.notes", Project: "AcmeNotes", Template: "single", OutDir: dir, Here: true, NoGit: true}
	var stdout, stderr bytes.Buffer
	err := Project(opts, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "would overwrite existing files") {
		t.Fatalf("expected conflict error, got %v", err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 || read(t, filepath.Join(dir, "gradlew")) != "mine" {
		t.Errorf("directory changed despite conflict: %v", entries)
	}
}

func TestExistingDirWithoutHereSuggestsIt(t *testing.T) {
	opts := newOptions(t, "single")
	os.Mkdir(filepath.Join(opts.OutDir, opts.Project), 0755)
	var stdout, stderr bytes.Buffer
	err := Project(opts, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "--here") {
		t.Fatalf("error should suggest --here, got %v", err)
	}
}

func TestRollbackKeepsUserFiles(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "docs", "notes.md"), "keep")
	write(t, filepath.Join(dir, ".gitignore"), "*.log\n")

	w, err := newWriter(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	err = w.writeAll([]file{
		{rel: "docs/generated.md", data: []byte("x"), perm: 0644},
		{rel: "app/src/Main.kt", data: []byte("x"), perm: 0644},
		{rel: ".gitignore", data: []byte(".gradle\n"), perm: 0644},
	})
	if err != nil {
		t.Fatal(err)
	}
	w.rollback()

	if read(t, filepath.Join(dir, "docs", "notes.md")) != "keep" {
		t.Error("user file removed by rollback")
	}
	for _, p := range []string{"docs/generated.md", "app"} {
		if _, err := os.Stat(filepath.Join(dir, p)); err == nil {
			t.Errorf("%s not rolled back", p)
		}
	}
	if got := read(t, filepath.Join(dir, ".gitignore")); got != "*.log\n" {
		t.Errorf(".gitignore not restored: %q", got)
	}
}
