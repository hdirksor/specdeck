package build

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

var imageExtensions = []string{".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp"}

// maxImageDimension is the longest edge, in pixels, that a raster image may
// have after build. Larger images are downscaled preserving aspect ratio.
const maxImageDimension = 800

// findImage looks in the same directory as yamlPath for a file whose stem
// matches yamlPath's stem and has a supported image extension. It returns
// "" if none is found, and an error if more than one extension matches.
func findImage(yamlPath string) (string, error) {
	dir := filepath.Dir(yamlPath)
	stem := strings.TrimSuffix(filepath.Base(yamlPath), filepath.Ext(yamlPath))

	var matches []string
	for _, ext := range imageExtensions {
		candidate := stem + ext
		if _, err := os.Stat(filepath.Join(dir, candidate)); err == nil {
			matches = append(matches, candidate)
		}
	}

	switch len(matches) {
	case 0:
		return "", nil
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("ambiguous image for %s: found %s (keep only one)", stem, strings.Join(matches, ", "))
	}
}

// CopyImage copies container c's associated image (if any) from its source
// directory under containersRoot into outDir. It is a no-op if c has no image.
func CopyImage(c Container, containersRoot, outDir string) error {
	if c.Image == "" {
		return nil
	}

	src := filepath.Join(containersRoot, filepath.Dir(c.Path), c.Image)
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading image %s: %w", c.Image, err)
	}

	data, err = resizeIfNeeded(data, filepath.Ext(c.Image))
	if err != nil {
		return fmt.Errorf("resizing image %s: %w", c.Image, err)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	dst := filepath.Join(outDir, c.Image)
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("writing image %s: %w", dst, err)
	}
	return nil
}

// resizeIfNeeded downscales PNG and JPEG image data whose longer edge
// exceeds maxImageDimension, preserving aspect ratio. Other formats (SVG,
// GIF, WebP) and images already within the limit are returned unchanged.
func resizeIfNeeded(data []byte, ext string) ([]byte, error) {
	ext = strings.ToLower(ext)
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		return data, nil
	}

	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}

	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= maxImageDimension && height <= maxImageDimension {
		return data, nil
	}

	dstWidth, dstHeight := width, height
	if width >= height {
		dstWidth = maxImageDimension
		dstHeight = height * maxImageDimension / width
	} else {
		dstHeight = maxImageDimension
		dstWidth = width * maxImageDimension / height
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	var buf bytes.Buffer
	switch ext {
	case ".png":
		err = png.Encode(&buf, dst)
	case ".jpg", ".jpeg":
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 85})
	}
	if err != nil {
		return nil, fmt.Errorf("encoding resized image: %w", err)
	}
	return buf.Bytes(), nil
}
