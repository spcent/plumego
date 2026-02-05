package file

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

// S3Storage implements Storage interface using S3-compatible storage.
type S3Storage struct {
	endpoint  string
	region    string
	bucket    string
	accessKey string
	secretKey string
	useSSL    bool
	pathStyle bool
	client    *http.Client
	signer    *S3Signer
	metadata  MetadataManager
	imageProc ImageProcessor
}

// NewS3Storage creates a new S3-compatible storage.
func NewS3Storage(config StorageConfig, metadata MetadataManager) (*S3Storage, error) {
	if config.S3Endpoint == "" || config.S3Bucket == "" {
		return nil, fmt.Errorf("s3: endpoint and bucket are required")
	}

	s := &S3Storage{
		endpoint:  config.S3Endpoint,
		region:    config.S3Region,
		bucket:    config.S3Bucket,
		accessKey: config.S3AccessKey,
		secretKey: config.S3SecretKey,
		useSSL:    config.S3UseSSL,
		pathStyle: config.S3PathStyle,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		metadata:  metadata,
		imageProc: NewImageProcessor(),
	}

	if s.region == "" {
		s.region = "us-east-1"
	}

	s.signer = NewS3Signer(s.accessKey, s.secretKey, s.region)

	return s, nil
}

// Put uploads a file to S3 storage.
func (s *S3Storage) Put(ctx context.Context, opts PutOptions) (*File, error) {
	// Generate file ID and path
	fileID := generateID()
	ext := path.Ext(opts.FileName)
	if ext == "" && opts.ContentType != "" {
		ext = mimeToExt(opts.ContentType)
	}

	// S3 path: tenant_id/2026/02/05/uuid.ext
	now := time.Now()
	objectKey := path.Join(
		opts.TenantID,
		fmt.Sprintf("%d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
		fileID+ext,
	)

	// Read content and calculate hash
	buf := new(bytes.Buffer)
	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(buf, hash), opts.Reader)
	if err != nil {
		return nil, err
	}

	hashString := hex.EncodeToString(hash.Sum(nil))

	// Check for deduplication
	if s.metadata != nil {
		existing, err := s.metadata.GetByHash(ctx, hashString)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	// Build request URL
	reqURL := s.buildURL(objectKey)

	// Create PUT request
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", opts.ContentType)
	req.ContentLength = size

	// Sign request
	if err := s.signer.SignRequest(req, hashString); err != nil {
		return nil, err
	}

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, &Error{
			Op:   "Put",
			Path: objectKey,
			Err:  err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, &Error{
			Op:   "Put",
			Path: objectKey,
			Err:  fmt.Errorf("s3: status %d: %s", resp.StatusCode, string(body)),
		}
	}

	// Build file metadata
	file := &File{
		ID:          fileID,
		TenantID:    opts.TenantID,
		Name:        opts.FileName,
		Path:        objectKey,
		Size:        size,
		MimeType:    opts.ContentType,
		Extension:   ext,
		Hash:        hashString,
		StorageType: "s3",
		Metadata:    opts.Metadata,
		UploadedBy:  opts.UploadedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Save metadata
	if s.metadata != nil {
		if err := s.metadata.Save(ctx, file); err != nil {
			// Rollback: attempt to delete uploaded file
			s.Delete(ctx, objectKey)
			return nil, err
		}
	}

	return file, nil
}

// Get retrieves a file from S3 storage.
func (s *S3Storage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	reqURL := s.buildURL(path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	// Sign request
	if err := s.signer.SignRequest(req, ""); err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, &Error{
			Op:   "Get",
			Path: path,
			Err:  err,
		}
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, &Error{
			Op:   "Get",
			Path: path,
			Err:  ErrNotFound,
		}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &Error{
			Op:   "Get",
			Path: path,
			Err:  fmt.Errorf("s3: status %d: %s", resp.StatusCode, string(body)),
		}
	}

	return resp.Body, nil
}

// Delete removes a file from S3 storage.
func (s *S3Storage) Delete(ctx context.Context, path string) error {
	reqURL := s.buildURL(path)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	// Sign request
	if err := s.signer.SignRequest(req, ""); err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return &Error{
			Op:   "Delete",
			Path: path,
			Err:  err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &Error{
			Op:   "Delete",
			Path: path,
			Err:  ErrNotFound,
		}
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &Error{
			Op:   "Delete",
			Path: path,
			Err:  fmt.Errorf("s3: status %d: %s", resp.StatusCode, string(body)),
		}
	}

	return nil
}

// Exists checks if a file exists in S3 storage.
func (s *S3Storage) Exists(ctx context.Context, path string) (bool, error) {
	reqURL := s.buildURL(path)

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, reqURL, nil)
	if err != nil {
		return false, err
	}

	if err := s.signer.SignRequest(req, ""); err != nil {
		return false, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return resp.StatusCode == http.StatusOK, nil
}

// Stat returns file information from S3 storage.
func (s *S3Storage) Stat(ctx context.Context, path string) (*FileStat, error) {
	reqURL := s.buildURL(path)

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, reqURL, nil)
	if err != nil {
		return nil, err
	}

	if err := s.signer.SignRequest(req, ""); err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("s3: status %d", resp.StatusCode)
	}

	return &FileStat{
		Path:        path,
		Size:        resp.ContentLength,
		ContentType: resp.Header.Get("Content-Type"),
		IsDir:       false,
	}, nil
}

