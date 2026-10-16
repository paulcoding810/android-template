package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

const (
	svgNamespace = "http://www.w3.org/2000/svg"
	// Adaptive icons are 108dp, of which launchers show the central 72dp.
	adaptiveVisible = 72.0 / 108.0
	defaultScale    = 66.0 / 108.0
)

var densities = []struct {
	name  string
	scale float64
}{
	{"mdpi", 1}, {"hdpi", 1.5}, {"xhdpi", 2}, {"xxhdpi", 3}, {"xxxhdpi", 4},
}

// SVG elements that would be silently dropped or drawn wrong by the renderer.
var unsupportedElements = map[string]bool{
	"text": true, "tspan": true, "textPath": true, "image": true, "mask": true,
	"clipPath": true, "filter": true, "pattern": true, "marker": true,
	"foreignObject": true, "symbol": true, "switch": true,
}

var hexColorRe = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

type iconOptions struct {
	svg   []byte
	bg    color.NRGBA
	scale float64
}

func runIcon(args []string, stdout, stderr io.Writer) error {
	fset := flag.NewFlagSet("gen-android icon", flag.ContinueOnError)
	fset.SetOutput(stderr)
	dir := fset.String("dir", ".", "project directory")
	bg := fset.String("bg", "#FFFFFF", "background color (#RGB or #RRGGBB)")
	scale := fset.Float64("scale", defaultScale, "fraction of the 108dp adaptive canvas the artwork fills")
	fset.Usage = func() {
		fmt.Fprintf(stderr, "Usage: gen-android icon <icon.svg> [flags]\n\nFlags:\n")
		fset.PrintDefaults()
	}

	positional, err := parseInterspersed(fset, args)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return errReported
	}
	if len(positional) != 1 {
		fset.Usage()
		return fmt.Errorf("expected 1 argument, got %d", len(positional))
	}

	opts, err := loadIconOptions(positional[0], *bg, *scale)
	if err != nil {
		return err
	}
	if err := applyIcon(*dir, opts); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Launcher icon updated in %s\n", *dir)
	return nil
}

func loadIconOptions(svgPath, bg string, scale float64) (iconOptions, error) {
	data, err := os.ReadFile(svgPath)
	if err != nil {
		return iconOptions{}, err
	}
	c, err := parseHexColor(bg)
	if err != nil {
		return iconOptions{}, err
	}
	if scale <= 0 || scale > 1 {
		return iconOptions{}, fmt.Errorf("invalid scale %v: must be in (0, 1]", scale)
	}
	opts := iconOptions{svg: data, bg: c, scale: scale}
	if _, err := opts.parse(); err != nil {
		return iconOptions{}, fmt.Errorf("%s: %w", svgPath, err)
	}
	return opts, nil
}

func parseHexColor(s string) (color.NRGBA, error) {
	if !hexColorRe.MatchString(s) {
		return color.NRGBA{}, fmt.Errorf("invalid color %q: expected #RGB or #RRGGBB", s)
	}
	hex := s[1:]
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	v, _ := strconv.ParseUint(hex, 16, 32)
	return color.NRGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 255}, nil
}

func (o iconOptions) hexColor() string {
	return fmt.Sprintf("#%02X%02X%02X", o.bg.R, o.bg.G, o.bg.B)
}

// parse returns a fresh icon; rendering mutates its transform, so each render needs its own.
func (o iconOptions) parse() (*oksvg.SvgIcon, error) {
	if err := checkSupported(o.svg); err != nil {
		return nil, err
	}
	icon, err := oksvg.ReadIconStream(bytes.NewReader(o.svg), oksvg.IgnoreErrorMode)
	if err != nil {
		return nil, fmt.Errorf("parsing svg: %w", err)
	}
	if icon.ViewBox.W <= 0 || icon.ViewBox.H <= 0 {
		return nil, errors.New("svg has no viewBox or width/height")
	}
	return icon, nil
}

func checkSupported(data []byte) error {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("parsing svg: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		// Editor metadata (sodipodi, inkscape, …) lives in other namespaces and is not drawn.
		if se.Name.Space != "" && se.Name.Space != svgNamespace {
			continue
		}
		if unsupportedElements[se.Name.Local] {
			return fmt.Errorf("unsupported svg element <%s>; convert it to paths (e.g. Inkscape: Object to Path)", se.Name.Local)
		}
	}
}

