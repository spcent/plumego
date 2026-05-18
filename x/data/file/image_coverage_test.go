package file

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"
)

// ---- GetInfo edge cases ----

func TestImageProcessor_GetInfo_EmptyReader(t *testing.T) {
	proc := newImageProcessor()
	_, err := proc.GetInfo(bytes.NewReader(nil))
	if err == nil {
		t.Fatal("GetInfo with empty reader should return an error")
	}
}

func TestImageProcessor_GetInfo_GIF(t *testing.T) {
	proc := newImageProcessor()

	img := createTestGIF(60, 40, color.RGBA{R: 200, A: 255})
	buf := new(bytes.Buffer)
	if err := gif.Encode(buf, img, nil); err != nil {
		t.Fatal(err)
	}

	info, err := proc.GetInfo(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("GetInfo GIF: %v", err)
	}
	if info.Format != "gif" {
		t.Fatalf("Format = %q, want gif", info.Format)
	}
	if info.Width != 60 || info.Height != 40 {
		t.Fatalf("dimensions = %dx%d, want 60x40", info.Width, info.Height)
	}
}

// ---- Resize with PNG and GIF ----

func TestImageProcessor_Resize_PNG(t *testing.T) {
	proc := newImageProcessor()
	img := createTestImage(80, 60, color.RGBA{G: 200, A: 255})
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		t.Fatal(err)
	}

	out, err := proc.Resize(bytes.NewReader(buf.Bytes()), 40, 30)
	if err != nil {
		t.Fatalf("Resize PNG: %v", err)
	}

	result, format, err := image.Decode(out)
	if err != nil {
		t.Fatalf("decode resized PNG: %v", err)
	}
	if format != "png" {
		t.Fatalf("output format = %q, want png", format)
	}
	if result.Bounds().Dx() != 40 || result.Bounds().Dy() != 30 {
		t.Fatalf("resized dims = %dx%d, want 40x30", result.Bounds().Dx(), result.Bounds().Dy())
	}
}

func TestImageProcessor_Resize_GIF(t *testing.T) {
	proc := newImageProcessor()
	img := createTestGIF(80, 60, color.RGBA{B: 200, A: 255})
	buf := new(bytes.Buffer)
	if err := gif.Encode(buf, img, nil); err != nil {
		t.Fatal(err)
	}

	out, err := proc.Resize(bytes.NewReader(buf.Bytes()), 40, 30)
	if err != nil {
		t.Fatalf("Resize GIF: %v", err)
	}
	if out == nil {
		t.Fatal("Resize GIF returned nil reader")
	}
}

// ---- Thumbnail with PNG and GIF ----

func TestImageProcessor_Thumbnail_PNG(t *testing.T) {
	proc := newImageProcessor()
	img := createTestImage(400, 300, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		t.Fatal(err)
	}

	thumb, err := proc.Thumbnail(bytes.NewReader(buf.Bytes()), 100, 100)
	if err != nil {
		t.Fatalf("Thumbnail PNG: %v", err)
	}

	result, format, err := image.Decode(thumb)
	if err != nil {
		t.Fatalf("decode PNG thumbnail: %v", err)
	}
	if format != "png" {
		t.Fatalf("output format = %q, want png", format)
	}
	if result.Bounds().Dx() > 100 || result.Bounds().Dy() > 100 {
		t.Fatalf("thumbnail %dx%d exceeds 100x100", result.Bounds().Dx(), result.Bounds().Dy())
	}
}

func TestImageProcessor_Thumbnail_GIF(t *testing.T) {
	proc := newImageProcessor()
	img := createTestGIF(200, 200, color.RGBA{G: 200, A: 255})
	buf := new(bytes.Buffer)
	if err := gif.Encode(buf, img, nil); err != nil {
		t.Fatal(err)
	}

	thumb, err := proc.Thumbnail(bytes.NewReader(buf.Bytes()), 80, 80)
	if err != nil {
		t.Fatalf("Thumbnail GIF: %v", err)
	}
	if thumb == nil {
		t.Fatal("Thumbnail GIF returned nil")
	}
}

// ---- Thumbnail already smaller than max (no resize needed) ----

func TestImageProcessor_Thumbnail_AlreadySmall_PNG(t *testing.T) {
	proc := newImageProcessor()
	img := createTestImage(30, 20, color.RGBA{R: 50, A: 255})
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		t.Fatal(err)
	}

	thumb, err := proc.Thumbnail(bytes.NewReader(buf.Bytes()), 200, 200)
	if err != nil {
		t.Fatalf("Thumbnail (already small) PNG: %v", err)
	}

	result, _, err := image.Decode(thumb)
	if err != nil {
		t.Fatalf("decode small PNG thumbnail: %v", err)
	}
	if result.Bounds().Dx() != 30 || result.Bounds().Dy() != 20 {
		t.Fatalf("small image should pass through unchanged: got %dx%d, want 30x20",
			result.Bounds().Dx(), result.Bounds().Dy())
	}
}

// ---- Resize with invalid JPEG header ----

func TestImageProcessor_Resize_TruncatedData(t *testing.T) {
	proc := newImageProcessor()
	img := createTestImage(50, 50, color.RGBA{R: 255, A: 255})
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, nil); err != nil {
		t.Fatal(err)
	}
	truncated := buf.Bytes()[:len(buf.Bytes())/2]

	_, err := proc.Resize(bytes.NewReader(truncated), 25, 25)
	if err == nil {
		t.Fatal("Resize with truncated JPEG should return an error")
	}
}

// ---- IsImage case-folding ----

func TestImageProcessor_IsImage_MixedCase(t *testing.T) {
	proc := newImageProcessor()
	for _, mime := range []string{"Image/JPEG", "IMAGE/PNG", "image/GIF", "IMAGE/WEBP"} {
		if !proc.IsImage(mime) {
			t.Errorf("IsImage(%q) = false, want true", mime)
		}
	}
}

// ---- SupportsThumbnail does not include WebP ----

func TestImageProcessor_SupportsThumbnail_WebPUnsupported(t *testing.T) {
	proc := newImageProcessor()
	for _, mime := range []string{"image/webp", "IMAGE/WEBP", "Image/WebP"} {
		if proc.SupportsThumbnail(mime) {
			t.Errorf("SupportsThumbnail(%q) = true, want false (WebP not supported)", mime)
		}
	}
}

// ---- GetInfo truncated reader ----

func TestImageProcessor_GetInfo_TruncatedJPEG(t *testing.T) {
	proc := newImageProcessor()

	img := createTestImage(100, 100, color.RGBA{R: 255, A: 255})
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, nil); err != nil {
		t.Fatal(err)
	}

	// Only include the first 10 bytes — not enough to decode config.
	_, err := proc.GetInfo(bytes.NewReader(buf.Bytes()[:10]))
	if err == nil {
		t.Fatal("GetInfo with truncated JPEG should return an error")
	}
}

// ---- GetInfo on text data ----

func TestImageProcessor_GetInfo_TextInput(t *testing.T) {
	proc := newImageProcessor()
	_, err := proc.GetInfo(strings.NewReader("hello world"))
	if err == nil {
		t.Fatal("GetInfo with text data should return an error")
	}
}

// createTestGIF creates a paletted image suitable for GIF encoding.
func createTestGIF(width, height int, c color.RGBA) *image.Paletted {
	palette := color.Palette{color.Transparent, c}
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetColorIndex(x, y, 1)
		}
	}
	return img
}
