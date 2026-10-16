package main

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func generateProject(t *testing.T, extra ...string) string {
	t.Helper()
	out := t.TempDir()
	args := append([]string{"com.example.myapp", "MyApp", "--no-git", "--output", out}, extra...)
	var stdout, stderr bytes.Buffer
	if err := run(args, &stdout, &stderr); err != nil {
		t.Fatalf("run: %v\n%s", err, stderr.String())
	}
	return filepath.Join(out, "MyApp")
}

func decodePNG(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return img
}

func alphaAt(img image.Image, x, y int) uint32 {
	_, _, _, a := img.At(x, y).RGBA()
	return a
}

func checkAndroidIcon(t *testing.T, resDir string) {
	t.Helper()
	for _, d := range densities {
		dir := filepath.Join(resDir, "mipmap-"+d.name)
		if _, err := os.Stat(filepath.Join(dir, "ic_launcher.webp")); err == nil {
			t.Errorf("old webp icon left in %s", dir)
		}
		legacy := int(48 * d.scale)
		for name, size := range map[string]int{
			"ic_launcher.png":            legacy,
			"ic_launcher_round.png":      legacy,
			"ic_launcher_foreground.png": int(108 * d.scale),
		} {
			img := decodePNG(t, filepath.Join(dir, name))
			if b := img.Bounds(); b.Dx() != size || b.Dy() != size {
				t.Errorf("%s/%s: got %dx%d, want %d", d.name, name, b.Dx(), b.Dy(), size)
			}
		}
	}

	fg := decodePNG(t, filepath.Join(resDir, "mipmap-xxxhdpi", "ic_launcher_foreground.png"))
	if alphaAt(fg, 0, 0) != 0 || alphaAt(fg, 216, 216) == 0 {
		t.Error("foreground should be transparent at the edge and drawn in the center")
	}
	round := decodePNG(t, filepath.Join(resDir, "mipmap-xxxhdpi", "ic_launcher_round.png"))
	if alphaAt(round, 0, 0) != 0 || alphaAt(round, 96, 96) == 0 {
		t.Error("round icon should be masked to a circle")
	}

	adaptive, err := os.ReadFile(filepath.Join(resDir, "mipmap-anydpi-v26", "ic_launcher.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(adaptive), "@mipmap/ic_launcher_foreground") || !strings.Contains(string(adaptive), "@color/ic_launcher_background") {
		t.Errorf("adaptive icon not updated:\n%s", adaptive)
	}
	bg, err := os.ReadFile(filepath.Join(resDir, "values", "ic_launcher_background.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bg), "#1E88E5") {
		t.Errorf("background color not written:\n%s", bg)
	}
	if matches, _ := filepath.Glob(filepath.Join(resDir, "drawable*", "ic_launcher_*")); len(matches) != 0 {
		t.Errorf("old drawable icons left: %v", matches)
	}
}

func TestIconOnCreate(t *testing.T) {
	root := generateProject(t, "--icon", "testdata/icon.svg", "--icon-bg", "#1e88e5")
	checkAndroidIcon(t, filepath.Join(root, "app", "src", "main", "res"))
}

func TestIconSubcommandKMM(t *testing.T) {
	root := generateProject(t, "--kmm")
	var stdout, stderr bytes.Buffer
	if err := run([]string{"icon", "testdata/icon.svg", "--dir", root, "--bg", "#1E88E5"}, &stdout, &stderr); err != nil {
		t.Fatalf("icon: %v\n%s", err, stderr.String())
	}
	checkAndroidIcon(t, filepath.Join(root, "androidApp", "src", "main", "res"))

	ios := decodePNG(t, filepath.Join(root, "iosApp", "iosApp", "Assets.xcassets", "AppIcon.appiconset", "app-icon-1024.png"))
	if b := ios.Bounds(); b.Dx() != 1024 || b.Dy() != 1024 {
		t.Errorf("iOS icon size %v", b)
	}
	if alphaAt(ios, 0, 0) != 0xffff {
		t.Error("iOS icon must be opaque")
	}
	if r, g, b, _ := ios.At(512, 512).RGBA(); r>>8 != 0xE5 || g>>8 != 0x39 || b>>8 != 0x35 {
		t.Errorf("iOS icon center should be the artwork color, got %x %x %x", r>>8, g>>8, b>>8)
	}
}

func TestIconRejectsBadInput(t *testing.T) {
	dir := t.TempDir()
	textSVG := filepath.Join(dir, "text.svg")
	os.WriteFile(textSVG, []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><text>A</text></svg>`), 0644)
	noViewBox := filepath.Join(dir, "empty.svg")
	os.WriteFile(noViewBox, []byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect width="1" height="1"/></svg>`), 0644)

	for _, tc := range []struct {
		name string
		args []string
	}{
		{"text element", []string{"icon", textSVG, "--dir", dir}},
		{"no viewBox", []string{"icon", noViewBox, "--dir", dir}},
		{"missing file", []string{"icon", filepath.Join(dir, "nope.svg"), "--dir", dir}},
		{"bad color", []string{"icon", "testdata/icon.svg", "--dir", dir, "--bg", "blue"}},
		{"bad scale", []string{"icon", "testdata/icon.svg", "--dir", dir, "--scale", "2"}},
		{"not a project", []string{"icon", "testdata/icon.svg", "--dir", dir}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := run(tc.args, &stdout, &stderr); err == nil {
				t.Fatal("expected error")
			}
		})
	}

	out := t.TempDir()
	var stdout, stderr bytes.Buffer
	if err := run([]string{"com.example.app", "App", "--no-git", "--output", out, "--icon", textSVG}, &stdout, &stderr); err == nil {
		t.Fatal("expected error for unsupported svg on create")
	}
	if entries, _ := os.ReadDir(out); len(entries) != 0 {
		t.Errorf("project created despite bad icon: %v", entries)
	}
}