// render draws the artwork centered, filling `fill` of a size×size canvas, over an optional background shape.
func (o iconOptions) render(size int, fill float64, bg bool, cornerRadius float64) (*image.NRGBA, error) {
	icon, err := o.parse()
	if err != nil {
		return nil, err
	}
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	if bg {
		for i := 0; i < len(img.Pix); i += 4 {
			img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = o.bg.R, o.bg.G, o.bg.B, 255
		}
	}

	target := float64(size) * math.Min(fill, 1)
	vb := icon.ViewBox
	s := math.Min(target/vb.W, target/vb.H)
	x := (float64(size) - vb.W*s) / 2
	y := (float64(size) - vb.H*s) / 2
	icon.Transform = rasterx.Identity.Translate(x, y).Scale(s, s).Translate(-vb.X, -vb.Y)

	scanner := rasterx.NewScannerGV(size, size, img, img.Bounds())
	icon.Draw(rasterx.NewDasher(size, size, scanner), 1)

	if cornerRadius > 0 {
		maskRoundedRect(img, cornerRadius*float64(size))
	}
	return img, nil
}

// maskRoundedRect clears pixels outside a rounded rectangle, anti-aliasing the edge.
func maskRoundedRect(img *image.NRGBA, radius float64) {
	size := float64(img.Bounds().Dx())
	half := size / 2
	for py := 0; py < img.Bounds().Dy(); py++ {
		for px := 0; px < img.Bounds().Dx(); px++ {
			dx := math.Max(math.Abs(float64(px)+0.5-half)-(half-radius), 0)
			dy := math.Max(math.Abs(float64(py)+0.5-half)-(half-radius), 0)
			coverage := math.Min(math.Max(radius-math.Hypot(dx, dy)+0.5, 0), 1)
			i := img.PixOffset(px, py)
			img.Pix[i+3] = uint8(float64(img.Pix[i+3]) * coverage)
		}
	}
}

func applyIcon(projectDir string, opts iconOptions) error {
	resDir, err := findResDir(projectDir)
	if err != nil {
		return err
	}

	// Artwork as seen on adaptive launchers, applied to icons that are shown uncropped.
	visible := opts.scale / adaptiveVisible

	if err := removeOldIcons(resDir); err != nil {
		return err
	}
	for _, d := range densities {
		dir := filepath.Join(resDir, "mipmap-"+d.name)
		legacy := int(math.Round(48 * d.scale))
		outputs := []struct {
			name         string
			size         int
			fill         float64
			bg           bool
			cornerRadius float64
		}{
			{"ic_launcher.png", legacy, visible, true, 0.1},
			{"ic_launcher_round.png", legacy, visible, true, 0.5},
			{"ic_launcher_foreground.png", int(math.Round(108 * d.scale)), opts.scale, false, 0},
		}
		for _, out := range outputs {
			img, err := opts.render(out.size, out.fill, out.bg, out.cornerRadius)
			if err != nil {
				return err
			}
			if err := writePNG(filepath.Join(dir, out.name), img); err != nil {
				return err
			}
		}
	}

	adaptive := `<?xml version="1.0" encoding="utf-8"?>
<adaptive-icon xmlns:android="http://schemas.android.com/apk/res/android">
    <background android:drawable="@color/ic_launcher_background" />
    <foreground android:drawable="@mipmap/ic_launcher_foreground" />
    <monochrome android:drawable="@mipmap/ic_launcher_foreground" />
</adaptive-icon>
`
	for _, name := range []string{"ic_launcher.xml", "ic_launcher_round.xml"} {
		if err := writeFile(filepath.Join(resDir, "mipmap-anydpi-v26", name), []byte(adaptive)); err != nil {
			return err
		}
	}
	bgXML := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<resources>
    <color name="ic_launcher_background">%s</color>
</resources>
`, opts.hexColor())
	if err := writeFile(filepath.Join(resDir, "values", "ic_launcher_background.xml"), []byte(bgXML)); err != nil {
		return err
	}

	iosIcon := filepath.Join(projectDir, "iosApp", "iosApp", "Assets.xcassets", "AppIcon.appiconset", "app-icon-1024.png")
	if _, err := os.Stat(iosIcon); err == nil {
		img, err := opts.render(1024, visible, true, 0)
		if err != nil {
			return err
		}
		if err := writePNG(iosIcon, img); err != nil {
			return err
		}
	}
	return nil
}

func findResDir(projectDir string) (string, error) {
	for _, module := range []string{"androidApp", "app"} {
		dir := filepath.Join(projectDir, module, "src", "main", "res")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, nil
		}
	}
	return "", fmt.Errorf("no app/src/main/res or androidApp/src/main/res in %q", projectDir)
}

// removeOldIcons deletes launcher assets that would clash with the generated ones.
func removeOldIcons(resDir string) error {
	entries, err := os.ReadDir(resDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() || !(strings.HasPrefix(e.Name(), "mipmap") || strings.HasPrefix(e.Name(), "drawable")) {
			continue
		}
		files, err := os.ReadDir(filepath.Join(resDir, e.Name()))
		if err != nil {
			return err
		}
		for _, f := range files {
			base := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			switch base {
			case "ic_launcher", "ic_launcher_round", "ic_launcher_foreground", "ic_launcher_background":
				if err := os.Remove(filepath.Join(resDir, e.Name(), f.Name())); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func writePNG(path string, img image.Image) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}
	return writeFile(path, buf.Bytes())
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
