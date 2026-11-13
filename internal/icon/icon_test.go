package icon

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testSVG = "testdata/icon.svg"

func touch(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
}

// fakeProject lays out the launcher-icon files a template ships with.
func fakeProject(t *testing.T, module string, ios bool) (root, resDir string) {
	t.Helper()
	root = t.TempDir()
	resDir = filepath.Join(root, module, "src", "main", "res")
	for _, d := range densities {
		touch(t, filepath.Join(resDir, "mipmap-"+d.name, "ic_launcher.webp"))
		touch(t, filepath.Join(resDir, "mipmap-"+d.name, "ic_launcher_round.webp"))
	}
	touch(t, filepath.Join(resDir, "drawable", "ic_launcher_background.xml"))
	touch(t, filepath.Join(resDir, "drawable-v24", "ic_launcher_foreground.xml"))
	touch(t, filepath.Join(resDir, "values", "strings.xml"))
	if ios {
		touch(t, filepath.Join(root, "iosApp", "iosApp", "Assets.xcassets", "AppIcon.appiconset", "app-icon-1024.png"))
	}
	return root, resDir
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

func load(t *testing.T) Options {
	t.Helper()
	opts, err := Load(testSVG, "#1e88e5", DefaultScale)
	if err != nil {
		t.Fatal(err)
	}
	return opts
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
	if _, err := os.Stat(filepath.Join(resDir, "values", "strings.xml")); err != nil {
		t.Error("unrelated resources must be kept")
	}
}

func TestApplyAndroid(t *testing.T) {
	root, resDir := fakeProject(t, "app", false)
	if err := Apply(root, load(t)); err != nil {
		t.Fatal(err)
	}
	checkAndroidIcon(t, resDir)
}

func TestApplyKMM(t *testing.T) {
	root, resDir := fakeProject(t, "androidApp", true)
	if err := Apply(root, load(t)); err != nil {
		t.Fatal(err)
	}
	checkAndroidIcon(t, resDir)

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

func TestApplyRejectsNonProject(t *testing.T) {
	if err := Apply(t.TempDir(), load(t)); err == nil {
		t.Fatal("expected error outside an app project")
	}
}

func TestLoadRejectsBadInput(t *testing.T) {
	dir := t.TempDir()
	textSVG := filepath.Join(dir, "text.svg")
	os.WriteFile(textSVG, []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><text>A</text></svg>`), 0644)
	pngFile := filepath.Join(dir, "logo.png")
	os.WriteFile(pngFile, []byte("\x89PNG\r\n\x1a\nrest"), 0644)
	htmlFile := filepath.Join(dir, "page.svg")
	os.WriteFile(htmlFile, []byte(`<html><body/></html>`), 0644)
	noViewBox := filepath.Join(dir, "empty.svg")
	os.WriteFile(noViewBox, []byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect width="1" height="1"/></svg>`), 0644)

	for _, tc := range []struct {
		name  string
		path  string
		bg    string
		scale float64
	}{
		{"text element", textSVG, "#FFFFFF", DefaultScale},
		{"no viewBox", noViewBox, "#FFFFFF", DefaultScale},
		{"missing file", filepath.Join(dir, "nope.svg"), "#FFFFFF", DefaultScale},
		{"png file", pngFile, "#FFFFFF", DefaultScale},
		{"html file", htmlFile, "#FFFFFF", DefaultScale},
		{"bad color", testSVG, "blue", DefaultScale},
		{"bad scale", testSVG, "#FFFFFF", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(tc.path, tc.bg, tc.scale)
			if err == nil {
				t.Fatal("expected error")
			}
			if tc.name == "png file" && !strings.Contains(err.Error(), "must be an SVG file") {
				t.Errorf("PNG input should get a clear message, got %v", err)
			}
		})
	}
}
