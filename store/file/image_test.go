package file

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestImageProcessor_IsImage(t *testing.T) {
	proc := NewImageProcessor()

	tests := []struct {
		mimeType string
		want     bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/gif", true},
		{"image/webp", true},
		{"IMAGE/JPEG", true}, // Case insensitive
		{"text/plain", false},
		{"application/pdf", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			got := proc.IsImage(tt.mimeType)
			if got != tt.want {
				t.Errorf("IsImage(%q) = %v, want %v", tt.mimeType, got, tt.want)
			}
		})
	}
}

func TestImageProcessor_GetInfo(t *testing.T) {
	proc := NewImageProcessor()

	// Create test image
	img := createTestImage(100, 200, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	// Encode as JPEG
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}

	// Get info
	info, err := proc.GetInfo(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}

	// Verify
	if info.Width != 100 {
		t.Errorf("Width = %d, want 100", info.Width)
	}
	if info.Height != 200 {
		t.Errorf("Height = %d, want 200", info.Height)
	}
	if info.Format != "jpeg" {
		t.Errorf("Format = %q, want %q", info.Format, "jpeg")
	}
}

func TestImageProcessor_GetInfo_PNG(t *testing.T) {
	proc := NewImageProcessor()

	// Create test PNG
	img := createTestImage(50, 75, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		t.Fatal(err)
	}

	// Get info
	info, err := proc.GetInfo(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}

	// Verify
	if info.Width != 50 {
		t.Errorf("Width = %d, want 50", info.Width)
	}
	if info.Height != 75 {
		t.Errorf("Height = %d, want 75", info.Height)
	}
	if info.Format != "png" {
		t.Errorf("Format = %q, want %q", info.Format, "png")
	}
}

func TestImageProcessor_Resize(t *testing.T) {
	proc := NewImageProcessor()

	// Create test image 100x200
	img := createTestImage(100, 200, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	buf := new(bytes.Buffer)
	jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})

	// Resize to 50x100
	resized, err := proc.Resize(bytes.NewReader(buf.Bytes()), 50, 100)
	if err != nil {
		t.Fatalf("Resize failed: %v", err)
	}

	// Decode resized image
	result, _, err := image.Decode(resized)
	if err != nil {
		t.Fatalf("Failed to decode resized image: %v", err)
	}

	// Verify dimensions
	bounds := result.Bounds()
	if bounds.Dx() != 50 {
		t.Errorf("Resized width = %d, want 50", bounds.Dx())
	}
	if bounds.Dy() != 100 {
		t.Errorf("Resized height = %d, want 100", bounds.Dy())
	}
}

func TestImageProcessor_Thumbnail(t *testing.T) {
	proc := NewImageProcessor()

	tests := []struct {
		name        string
		srcWidth    int
		srcHeight   int
		maxWidth    int
		maxHeight   int
		wantWidth   int
		wantHeight  int
	}{
		{
			name:       "landscape image",
			srcWidth:   1920,
			srcHeight:  1080,
			maxWidth:   200,
			maxHeight:  200,
			wantWidth:  200,
			wantHeight: 112, // 1080 * 200 / 1920
		},
		{
			name:       "portrait image",
			srcWidth:   1080,
			srcHeight:  1920,
			maxWidth:   200,
			maxHeight:  200,
			wantWidth:  112, // 1080 * 200 / 1920
			wantHeight: 200,
		},
		{
			name:       "square image",
			srcWidth:   500,
			srcHeight:  500,
			maxWidth:   100,
			maxHeight:  100,
			wantWidth:  100,
			wantHeight: 100,
		},
		{
			name:       "small image",
			srcWidth:   50,
			srcHeight:  50,
			maxWidth:   200,
			maxHeight:  200,
			wantWidth:  50, // Doesn't upscale
			wantHeight: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test image
			img := createTestImage(tt.srcWidth, tt.srcHeight, color.RGBA{R: 0, G: 0, B: 255, A: 255})
			buf := new(bytes.Buffer)
			jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})

			// Generate thumbnail
			thumb, err := proc.Thumbnail(bytes.NewReader(buf.Bytes()), tt.maxWidth, tt.maxHeight)
			if err != nil {
				t.Fatalf("Thumbnail failed: %v", err)
			}

			// Decode thumbnail
			result, _, err := image.Decode(thumb)
			if err != nil {
				t.Fatalf("Failed to decode thumbnail: %v", err)
			}

			// Verify dimensions
			bounds := result.Bounds()
			gotWidth := bounds.Dx()
			gotHeight := bounds.Dy()

			// Allow small tolerance due to rounding
			if abs(gotWidth-tt.wantWidth) > 2 {
				t.Errorf("Thumbnail width = %d, want ~%d", gotWidth, tt.wantWidth)
			}
			if abs(gotHeight-tt.wantHeight) > 2 {
				t.Errorf("Thumbnail height = %d, want ~%d", gotHeight, tt.wantHeight)
			}
		})
	}
}

func TestImageProcessor_InvalidData(t *testing.T) {
	proc := NewImageProcessor()

	// Test with non-image data
	invalidData := bytes.NewReader([]byte("this is not an image"))

	_, err := proc.GetInfo(invalidData)
	if err == nil {
		t.Error("GetInfo should fail with invalid data")
	}

	_, err = proc.Resize(invalidData, 100, 100)
	if err == nil {
		t.Error("Resize should fail with invalid data")
	}

	_, err = proc.Thumbnail(invalidData, 100, 100)
	if err == nil {
		t.Error("Thumbnail should fail with invalid data")
	}
}

// Benchmark image operations
func BenchmarkImageProcessor_GetInfo(b *testing.B) {
	proc := NewImageProcessor()
	img := createTestImage(1920, 1080, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	buf := new(bytes.Buffer)
	jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	data := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proc.GetInfo(bytes.NewReader(data))
	}
}

func BenchmarkImageProcessor_Thumbnail(b *testing.B) {
	proc := NewImageProcessor()
	img := createTestImage(1920, 1080, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	buf := new(bytes.Buffer)
	jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	data := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proc.Thumbnail(bytes.NewReader(data), 200, 200)
	}
}

func BenchmarkImageProcessor_Resize(b *testing.B) {
	proc := NewImageProcessor()
	img := createTestImage(1920, 1080, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	buf := new(bytes.Buffer)
	jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	data := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proc.Resize(bytes.NewReader(data), 960, 540)
	}
}

// Helper functions

// createTestImage creates a test image with the specified dimensions and color.
func createTestImage(width, height int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

// abs returns the absolute value of x.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
