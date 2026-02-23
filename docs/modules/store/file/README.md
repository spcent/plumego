# File Storage Module

> **Package**: `github.com/spcent/plumego/store/file` | **Backends**: Local, S3-compatible

## Overview

The `file` package provides unified file storage with local filesystem and S3-compatible backends. A single `Storage` interface enables seamless switching between development (local) and production (S3/MinIO) environments.

## Storage Interface

```go
type Storage interface {
    Store(ctx context.Context, key string, reader io.Reader, opts ...StoreOption) error
    Get(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    Metadata(ctx context.Context, key string) (*FileMetadata, error)
    List(ctx context.Context, prefix string) ([]string, error)
    URL(ctx context.Context, key string, expires time.Duration) (string, error)
}
```

## Local Storage

```go
import "github.com/spcent/plumego/store/file"

// Store files on local filesystem
fs := file.NewLocal("./uploads",
    file.WithMaxFileSize(10*1024*1024), // 10MB limit
    file.WithAllowedTypes([]string{"image/jpeg", "image/png", "application/pdf"}),
)

// Upload
f, _ := os.Open("avatar.png")
defer f.Close()
err := fs.Store(ctx, "avatars/user-123.png", f,
    file.WithContentType("image/png"),
    file.WithMetadata(map[string]string{"user_id": "123"}),
)

// Download
reader, err := fs.Get(ctx, "avatars/user-123.png")
defer reader.Close()
io.Copy(w, reader)

// Delete
err = fs.Delete(ctx, "avatars/user-123.png")

// Check existence
exists, err := fs.Exists(ctx, "avatars/user-123.png")
```

## S3-Compatible Storage

Works with AWS S3, MinIO, Cloudflare R2, DigitalOcean Spaces, and any S3-compatible service.

```go
fs := file.NewS3(file.S3Config{
    Endpoint:        os.Getenv("S3_ENDPOINT"),   // Optional: for non-AWS
    Region:          os.Getenv("S3_REGION"),
    Bucket:          os.Getenv("S3_BUCKET"),
    AccessKeyID:     os.Getenv("S3_ACCESS_KEY"),
    SecretAccessKey: os.Getenv("S3_SECRET_KEY"),
    UseSSL:          true,
    PathStyle:       false, // Use virtual-hosted style (default for AWS)
},
    file.WithKeyPrefix("uploads/"),           // Namespace all files
    file.WithMaxFileSize(50*1024*1024),       // 50MB
    file.WithAllowedTypes([]string{
        "image/jpeg", "image/png", "image/webp",
        "application/pdf",
        "video/mp4",
    }),
)
```

### AWS S3 Configuration

```go
fs := file.NewS3(file.S3Config{
    Region:          "us-east-1",
    Bucket:          "my-app-uploads",
    AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
    SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
})
```

### MinIO Configuration

```go
fs := file.NewS3(file.S3Config{
    Endpoint:        "localhost:9000",
    Region:          "us-east-1",
    Bucket:          "uploads",
    AccessKeyID:     "minioadmin",
    SecretAccessKey: "minioadmin",
    UseSSL:          false,
    PathStyle:       true, // Required for MinIO
})
```

## HTTP Upload Handler

```go
func handleFileUpload(fs file.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse multipart form (32MB max in memory)
        if err := r.ParseMultipartForm(32 << 20); err != nil {
            http.Error(w, "Failed to parse form", http.StatusBadRequest)
            return
        }

        uploadedFile, header, err := r.FormFile("file")
        if err != nil {
            http.Error(w, "File not found in request", http.StatusBadRequest)
            return
        }
        defer uploadedFile.Close()

        // Detect content type
        buf := make([]byte, 512)
        uploadedFile.Read(buf)
        contentType := http.DetectContentType(buf)
        uploadedFile.Seek(0, io.SeekStart)

        // Generate unique key
        ext := filepath.Ext(header.Filename)
        key := fmt.Sprintf("uploads/%s%s", uuid.New().String(), ext)

        // Store
        err = fs.Store(r.Context(), key, uploadedFile,
            file.WithContentType(contentType),
            file.WithMetadata(map[string]string{
                "original_name": header.Filename,
                "uploaded_by":   auth.UserIDFromContext(r.Context()),
            }),
        )
        if err != nil {
            http.Error(w, "Upload failed", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]string{
            "key": key,
            "url": fmt.Sprintf("/files/%s", key),
        })
    }
}
```

