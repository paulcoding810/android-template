package generate

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type file struct {
	rel  string // slash-separated path inside the project
	data []byte
	perm fs.FileMode
}

// Files in an existing target that mean a Gradle project is already there.
var gradleMarkers = []string{"settings.gradle.kts", "settings.gradle", "build.gradle.kts", "build.gradle"}

// writer writes into a target directory and can undo exactly what it did.
type writer struct {
	root     string
	existed  bool
	created  []string          // first paths it created under root, removed on rollback
	original map[string][]byte // files it appended to, restored on rollback
}

func newWriter(root string, allowExisting bool) (*writer, error) {
	info, err := os.Stat(root)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return &writer{root: root, original: map[string][]byte{}}, nil
	case err != nil:
		return nil, err
	case !info.IsDir():
		return nil, fmt.Errorf("%q is not a directory", root)
	case !allowExisting:
		return nil, fmt.Errorf("directory %q already exists (use --here to generate into an existing directory)", root)
	}
	for _, marker := range gradleMarkers {
		if _, err := os.Stat(filepath.Join(root, marker)); err == nil {
			return nil, fmt.Errorf("%q already contains a Gradle project (%s)", root, marker)
		}
	}
	return &writer{root: root, existed: true, original: map[string][]byte{}}, nil
}

// keepExisting are root files the user may already have; theirs wins.
var keepExisting = map[string]bool{"README.md": true}

func (w *writer) writeAll(files []file) error {
	var conflicts []string
	for _, f := range files {
		if _, err := os.Stat(w.path(f.rel)); err == nil && !keepExisting[f.rel] && f.rel != ".gitignore" {
			conflicts = append(conflicts, f.rel)
		}
	}
	if len(conflicts) > 0 {
		if len(conflicts) > 5 {
			conflicts = append(conflicts[:5], "…")
		}
		return fmt.Errorf("would overwrite existing files in %q: %s", w.root, strings.Join(conflicts, ", "))
	}

	for _, f := range files {
		target := w.path(f.rel)
		existing, err := os.ReadFile(target)
		switch {
		case err == nil && keepExisting[f.rel]:
			continue
		case err == nil && f.rel == ".gitignore":
			if err := w.appendMissingLines(target, existing, f.data); err != nil {
				return err
			}
			continue
		}
		if err := w.track(target); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(target, f.data, f.perm); err != nil {
			return err
		}
	}
	return nil
}

func (w *writer) path(rel string) string {
	return filepath.Join(w.root, filepath.FromSlash(rel))
}

// track records the highest not-yet-existing path between root and target.
func (w *writer) track(target string) error {
	rel, err := filepath.Rel(w.root, target)
	if err != nil {
		return err
	}
	p := w.root
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		p = filepath.Join(p, part)
		if _, err := os.Stat(p); errors.Is(err, fs.ErrNotExist) {
			w.created = append(w.created, p)
			return nil
		}
	}
	return nil
}

func (w *writer) appendMissingLines(target string, existing, template []byte) error {
	have := map[string]bool{}
	for _, line := range strings.Split(string(existing), "\n") {
		have[strings.TrimSpace(line)] = true
	}
	var add []string
	for _, line := range strings.Split(string(template), "\n") {
		if l := strings.TrimSpace(line); l != "" && !have[l] {
			add = append(add, l)
			have[l] = true
		}
	}
	if len(add) == 0 {
		return nil
	}
	w.original[target] = existing
	out := strings.TrimRight(string(existing), "\n") + "\n\n# Added by gen-android\n" + strings.Join(add, "\n") + "\n"
	return os.WriteFile(target, []byte(out), 0644)
}

// rollback removes what this run created and restores files it appended to; it never touches other user files.
func (w *writer) rollback() {
	if !w.existed {
		os.RemoveAll(w.root)
		return
	}
	for i := len(w.created) - 1; i >= 0; i-- {
		os.RemoveAll(w.created[i])
	}
	for path, data := range w.original {
		os.WriteFile(path, data, 0644)
	}
}
