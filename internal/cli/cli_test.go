package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paulcoding810/android-template/templates"
)

const iconSVG = "../icon/testdata/icon.svg"

func run(t *testing.T, args ...string) error {
	t.Helper()
	var stdout, stderr bytes.Buffer
	return Run(args, &stdout, &stderr)
}

func TestRunCreatesProject(t *testing.T) {
	out := t.TempDir()
	// Flags before, between and after positional arguments.
	if err := run(t, "--no-git", "com.example.app", "--single", "App", "--output", out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "App", "app", "build.gradle.kts")); err != nil {
		t.Fatalf("single-module project not created: %v", err)
	}
}

func TestRunRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct {
		name         string
		args         []string
		wantReported bool
	}{
		{"missing args", []string{"com.example.app"}, false},
		{"unknown flag", []string{"com.example.app", "App", "--kmp"}, true},
		{"conflicting flags", []string{"com.example.app", "App", "--single", "--kmm"}, false},
		{"invalid package", []string{"com.my-app", "App"}, false},
		{"unsupported icon", []string{"com.example.app", "App", "--icon", "missing.svg"}, false},
		{"icon without svg", []string{"icon"}, false},
		{"icon unknown flag", []string{"icon", iconSVG, "--color", "#000"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := t.TempDir()
			args := tc.args
			if args[0] != "icon" {
				args = append(args, "--no-git", "--output", out)
			}
			err := run(t, args...)
			if err == nil {
				t.Fatal("expected error")
			}
			if errors.Is(err, ErrReported) != tc.wantReported {
				t.Errorf("ErrReported = %v, want %v (err: %v)", !tc.wantReported, tc.wantReported, err)
			}
			if entries, _ := os.ReadDir(out); len(entries) != 0 {
				t.Errorf("output directory not empty after failure: %v", entries)
			}
		})
	}
}

func TestHelp(t *testing.T) {
	for _, args := range [][]string{{"-h"}, {"icon", "-h"}} {
		if err := run(t, args...); err != nil {
			t.Errorf("%v: %v", args, err)
		}
	}
}

func TestVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := Run([]string{"--version"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(stdout.String(), "gen-android ") {
		t.Errorf("unexpected version output %q", stdout.String())
	}
}

func TestList(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := Run([]string{"--list"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	for _, tmpl := range templates.All {
		if !strings.Contains(stdout.String(), tmpl.Name) {
			t.Errorf("--list output missing %q", tmpl.Name)
		}
	}
}

func TestCreateWithIcon(t *testing.T) {
	out := t.TempDir()
	if err := run(t, "com.example.app", "App", "--no-git", "--output", out, "--icon", iconSVG, "--icon-bg", "#1E88E5"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "App", "app", "src", "main", "res", "mipmap-xxxhdpi", "ic_launcher_foreground.png")); err != nil {
		t.Errorf("icon not generated: %v", err)
	}
}

func TestCreateWithBadIconLeavesNothing(t *testing.T) {
	dir := t.TempDir()
	textSVG := filepath.Join(dir, "text.svg")
	os.WriteFile(textSVG, []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><text>A</text></svg>`), 0644)
	out := t.TempDir()
	if err := run(t, "com.example.app", "App", "--no-git", "--output", out, "--icon", textSVG); err == nil {
		t.Fatal("expected error for unsupported svg")
	}
	if entries, _ := os.ReadDir(out); len(entries) != 0 {
		t.Errorf("project created despite bad icon: %v", entries)
	}
}

func TestIconSubcommand(t *testing.T) {
	out := t.TempDir()
	if err := run(t, "com.example.app", "App", "--kmm", "--no-git", "--output", out); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(out, "App")
	if err := run(t, "icon", iconSVG, "--dir", root, "--bg", "#263238", "--scale", "0.5"); err != nil {
		t.Fatal(err)
	}
	bg, err := os.ReadFile(filepath.Join(root, "androidApp", "src", "main", "res", "values", "ic_launcher_background.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bg), "#263238") {
		t.Errorf("background color not applied:\n%s", bg)
	}
}