## File Download Handler

```go
func handleFileDownload(fs file.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        key := r.PathValue("key")

        // Get metadata for content-type header
        meta, err := fs.Metadata(r.Context(), key)
        if err != nil {
            http.Error(w, "File not found", http.StatusNotFound)
            return
        }

        // Set headers
        w.Header().Set("Content-Type", meta.ContentType)
        w.Header().Set("Content-Length", strconv.FormatInt(meta.Size, 10))
        w.Header().Set("Last-Modified", meta.LastModified.Format(http.TimeFormat))

        // Stream file
        reader, err := fs.Get(r.Context(), key)
        if err != nil {
            http.Error(w, "Download failed", http.StatusInternalServerError)
            return
        }
        defer reader.Close()

        io.Copy(w, reader)
    }
}
```

## Presigned URLs

Generate time-limited URLs for direct client access (bypass your server):

```go
// Generate URL valid for 1 hour
url, err := fs.URL(ctx, "uploads/document.pdf", time.Hour)
if err != nil {
    return err
}

// Return to client for direct download/upload
json.NewEncoder(w).Encode(map[string]string{
    "download_url": url,
    "expires_at":   time.Now().Add(time.Hour).Format(time.RFC3339),
})
```

## Image Processing

```go
func handleImageUpload(fs file.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        uploadedFile, header, _ := r.FormFile("image")
        defer uploadedFile.Close()

        // Resize image
        img, _, err := image.Decode(uploadedFile)
        if err != nil {
            http.Error(w, "Invalid image", http.StatusBadRequest)
            return
        }

        // Store original
        uploadedFile.Seek(0, io.SeekStart)
        key := "images/" + uuid.New().String()
        fs.Store(ctx, key+"/original.jpg", uploadedFile,
            file.WithContentType("image/jpeg"))

        // Store thumbnail
        thumbnail := resize(img, 200, 200)
        var thumbBuf bytes.Buffer
        jpeg.Encode(&thumbBuf, thumbnail, nil)
        fs.Store(ctx, key+"/thumb.jpg", &thumbBuf,
            file.WithContentType("image/jpeg"))

        json.NewEncoder(w).Encode(map[string]string{
            "original":  "/files/" + key + "/original.jpg",
            "thumbnail": "/files/" + key + "/thumb.jpg",
        })
    }
}
```

## File Metadata

```go
type FileMetadata struct {
    Key          string
    ContentType  string
    Size         int64
    LastModified time.Time
    ETag         string
    Metadata     map[string]string // Custom metadata
}

meta, err := fs.Metadata(ctx, "uploads/file.pdf")
fmt.Printf("Size: %d bytes\n", meta.Size)
fmt.Printf("Type: %s\n", meta.ContentType)
fmt.Printf("Uploaded by: %s\n", meta.Metadata["uploaded_by"])
```

## Best Practices

### 1. Validate File Types
```go
// Check content type, not just extension
contentType := http.DetectContentType(firstBytes)
allowedTypes := []string{"image/jpeg", "image/png"}
if !slices.Contains(allowedTypes, contentType) {
    return errors.New("file type not allowed")
}
```

### 2. Use Unique Keys
```go
// Never use original filenames directly (security risk)
key := "uploads/" + uuid.New().String() + filepath.Ext(originalName)
```

### 3. Limit File Sizes
```go
// Limit upload size at HTTP level
r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB
```

## Related Documentation

- [Security: Input Validation](../../security/input.md) — File type validation
- [Database](../db/README.md) — Store file metadata in DB
