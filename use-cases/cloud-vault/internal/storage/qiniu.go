package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/qiniu/go-sdk/v7/auth"
	qstorage "github.com/qiniu/go-sdk/v7/storage"
)

// QiniuConfig holds the credentials and options for Qiniu Kodo.
type QiniuConfig struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
	Region    string
	UseHTTPS  bool
}

// QiniuStorage stores objects in Qiniu Kodo.
type QiniuStorage struct {
	mac    *auth.Credentials
	qcfg   *qstorage.Config
	config QiniuConfig
}

// NewQiniuStorage constructs a QiniuStorage from the given config.
func NewQiniuStorage(cfg QiniuConfig) (*QiniuStorage, error) {
	region, err := resolveQiniuRegion(cfg.Region)
	if err != nil {
		return nil, err
	}
	mac := auth.New(cfg.AccessKey, cfg.SecretKey)
	qcfg := &qstorage.Config{
		Region:   region,
		UseHTTPS: cfg.UseHTTPS,
	}
	return &QiniuStorage{mac: mac, qcfg: qcfg, config: cfg}, nil
}

func resolveQiniuRegion(region string) (*qstorage.Region, error) {
	switch region {
	case "z0":
		return &qstorage.ZoneHuadong, nil
	case "z1":
		return &qstorage.ZoneHuabei, nil
	case "z2":
		return &qstorage.ZoneHuanan, nil
	case "na0":
		return &qstorage.ZoneBeimei, nil
	case "as0":
		return &qstorage.ZoneXinjiapo, nil
	default:
		return nil, fmt.Errorf("unknown qiniu region: %q", region)
	}
}

func (s *QiniuStorage) uploadToken(key string) string {
	putPolicy := qstorage.PutPolicy{Scope: s.config.Bucket + ":" + key}
	return putPolicy.UploadToken(s.mac)
}

func (s *QiniuStorage) Put(ctx context.Context, key string, body io.Reader, size int64, _ string) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("qiniu put %q: read body: %w", key, err)
	}
	uploader := qstorage.NewFormUploader(s.qcfg)
	var ret qstorage.PutRet
	err = uploader.Put(ctx, &ret, s.uploadToken(key), key, bytes.NewReader(data), int64(len(data)), &qstorage.PutExtra{})
	if err != nil {
		return fmt.Errorf("qiniu put %q: upload: %w", key, err)
	}
	return nil
}

func (s *QiniuStorage) Get(_ context.Context, key string) (io.ReadCloser, error) {
	deadline := time.Now().Add(time.Hour).Unix()
	rawURL := qstorage.MakePrivateURL(s.mac, s.config.Domain, key, deadline)
	resp, err := http.Get(rawURL) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("qiniu get %q: %w", key, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("qiniu get %q: unexpected status %d", key, resp.StatusCode)
	}
	return resp.Body, nil
}

func (s *QiniuStorage) Delete(_ context.Context, key string) error {
	manager := qstorage.NewBucketManager(s.mac, s.qcfg)
	if err := manager.Delete(s.config.Bucket, key); err != nil {
		return fmt.Errorf("qiniu delete %q: %w", key, err)
	}
	return nil
}

func (s *QiniuStorage) Exists(_ context.Context, key string) (bool, error) {
	manager := qstorage.NewBucketManager(s.mac, s.qcfg)
	_, err := manager.Stat(s.config.Bucket, key)
	if err != nil {
		// Treat any stat error as "not found" for V0.1
		return false, nil
	}
	return true, nil
}
