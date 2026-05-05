package file

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	storefile "github.com/spcent/plumego/store/file"
)

const s3ErrorBodyLimit = 4 << 10

// S3Storage implements Storage using an S3-compatible object store.
// Objects are organised as: {tenantID}/{YYYY}/{MM}/{DD}/{id}{ext}
type S3Storage struct {
	endpoint  string
	region    string
	bucket    string
	accessKey string
	secretKey string
	useSSL    bool
	pathStyle bool
	tempDir   string
	client    *http.Client
	signer    *S3Signer
	metadata  MetadataManager
}

// NewS3Storage creates a new S3-compatible storage.
func NewS3Storage(config S3Config, metadata MetadataManager) (*S3Storage, error) {
	if config.Endpoint == "" || config.Bucket == "" {
		return nil, fmt.Errorf("s3: endpoint and bucket are required")
	}

	s := &S3Storage{
		endpoint:  config.Endpoint,
		region:    config.Region,
		bucket:    config.Bucket,
		accessKey: config.AccessKey,
		secretKey: config.SecretKey,
		useSSL:    config.UseSSL,
		pathStyle: config.PathStyle,
		tempDir:   config.TempDir,
		client:    &http.Client{Timeout: 60 * time.Second},
		metadata:  metadata,
	}

	if s.region == "" {
		s.region = "us-east-1"
	}

	s.signer = NewS3Signer(s.accessKey, s.secretKey, s.region)
	return s, nil
}

// Put uploads a file to S3 storage under the tenant's key prefix.
func (s *S3Storage) Put(ctx context.Context, opts PutOptions) (*File, error) {
	tenantID, err := cleanTenantID(opts.TenantID)
	if err != nil {
		return nil, &storefile.Error{Op: "Put", Path: opts.TenantID, Err: err}
	}

	fileID, err := generateID()
	if err != nil {
		return nil, &storefile.Error{Op: "Put", Path: opts.TenantID, Err: err}
	}

	ext := path.Ext(opts.FileName)
	if ext == "" && opts.ContentType != "" {
		ext = mimeToExt(opts.ContentType)
	}

	now := time.Now()
	objectKey := path.Join(
		tenantID,
		fmt.Sprintf("%d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
		fileID+ext,
	)

	tmpFile, err := os.CreateTemp(s.tempDir, "plumego-s3-upload-*")
	if err != nil {
		return nil, &storefile.Error{Op: "Put", Path: objectKey, Err: err}
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	defer tmpFile.Close()

	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(tmpFile, hash), opts.Reader)
	if err != nil {
		return nil, err
	}
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return nil, &storefile.Error{Op: "Put", Path: objectKey, Err: err}
	}

	hashString := hex.EncodeToString(hash.Sum(nil))

	if s.metadata != nil {
		existing, err := s.metadata.GetByHash(ctx, tenantID, hashString)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	reqURL := s.buildURL(objectKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, tmpFile)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", opts.ContentType)
	req.ContentLength = size

	if err := s.signer.SignRequest(req, hashString); err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, &storefile.Error{Op: "Put", Path: objectKey, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body := readS3ErrorBody(resp.Body)
		return nil, &storefile.Error{
			Op:   "Put",
			Path: objectKey,
			Err:  fmt.Errorf("s3: status %d: %s", resp.StatusCode, body),
		}
	}

	file := &File{
		ID:          fileID,
		TenantID:    tenantID,
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

	if s.metadata != nil {
		if err := s.metadata.Save(ctx, file); err != nil {
			s.Delete(ctx, objectKey)
			return nil, err
		}
	}

	return file, nil
}

// Get retrieves a file from S3 storage.
func (s *S3Storage) Get(ctx context.Context, p string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.buildURL(p), nil)
	if err != nil {
		return nil, err
	}

	if err := s.signer.SignRequest(req, ""); err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, &storefile.Error{Op: "Get", Path: p, Err: err}
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, &storefile.Error{Op: "Get", Path: p, Err: storefile.ErrNotFound}
	}

	if resp.StatusCode != http.StatusOK {
		body := readS3ErrorBody(resp.Body)
		resp.Body.Close()
		return nil, &storefile.Error{
			Op:   "Get",
			Path: p,
			Err:  fmt.Errorf("s3: status %d: %s", resp.StatusCode, body),
		}
	}

	return resp.Body, nil
}

