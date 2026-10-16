package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateTemplates(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"main", nil},
		{"single", []string{"--single"}},
		{"kmm", []string{"--kmm", "--team-id", "ABCDE12345"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := t.TempDir()
			args := append([]string{"com.example.myapp", "MyApp", "--no-git", "--output", out}, tc.args...)
			var stdout, stderr bytes.Buffer
			if err := run(args, &stdout, &stderr); err != nil {
				t.Fatalf("run: %v\n%s", err, stderr.String())
			}

			root := filepath.Join(out, "MyApp")
			files := 0
			err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				rel, _ := filepath.Rel(root, p)
				if skippedNames[d.Name()] {
					t.Errorf("skipped path copied: %s", rel)
				}
				if strings.Contains(filepath.ToSlash(rel), "paulcoding") {
					t.Errorf("path not renamed: %s", rel)
				}
				if d.IsDir() {
					return nil
				}
				files++
				data, err := os.ReadFile(p)
				if err != nil {
					return err
				}
				lower := strings.ToLower(string(data))
				if strings.Contains(lower, "paulcoding") || strings.Contains(lower, "androidtemplate") {
					t.Errorf("template identifier left in %s", rel)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if files == 0 {
				t.Fatal("no files generated")
			}

			settings, err := os.ReadFile(filepath.Join(root, "settings.gradle.kts"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(settings), `rootProject.name = "MyApp"`) {
				t.Error("rootProject.name not set to project name")
			}

			if info, err := os.Stat(filepath.Join(root, "gradlew")); err != nil || info.Mode()&0100 == 0 {
				t.Errorf("gradlew missing or not executable: %v", err)
			}

			if tc.name == "kmm" {
				cfg, err := os.ReadFile(filepath.Join(root, "iosApp", "Configuration", "Config.xcconfig"))
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(cfg), "TEAM_ID=ABCDE12345\n") {
					t.Error("team ID not written to Config.xcconfig")
				}
			}
		})
	}
}

func TestRunRejectsBadInput(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"missing args", []string{"com.example.app"}},
		{"unknown flag", []string{"com.example.app", "App", "--kmp"}},
		{"conflicting flags", []string{"com.example.app", "App", "--single", "--kmm"}},
		{"single segment package", []string{"example", "App"}},
		{"dash in package", []string{"com.my-app", "App"}},
		{"keyword in package", []string{"com.example.class", "App"}},
		{"uppercase package", []string{"com.Example.app", "App"}},
		{"space in project", []string{"com.example.app", "My App"}},
		{"team id without kmm", []string{"com.example.app", "App", "--team-id", "ABCDE12345"}},
		{"bad team id", []string{"com.example.app", "App", "--kmm", "--team-id", "abc"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := t.TempDir()
			args := append(tc.args, "--no-git", "--output", out)
			var stdout, stderr bytes.Buffer
			if err := run(args, &stdout, &stderr); err == nil {
				t.Fatal("expected error")
			}
			entries, _ := os.ReadDir(out)
			if len(entries) != 0 {
				t.Errorf("output directory not empty after failure: %v", entries)
			}
		})
	}
}

func TestRunRejectsExistingDir(t *testing.T) {
	out := t.TempDir()
	if err := os.Mkdir(filepath.Join(out, "App"), 0755); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if err := run([]string{"--no-git", "--output", out, "com.example.app", "App"}, &stdout, &stderr); err == nil {
		t.Fatal("expected error for existing directory")
	}
}

func TestList(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run([]string{"--list"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	for name := range templateDirs {
		if !strings.Contains(stdout.String(), name) {
			t.Errorf("--list output missing %q", name)
		}
	}
}
