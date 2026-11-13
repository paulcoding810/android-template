package cli

import (
	"bytes"
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
	if err := run(t, "--no-git", "com.example.app", "-t", "single", "App", "-o", out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "App", "app", "build.gradle.kts")); err != nil {
		t.Fatalf("single-module project not created: %v", err)
	}
}

func TestRunRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct {
		name      string
		args      []string
		usageHint bool // usage mistakes point to -h; invalid values explain themselves
	}{
		{"missing args", []string{"com.example.app"}, true},
		{"unknown flag", []string{"com.example.app", "App", "--kmp"}, true},
		{"unknown template", []string{"com.example.app", "App", "--template", "flutter"}, false},
		{"template without value", []string{"com.example.app", "App", "--template"}, true},
		{"old template flag", []string{"com.example.app", "App", "--single"}, true},
		{"invalid package", []string{"com.my-app", "App"}, false},
		{"unsupported icon", []string{"com.example.app", "App", "--icon", "missing.svg"}, false},
		{"icon without svg", []string{"icon"}, true},
		{"icon unknown flag", []string{"icon", iconSVG, "--color", "#000"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := t.TempDir()
			args := tc.args
			if args[0] != "icon" {
				args = append([]string{"--no-git", "--output", out}, args...)
			}
			var stdout, stderr bytes.Buffer
			err := Run(args, &stdout, &stderr)
			if err == nil {
				t.Fatal("expected error")
			}
			if hint := strings.Contains(err.Error(), "gen-android -h"); hint != tc.usageHint {
				t.Errorf("usage hint = %v, want %v (err: %v)", hint, tc.usageHint, err)
			}
			if stdout.Len()+stderr.Len() != 0 {
				t.Errorf("errors must not print usage; got:\n%s%s", stdout.String(), stderr.String())
			}
			if entries, _ := os.ReadDir(out); len(entries) != 0 {
				t.Errorf("output directory not empty after failure: %v", entries)
			}
		})
	}
}

func TestUnknownTemplateListsAvailable(t *testing.T) {
	err := run(t, "com.example.app", "App", "-t", "flutter", "--no-git", "--output", t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "available: main, single, empty, kmm") {
		t.Errorf("error should list available templates, got %v", err)
	}
}

func TestHelp(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want []string
	}{
		{[]string{"-h"}, []string{"--scale", "no Kotlin keywords", "gradle.properties", "-app-name", "-min-sdk", "-here", "-quiet", "bundle ID"}},
		{[]string{"icon", "-h"}, []string{"-dir", "-icon-bg", "SVG may not use text"}},
	} {
		var stdout, stderr bytes.Buffer
		if err := Run(tc.args, &stdout, &stderr); err != nil {
			t.Errorf("%v: %v", tc.args, err)
		}
		for _, want := range tc.want {
			if !strings.Contains(stdout.String(), want) {
				t.Errorf("%v help missing %q", tc.args, want)
			}
		}
	}
}

func TestQuietAndNextSteps(t *testing.T) {
	out := t.TempDir()
	var stdout, stderr bytes.Buffer
	if err := Run([]string{"com.example.app", "App", "-t", "kmm", "--no-git", "-o", out}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Next:", "cd " + filepath.Join(out, "App"), "iosApp.xcodeproj"} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("output missing %q:\n%s", want, stdout.String())
		}
	}

	stdout.Reset()
	if err := Run([]string{"com.example.app", "Quiet", "-t", "empty", "--no-git", "-o", out, "--quiet"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if stdout.Len() != 0 {
		t.Errorf("--quiet printed:\n%s", stdout.String())
	}
}

func TestIconScaleFlagOnCreate(t *testing.T) {
	out := t.TempDir()
	if err := run(t, "com.example.app", "App", "--no-git", "-o", out, "--icon", iconSVG, "--icon-scale", "0.4"); err != nil {
		t.Fatal(err)
	}
	if err := run(t, "icon", iconSVG, "--dir", filepath.Join(out, "App"), "--icon-bg", "#000000", "--icon-scale", "0.5"); err != nil {
		t.Errorf("icon subcommand should accept the create flag names: %v", err)
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
	if err := run(t, "com.example.app", "App", "--template", "kmm", "--no-git", "--output", out); err != nil {
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
