package file

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
)

// imageProcessor implements the ImageProcessor interface.
type imageProcessor struct{}

// NewImageProcessor creates a new image processor.
func NewImageProcessor() ImageProcessor {
	return &imageProcessor{}
}

// Resize scales an image to the specified dimensions.
func (p *imageProcessor) Resize(src io.Reader, width, height int) (io.Reader, error) {
	// Decode image
	img, format, err := image.Decode(src)
	if err != nil {
		return nil, &Error{
			Op:  "Resize",
			Err: err,
		}
	}

	// Resize using nearest-neighbor algorithm
	resized := resizeImage(img, width, height)

	// Encode back to image
	buf := new(bytes.Buffer)
	switch format {
	case "jpeg":
		err = jpeg.Encode(buf, resized, &jpeg.Options{Quality: 85})
	case "png":
		err = png.Encode(buf, resized)
	case "gif":
		err = gif.Encode(buf, resized, nil)
	default:
		err = fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return nil, err
	}

	return buf, nil
}

// Thumbnail generates a thumbnail maintaining aspect ratio.
func (p *imageProcessor) Thumbnail(src io.Reader, maxWidth, maxHeight int) (io.Reader, error) {
	// Decode image
	img, format, err := image.Decode(src)
	if err != nil {
		return nil, &Error{
			Op:  "Thumbnail",
			Err: err,
		}
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Don't upscale small images
	if width <= maxWidth && height <= maxHeight {
		// Return original image without resizing
		buf := new(bytes.Buffer)
		var err error
		switch format {
		case "jpeg":
			err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
		case "png":
			err = png.Encode(buf, img)
		case "gif":
			err = gif.Encode(buf, img, nil)
		default:
			err = fmt.Errorf("unsupported format: %s", format)
		}
		if err != nil {
			return nil, err
		}
		return buf, nil
	}

	// Calculate new dimensions maintaining aspect ratio
	var newWidth, newHeight int
	if width > height {
		newWidth = maxWidth
		newHeight = height * maxWidth / width
		if newHeight > maxHeight {
			newHeight = maxHeight
			newWidth = width * maxHeight / height
		}
	} else {
		newHeight = maxHeight
		newWidth = width * maxHeight / height
		if newWidth > maxWidth {
			newWidth = maxWidth
			newHeight = height * maxWidth / width
		}
	}

	// Resize
	resized := resizeImage(img, newWidth, newHeight)

	// Encode
	buf := new(bytes.Buffer)
	switch format {
	case "jpeg":
		err = jpeg.Encode(buf, resized, &jpeg.Options{Quality: 85})
	case "png":
		err = png.Encode(buf, resized)
	case "gif":
		err = gif.Encode(buf, resized, nil)
	default:
		err = fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return nil, err
	}

	return buf, nil
}

// GetInfo extracts image metadata.
func (p *imageProcessor) GetInfo(src io.Reader) (*ImageInfo, error) {
	config, format, err := image.DecodeConfig(src)
	if err != nil {
		return nil, &Error{
			Op:  "GetInfo",
			Err: err,
		}
	}

	return &ImageInfo{
		Width:  config.Width,
		Height: config.Height,
		Format: format,
	}, nil
}

// IsImage checks if the MIME type represents an image.
func (p *imageProcessor) IsImage(mimeType string) bool {
	mimeType = strings.ToLower(mimeType)
	return mimeType == "image/jpeg" ||
		mimeType == "image/png" ||
		mimeType == "image/gif" ||
		mimeType == "image/webp"
}

// resizeImage resizes an image using nearest-neighbor algorithm.
// This is a simple implementation suitable for thumbnails.
func resizeImage(src image.Image, width, height int) image.Image {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := x * srcW / width
			srcY := y * srcH / height
			dst.Set(x, y, src.At(srcX+srcBounds.Min.X, srcY+srcBounds.Min.Y))
		}
	}

	return dst
}