// List returns a list of files in S3 storage matching the prefix.
func (s *S3Storage) List(ctx context.Context, prefix string, limit int) ([]*FileStat, error) {
	reqURL := s.buildURL("")

	// Build query parameters
	query := url.Values{}
	query.Set("list-type", "2")
	if prefix != "" {
		query.Set("prefix", prefix)
	}
	if limit > 0 {
		query.Set("max-keys", fmt.Sprintf("%d", limit))
	}
	reqURL += "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	if err := s.signer.SignRequest(req, ""); err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("s3: status %d: %s", resp.StatusCode, string(body))
	}

	// Parse XML response
	var listResult struct {
		Contents []struct {
			Key          string    `xml:"Key"`
			Size         int64     `xml:"Size"`
			LastModified time.Time `xml:"LastModified"`
		} `xml:"Contents"`
	}

	if err := xml.NewDecoder(resp.Body).Decode(&listResult); err != nil {
		return nil, err
	}

	results := make([]*FileStat, 0, len(listResult.Contents))
	for _, item := range listResult.Contents {
		results = append(results, &FileStat{
			Path:         item.Key,
			Size:         item.Size,
			ModifiedTime: item.LastModified,
			IsDir:        false,
		})
	}

	return results, nil
}

// GetURL returns a presigned URL for accessing the file.
func (s *S3Storage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	if expiry <= 0 {
		expiry = 15 * time.Minute
	}

	reqURL := s.buildURL(path)

	// Create presign request
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}

	// Generate presigned URL
	signedURL, err := s.signer.PresignRequest(req, expiry)
	if err != nil {
		return "", err
	}

	return signedURL, nil
}

// Copy copies a file within S3 storage.
func (s *S3Storage) Copy(ctx context.Context, srcPath, dstPath string) error {
	reqURL := s.buildURL(dstPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, nil)
	if err != nil {
		return err
	}

	// Set copy source header
	req.Header.Set("x-amz-copy-source", fmt.Sprintf("/%s/%s", s.bucket, srcPath))

	if err := s.signer.SignRequest(req, ""); err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("s3: status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// buildURL constructs the request URL for S3 operations.
func (s *S3Storage) buildURL(objectKey string) string {
	scheme := "https"
	if !s.useSSL {
		scheme = "http"
	}

	if s.pathStyle {
		// Path-style: http://s3.amazonaws.com/bucket/key
		return fmt.Sprintf("%s://%s/%s/%s", scheme, s.endpoint, s.bucket, objectKey)
	}

	// Virtual-hosted-style: http://bucket.s3.amazonaws.com/key
	return fmt.Sprintf("%s://%s.%s/%s", scheme, s.bucket, s.endpoint, objectKey)
}
