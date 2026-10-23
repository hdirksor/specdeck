package build_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/hdirksor/specdeck/internal/build"
)

// -- image test helpers --

func encodePNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encoding test png: %v", err)
	}
	return buf.Bytes()
}

func encodeJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("encoding test jpeg: %v", err)
	}
	return buf.Bytes()
}

func decodeImageDimensions(t *testing.T, path string) (width, height int) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("opening image: %v", err)
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		t.Fatalf("decoding image config: %v", err)
	}
	return cfg.Width, cfg.Height
}

// -- tests --

func TestCopyImage_CopiesFile(t *testing.T) {
	containersRoot := t.TempDir()
	pngData := encodePNG(t, 100, 100)
	writeFile(t, filepath.Join(containersRoot, "screens", "login.png"), string(pngData))

	c := build.Container{Path: filepath.Join("screens", "login.yml"), Image: "login.png"}

	outDir := t.TempDir()
	if err := build.CopyImage(c, containersRoot, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, filepath.Join(outDir, "login.png"))
	if got != string(pngData) {
		t.Errorf("copied image content: want %d bytes, got %d bytes", len(pngData), len(got))
	}
}

func TestCopyImage_NoopWhenImageEmpty(t *testing.T) {
	containersRoot := t.TempDir()
	c := build.Container{Path: "login.yml"}

	outDir := filepath.Join(t.TempDir(), "does-not-exist-yet")
	if err := build.CopyImage(c, containersRoot, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(outDir); !os.IsNotExist(err) {
		t.Errorf("expected outDir not to be created, got err: %v", err)
	}
}

func TestCopyImage_CreatesOutDir(t *testing.T) {
	containersRoot := t.TempDir()
	writeFile(t, filepath.Join(containersRoot, "login.svg"), "fake-svg-bytes")

	c := build.Container{Path: "login.yml", Image: "login.svg"}

	outDir := filepath.Join(t.TempDir(), "nested", "dist", "dir")
	if err := build.CopyImage(c, containersRoot, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "login.svg")); err != nil {
		t.Errorf("expected copied image to exist: %v", err)
	}
}

func TestCopyImage_ResizesLargePNG(t *testing.T) {
	containersRoot := t.TempDir()
	pngData := encodePNG(t, 2000, 1000)
	writeFile(t, filepath.Join(containersRoot, "login.png"), string(pngData))

	c := build.Container{Path: "login.yml", Image: "login.png"}

	outDir := t.TempDir()
	if err := build.CopyImage(c, containersRoot, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	width, height := decodeImageDimensions(t, filepath.Join(outDir, "login.png"))
	if width != 800 {
		t.Errorf("width: want 800, got %d", width)
	}
	if height != 400 {
		t.Errorf("height: want 400 (aspect preserved), got %d", height)
	}
}

func TestCopyImage_ResizesLargeJPEG(t *testing.T) {
	containersRoot := t.TempDir()
	jpegData := encodeJPEG(t, 1000, 2000)
	writeFile(t, filepath.Join(containersRoot, "login.jpg"), string(jpegData))

	c := build.Container{Path: "login.yml", Image: "login.jpg"}

	outDir := t.TempDir()
	if err := build.CopyImage(c, containersRoot, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	width, height := decodeImageDimensions(t, filepath.Join(outDir, "login.jpg"))
	if height != 800 {
		t.Errorf("height: want 800, got %d", height)
	}
	if width != 400 {
		t.Errorf("width: want 400 (aspect preserved), got %d", width)
	}
}

func TestCopyImage_SmallPNGUnchanged(t *testing.T) {
	containersRoot := t.TempDir()
	pngData := encodePNG(t, 100, 50)
	writeFile(t, filepath.Join(containersRoot, "login.png"), string(pngData))

	c := build.Container{Path: "login.yml", Image: "login.png"}

	outDir := t.TempDir()
	if err := build.CopyImage(c, containersRoot, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, filepath.Join(outDir, "login.png"))
	if got != string(pngData) {
		t.Errorf("expected byte-identical copy when under threshold, got different bytes (want %d bytes, got %d bytes)", len(pngData), len(got))
	}
}

func TestCopyImage_LargeGIFPassesThroughUnchanged(t *testing.T) {
	containersRoot := t.TempDir()
	gifData := "fake-gif-bytes-representing-a-large-animated-gif"
	writeFile(t, filepath.Join(containersRoot, "login.gif"), gifData)

	c := build.Container{Path: "login.yml", Image: "login.gif"}

	outDir := t.TempDir()
	if err := build.CopyImage(c, containersRoot, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, filepath.Join(outDir, "login.gif"))
	if got != gifData {
		t.Errorf("expected gif to pass through unchanged, want %q got %q", gifData, got)
	}
}

func TestCopyImage_WebpPassesThroughUnchanged(t *testing.T) {
	containersRoot := t.TempDir()
	webpData := "fake-webp-bytes"
	writeFile(t, filepath.Join(containersRoot, "login.webp"), webpData)

	c := build.Container{Path: "login.yml", Image: "login.webp"}

	outDir := t.TempDir()
	if err := build.CopyImage(c, containersRoot, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, filepath.Join(outDir, "login.webp"))
	if got != webpData {
		t.Errorf("expected webp to pass through unchanged, want %q got %q", webpData, got)
	}
}

func TestCopyImage_CorruptPNGReturnsError(t *testing.T) {
	containersRoot := t.TempDir()
	writeFile(t, filepath.Join(containersRoot, "login.png"), "not-a-real-png-file")

	c := build.Container{Path: "login.yml", Image: "login.png"}

	outDir := t.TempDir()
	if err := build.CopyImage(c, containersRoot, outDir); err == nil {
		t.Error("expected error decoding corrupt png, got nil")
	}
}