// Delete removes a file from S3 storage.
func (s *S3Storage) Delete(ctx context.Context, p string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, s.buildURL(p), nil)
	if err != nil {
		return err
	}

	if err := s.signer.SignRequest(req, ""); err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return &storefile.Error{Op: "Delete", Path: p, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &storefile.Error{Op: "Delete", Path: p, Err: storefile.ErrNotFound}
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body := readS3ErrorBody(resp.Body)
		return &storefile.Error{
			Op:   "Delete",
			Path: p,
			Err:  fmt.Errorf("s3: status %d: %s", resp.StatusCode, body),
		}
	}

	return nil
}

// Exists checks if a file exists in S3 storage.
func (s *S3Storage) Exists(ctx context.Context, p string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, s.buildURL(p), nil)
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
func (s *S3Storage) Stat(ctx context.Context, p string) (*storefile.FileStat, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, s.buildURL(p), nil)
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
		return nil, storefile.ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("s3: status %d", resp.StatusCode)
	}

	return &storefile.FileStat{
		Path:        p,
		Size:        resp.ContentLength,
		ContentType: resp.Header.Get("Content-Type"),
	}, nil
}

// List returns files in S3 storage matching the prefix.
func (s *S3Storage) List(ctx context.Context, prefix string, limit int) ([]*storefile.FileStat, error) {
	reqURL := s.buildURL("")
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
		body := readS3ErrorBody(resp.Body)
		return nil, fmt.Errorf("s3: status %d: %s", resp.StatusCode, body)
	}

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

	results := make([]*storefile.FileStat, 0, len(listResult.Contents))
	for _, item := range listResult.Contents {
		results = append(results, &storefile.FileStat{
			Path:         item.Key,
			Size:         item.Size,
			ModifiedTime: item.LastModified,
		})
	}

	return results, nil
}

// GetURL returns a presigned URL for accessing the file.
func (s *S3Storage) GetURL(ctx context.Context, p string, expiry time.Duration) (string, error) {
	if expiry <= 0 {
		expiry = 15 * time.Minute
	}

	req, err := http.NewRequest(http.MethodGet, s.buildURL(p), nil)
	if err != nil {
		return "", err
	}

	return s.signer.PresignRequest(req, expiry)
}

// Copy copies a file within S3 storage.
func (s *S3Storage) Copy(ctx context.Context, srcPath, dstPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, s.buildURL(dstPath), nil)
	if err != nil {
		return err
	}

	req.Header.Set("x-amz-copy-source", "/"+escapeObjectKey(s.bucket+"/"+strings.TrimLeft(srcPath, "/")))

	if err := s.signer.SignRequest(req, ""); err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := readS3ErrorBody(resp.Body)
		return fmt.Errorf("s3: status %d: %s", resp.StatusCode, body)
	}

	return nil
}

func (s *S3Storage) buildURL(objectKey string) string {
	objectKey = escapeObjectKey(objectKey)

	scheme := "https"
	if !s.useSSL {
		scheme = "http"
	}

	if s.pathStyle {
		return fmt.Sprintf("%s://%s/%s/%s", scheme, s.endpoint, s.bucket, objectKey)
	}

	return fmt.Sprintf("%s://%s.%s/%s", scheme, s.bucket, s.endpoint, objectKey)
}

func escapeObjectKey(objectKey string) string {
	objectKey = strings.TrimLeft(objectKey, "/")
	if objectKey == "" {
		return ""
	}

	parts := strings.Split(objectKey, "/")
	escaped := make([]string, 0, len(parts))
	for _, part := range parts {
		switch part {
		case "", ".":
			continue
		case "..":
			escaped = append(escaped, "%2E%2E")
		default:
			escaped = append(escaped, url.PathEscape(part))
		}
	}
	return strings.Join(escaped, "/")
}

func readS3ErrorBody(body io.Reader) string {
	data, _ := io.ReadAll(io.LimitReader(body, s3ErrorBodyLimit+1))
	if len(data) <= s3ErrorBodyLimit {
		return string(data)
	}
	return string(data[:s3ErrorBodyLimit]) + "...(truncated)"
}
