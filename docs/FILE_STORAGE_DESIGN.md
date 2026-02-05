# 文件存储模块详细设计

> **版本**: v1.0.0 | **日期**: 2026-02-05 | **模块路径**: `store/file/`

---

## 目录

1. [概述](#1-概述)
2. [架构设计](#2-架构设计)
3. [核心接口](#3-核心接口)
4. [本地存储实现](#4-本地存储实现)
5. [S3存储实现](#5-s3存储实现)
6. [图片处理](#6-图片处理)
7. [文件元数据管理](#7-文件元数据管理)
8. [HTTP处理器](#8-http处理器)
9. [使用示例](#9-使用示例)
10. [测试](#10-测试)

---

## 1. 概述

### 1.1 设计目标

- **标准库优先**: 不引入第三方依赖，仅使用Go标准库
- **接口分离**: 清晰的接口定义，易于扩展和测试
- **多存储支持**: 统一接口支持本地存储和S3兼容存储
- **高性能**: 流式处理、并发安全、资源复用
- **安全可靠**: 文件校验、路径安全、权限控制

### 1.2 核心能力

- 文件上传、下载、删除
- 文件去重（基于SHA256哈希）
- 图片缩略图生成
- 本地存储（文件系统）
- S3兼容存储（使用标准库HTTP客户端）
- 访问URL生成（预签名）
- 文件元数据管理

### 1.3 技术约束

- **仅使用Go标准库**: `os`, `io`, `net/http`, `crypto`, `image`
- **不引入第三方依赖**: 包括MinIO SDK、AWS SDK等
- **手动实现**: AWS Signature V4签名算法
- **标准图片格式**: 支持JPEG、PNG、GIF

---

## 2. 架构设计

### 2.1 模块结构

```
store/file/
├── file.go              # 核心接口定义
├── types.go             # 数据类型定义
├── metadata.go          # 元数据管理器
├── local.go             # 本地存储实现
├── s3.go                # S3存储实现
├── s3_signer.go         # AWS Signature V4签名
├── image.go             # 图片处理工具
├── handler.go           # HTTP处理器
├── errors.go            # 错误定义
├── utils.go             # 工具函数
├── file_test.go         # 接口测试
├── local_test.go        # 本地存储测试
├── s3_test.go           # S3存储测试
└── image_test.go        # 图片处理测试
```

### 2.2 分层架构

```
┌─────────────────────────────────────────────────┐
│              HTTP Handler Layer                  │
│  (文件上传/下载/删除/查询 HTTP接口)                │
└─────────────────────────────────────────────────┘
                       ▼
┌─────────────────────────────────────────────────┐
│            Storage Interface Layer               │
│  (统一的Storage接口，定义核心操作)                │
└─────────────────────────────────────────────────┘
                       ▼
         ┌─────────────────────────────┐
         │                             │
┌────────▼──────────┐      ┌──────────▼────────┐
│  Local Storage    │      │    S3 Storage     │
│  (本地文件系统)    │      │  (S3兼容存储)     │
└───────────────────┘      └───────────────────┘
         │                             │
         │                             │
┌────────▼──────────┐      ┌──────────▼────────┐
│   File System     │      │   HTTP Client     │
│   (os/io)         │      │   + Signature V4  │
└───────────────────┘      └───────────────────┘
```

### 2.3 依赖关系

```
handler.go
    └── file.go (Storage interface)
            ├── local.go (LocalStorage)
            │       ├── image.go
            │       └── metadata.go
            └── s3.go (S3Storage)
                    ├── s3_signer.go
                    └── metadata.go
```

---

## 3. 核心接口

### 3.1 接口定义

```go
// store/file/file.go

package file

import (
    "context"
    "io"
    "time"
)

// Storage 文件存储接口
type Storage interface {
    // Put 上传文件
    Put(ctx context.Context, opts PutOptions) (*File, error)

    // Get 获取文件（返回Reader）
    Get(ctx context.Context, path string) (io.ReadCloser, error)

    // Delete 删除文件
    Delete(ctx context.Context, path string) error

    // Exists 检查文件是否存在
    Exists(ctx context.Context, path string) (bool, error)

    // Stat 获取文件信息
    Stat(ctx context.Context, path string) (*FileStat, error)

    // List 列出文件（支持前缀过滤）
    List(ctx context.Context, prefix string, limit int) ([]*FileStat, error)

    // GetURL 获取文件访问URL
    GetURL(ctx context.Context, path string, expiry time.Duration) (string, error)

    // Copy 复制文件
    Copy(ctx context.Context, srcPath, dstPath string) error
}

// MetadataManager 元数据管理器接口
type MetadataManager interface {
    // Save 保存文件元数据
    Save(ctx context.Context, file *File) error

    // Get 获取文件元数据
    Get(ctx context.Context, id string) (*File, error)

    // GetByPath 根据路径获取元数据
    GetByPath(ctx context.Context, path string) (*File, error)

    // GetByHash 根据哈希获取元数据（去重）
    GetByHash(ctx context.Context, hash string) (*File, error)

    // List 列出文件元数据
    List(ctx context.Context, query Query) ([]*File, int64, error)

    // Delete 删除元数据（软删除）
    Delete(ctx context.Context, id string) error

    // UpdateAccessTime 更新访问时间
    UpdateAccessTime(ctx context.Context, id string) error
}

// ImageProcessor 图片处理器接口
type ImageProcessor interface {
    // Resize 调整图片大小
    Resize(src io.Reader, width, height int) (io.Reader, error)

    // Thumbnail 生成缩略图
    Thumbnail(src io.Reader, maxWidth, maxHeight int) (io.Reader, error)

    // GetInfo 获取图片信息
    GetInfo(src io.Reader) (*ImageInfo, error)

    // IsImage 判断是否为图片
    IsImage(mimeType string) bool
}
```

### 3.2 数据类型定义

```go
// store/file/types.go

package file

import (
    "time"
)

// File 文件元数据
type File struct {
    ID            string                 `json:"id" db:"id"`
    TenantID      string                 `json:"tenant_id" db:"tenant_id"`
    Name          string                 `json:"name" db:"name"`
    Path          string                 `json:"path" db:"path"`
    Size          int64                  `json:"size" db:"size"`
    MimeType      string                 `json:"mime_type" db:"mime_type"`
    Extension     string                 `json:"extension" db:"extension"`
    Hash          string                 `json:"hash" db:"hash"`
    Width         int                    `json:"width,omitempty" db:"width"`
    Height        int                    `json:"height,omitempty" db:"height"`
    ThumbnailPath string                 `json:"thumbnail_path,omitempty" db:"thumbnail_path"`
    StorageType   string                 `json:"storage_type" db:"storage_type"`
    Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
    UploadedBy    string                 `json:"uploaded_by" db:"uploaded_by"`
    CreatedAt     time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
    LastAccessAt  *time.Time             `json:"last_access_at,omitempty" db:"last_access_at"`
    DeletedAt     *time.Time             `json:"deleted_at,omitempty" db:"deleted_at"`
}

// PutOptions 上传选项
type PutOptions struct {
    TenantID      string                 // 租户ID
    Reader        io.Reader              // 文件内容
    FileName      string                 // 原始文件名
    ContentType   string                 // MIME类型
    Size          int64                  // 文件大小（可选，-1表示未知）
    UploadedBy    string                 // 上传人
    GenerateThumb bool                   // 是否生成缩略图
    ThumbWidth    int                    // 缩略图宽度（默认200）
    ThumbHeight   int                    // 缩略图高度（默认200）
    Metadata      map[string]interface{} // 附加元数据
}

// FileStat 文件状态信息
type FileStat struct {
    Path         string    // 文件路径
    Size         int64     // 文件大小
    ModifiedTime time.Time // 修改时间
    IsDir        bool      // 是否目录
    ContentType  string    // 内容类型
}

// Query 查询选项
type Query struct {
    TenantID    string
    UploadedBy  string
    MimeType    string
    StartTime   time.Time
    EndTime     time.Time
    Page        int
    PageSize    int
    OrderBy     string
}

// ImageInfo 图片信息
type ImageInfo struct {
    Width   int
    Height  int
    Format  string // jpeg, png, gif
}

// StorageConfig 存储配置
type StorageConfig struct {
    Type string // local, s3

    // Local storage config
    LocalBasePath string
    LocalBaseURL  string

    // S3 storage config
    S3Endpoint   string
    S3Region     string
    S3Bucket     string
    S3AccessKey  string
    S3SecretKey  string
    S3UseSSL     bool
    S3PathStyle  bool // 路径样式（true: s3.com/bucket/key, false: bucket.s3.com/key）
}
```

### 3.3 错误定义

```go
// store/file/errors.go

package file

import (
    "errors"
    "fmt"
)

var (
    // ErrNotFound 文件不存在
    ErrNotFound = errors.New("file: not found")

    // ErrAlreadyExists 文件已存在
    ErrAlreadyExists = errors.New("file: already exists")

    // ErrInvalidPath 无效路径
    ErrInvalidPath = errors.New("file: invalid path")

    // ErrInvalidSize 无效文件大小
    ErrInvalidSize = errors.New("file: invalid size")

    // ErrUnsupportedFormat 不支持的格式
    ErrUnsupportedFormat = errors.New("file: unsupported format")

    // ErrStorageUnavailable 存储不可用
    ErrStorageUnavailable = errors.New("file: storage unavailable")
)

// Error 文件错误
type Error struct {
    Op   string // 操作名称
    Path string // 文件路径
    Err  error  // 底层错误
}

func (e *Error) Error() string {
    if e.Path != "" {
        return fmt.Sprintf("file: %s %s: %v", e.Op, e.Path, e.Err)
    }
    return fmt.Sprintf("file: %s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
    return e.Err
}
```

---

## 4. 本地存储实现

### 4.1 本地存储实现

```go
// store/file/local.go

package file

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
    "time"
)

// LocalStorage 本地文件存储
type LocalStorage struct {
    basePath string            // 存储根目录
    baseURL  string            // 访问URL前缀
    metadata MetadataManager   // 元数据管理器
    imageProc ImageProcessor   // 图片处理器
}

// NewLocalStorage 创建本地存储
func NewLocalStorage(basePath, baseURL string, metadata MetadataManager) (*LocalStorage, error) {
    // 确保基础路径存在
    if err := os.MkdirAll(basePath, 0755); err != nil {
        return nil, &Error{
            Op:   "NewLocalStorage",
            Path: basePath,
            Err:  err,
        }
    }

    return &LocalStorage{
        basePath:  basePath,
        baseURL:   baseURL,
        metadata:  metadata,
        imageProc: NewImageProcessor(),
    }, nil
}

// Put 上传文件
func (s *LocalStorage) Put(ctx context.Context, opts PutOptions) (*File, error) {
    // 生成文件ID
    fileID := generateID()

    // 获取文件扩展名
    ext := filepath.Ext(opts.FileName)
    if ext == "" && opts.ContentType != "" {
        ext = mimeToExt(opts.ContentType)
    }

    // 生成存储路径：basePath/tenant_id/2026/02/05/uuid.ext
    now := time.Now()
    relativePath := filepath.Join(
        opts.TenantID,
        fmt.Sprintf("%d", now.Year()),
        fmt.Sprintf("%02d", now.Month()),
        fmt.Sprintf("%02d", now.Day()),
        fileID+ext,
    )

    fullPath := filepath.Join(s.basePath, relativePath)

    // 确保目录存在
    dir := filepath.Dir(fullPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return nil, &Error{
            Op:   "Put",
            Path: fullPath,
            Err:  err,
        }
    }

    // 创建临时文件
    tmpFile, err := os.CreateTemp(dir, "upload-*")
    if err != nil {
        return nil, &Error{
            Op:   "Put",
            Path: fullPath,
            Err:  err,
        }
    }
    tmpPath := tmpFile.Name()
    defer os.Remove(tmpPath)

    // 计算哈希并写入临时文件
    hash := sha256.New()
    size, err := io.Copy(io.MultiWriter(tmpFile, hash), opts.Reader)
    if err != nil {
        tmpFile.Close()
        return nil, &Error{
            Op:   "Put",
            Path: fullPath,
            Err:  err,
        }
    }
    tmpFile.Close()

    hashString := hex.EncodeToString(hash.Sum(nil))

    // 检查去重
    if s.metadata != nil {
        existing, err := s.metadata.GetByHash(ctx, hashString)
        if err == nil && existing != nil {
            // 文件已存在，直接返回
            return existing, nil
        }
    }

    // 移动到最终位置
    if err := os.Rename(tmpPath, fullPath); err != nil {
        return nil, &Error{
            Op:   "Put",
            Path: fullPath,
            Err:  err,
        }
    }

    // 构建文件元数据
    file := &File{
        ID:          fileID,
        TenantID:    opts.TenantID,
        Name:        opts.FileName,
        Path:        relativePath,
        Size:        size,
        MimeType:    opts.ContentType,
        Extension:   ext,
        Hash:        hashString,
        StorageType: "local",
        Metadata:    opts.Metadata,
        UploadedBy:  opts.UploadedBy,
        CreatedAt:   now,
        UpdatedAt:   now,
    }

    // 如果是图片，获取尺寸信息
    if opts.ContentType != "" && s.imageProc.IsImage(opts.ContentType) {
        f, err := os.Open(fullPath)
        if err == nil {
            imgInfo, err := s.imageProc.GetInfo(f)
            f.Close()
            if err == nil {
                file.Width = imgInfo.Width
                file.Height = imgInfo.Height
            }
        }

        // 生成缩略图
        if opts.GenerateThumb {
            thumbPath, err := s.generateThumbnail(fullPath, relativePath, opts.ThumbWidth, opts.ThumbHeight)
            if err == nil {
                file.ThumbnailPath = thumbPath
            }
        }
    }

    // 保存元数据
    if s.metadata != nil {
        if err := s.metadata.Save(ctx, file); err != nil {
            // 元数据保存失败，删除文件
            os.Remove(fullPath)
            return nil, err
        }
    }

    return file, nil
}

// Get 获取文件
func (s *LocalStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
    // 安全检查：防止路径穿越
    if !isPathSafe(path) {
        return nil, &Error{
            Op:   "Get",
            Path: path,
            Err:  ErrInvalidPath,
        }
    }

    fullPath := filepath.Join(s.basePath, path)

    file, err := os.Open(fullPath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, &Error{
                Op:   "Get",
                Path: path,
                Err:  ErrNotFound,
            }
        }
        return nil, &Error{
            Op:   "Get",
            Path: path,
            Err:  err,
        }
    }

    return file, nil
}

// Delete 删除文件
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
    if !isPathSafe(path) {
        return &Error{
            Op:   "Delete",
            Path: path,
            Err:  ErrInvalidPath,
        }
    }

    fullPath := filepath.Join(s.basePath, path)

    if err := os.Remove(fullPath); err != nil {
        if os.IsNotExist(err) {
            return &Error{
                Op:   "Delete",
                Path: path,
                Err:  ErrNotFound,
            }
        }
        return &Error{
            Op:   "Delete",
            Path: path,
            Err:  err,
        }
    }

    return nil
}

// Exists 检查文件是否存在
func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
    if !isPathSafe(path) {
        return false, ErrInvalidPath
    }

    fullPath := filepath.Join(s.basePath, path)
    _, err := os.Stat(fullPath)
    if err == nil {
        return true, nil
    }
    if os.IsNotExist(err) {
        return false, nil
    }
    return false, err
}

// Stat 获取文件信息
func (s *LocalStorage) Stat(ctx context.Context, path string) (*FileStat, error) {
    if !isPathSafe(path) {
        return nil, ErrInvalidPath
    }

    fullPath := filepath.Join(s.basePath, path)
    info, err := os.Stat(fullPath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, ErrNotFound
        }
        return nil, err
    }

    return &FileStat{
        Path:         path,
        Size:         info.Size(),
        ModifiedTime: info.ModTime(),
        IsDir:        info.IsDir(),
    }, nil
}

// List 列出文件
func (s *LocalStorage) List(ctx context.Context, prefix string, limit int) ([]*FileStat, error) {
    if !isPathSafe(prefix) {
        return nil, ErrInvalidPath
    }

    fullPath := filepath.Join(s.basePath, prefix)

    var results []*FileStat
    count := 0

    err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if limit > 0 && count >= limit {
            return filepath.SkipDir
        }

        if !info.IsDir() {
            relPath, err := filepath.Rel(s.basePath, path)
            if err != nil {
                return err
            }

            results = append(results, &FileStat{
                Path:         relPath,
                Size:         info.Size(),
                ModifiedTime: info.ModTime(),
                IsDir:        false,
            })
            count++
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    return results, nil
}

// GetURL 获取访问URL
func (s *LocalStorage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
    // 本地存储不支持过期URL，直接返回静态URL
    return s.baseURL + "/" + path, nil
}

// Copy 复制文件
func (s *LocalStorage) Copy(ctx context.Context, srcPath, dstPath string) error {
    if !isPathSafe(srcPath) || !isPathSafe(dstPath) {
        return ErrInvalidPath
    }

    srcFullPath := filepath.Join(s.basePath, srcPath)
    dstFullPath := filepath.Join(s.basePath, dstPath)

    // 确保目标目录存在
    if err := os.MkdirAll(filepath.Dir(dstFullPath), 0755); err != nil {
        return err
    }

    // 打开源文件
    src, err := os.Open(srcFullPath)
    if err != nil {
        return err
    }
    defer src.Close()

    // 创建目标文件
    dst, err := os.Create(dstFullPath)
    if err != nil {
        return err
    }
    defer dst.Close()

    // 复制内容
    _, err = io.Copy(dst, src)
    return err
}

// generateThumbnail 生成缩略图
func (s *LocalStorage) generateThumbnail(srcPath, relativePath string, width, height int) (string, error) {
    if width <= 0 {
        width = 200
    }
    if height <= 0 {
        height = 200
    }

    // 打开源文件
    src, err := os.Open(srcPath)
    if err != nil {
        return "", err
    }
    defer src.Close()

    // 生成缩略图
    thumbReader, err := s.imageProc.Thumbnail(src, width, height)
    if err != nil {
        return "", err
    }

    // 生成缩略图路径
    ext := filepath.Ext(relativePath)
    thumbRelPath := strings.TrimSuffix(relativePath, ext) + "_thumb" + ext
    thumbFullPath := filepath.Join(s.basePath, thumbRelPath)

    // 确保目录存在
    if err := os.MkdirAll(filepath.Dir(thumbFullPath), 0755); err != nil {
        return "", err
    }

    // 保存缩略图
    thumbFile, err := os.Create(thumbFullPath)
    if err != nil {
        return "", err
    }
    defer thumbFile.Close()

    if _, err := io.Copy(thumbFile, thumbReader); err != nil {
        os.Remove(thumbFullPath)
        return "", err
    }

    return thumbRelPath, nil
}
```

### 4.2 工具函数

```go
// store/file/utils.go

package file

import (
    "crypto/rand"
    "encoding/hex"
    "path/filepath"
    "strings"
)

// generateID 生成文件ID（32字节随机）
func generateID() string {
    b := make([]byte, 16)
    rand.Read(b)
    return hex.EncodeToString(b)
}

// isPathSafe 检查路径安全性（防止路径穿越）
func isPathSafe(path string) bool {
    // 禁止路径穿越
    if strings.Contains(path, "..") {
        return false
    }

    // 禁止绝对路径
    if filepath.IsAbs(path) {
        return false
    }

    // 清理路径
    cleaned := filepath.Clean(path)
    if cleaned != path {
        return false
    }

    return true
}

// mimeToExt 根据MIME类型获取扩展名
func mimeToExt(mimeType string) string {
    mimeMap := map[string]string{
        "image/jpeg":      ".jpg",
        "image/png":       ".png",
        "image/gif":       ".gif",
        "image/webp":      ".webp",
        "image/svg+xml":   ".svg",
        "application/pdf": ".pdf",
        "text/plain":      ".txt",
        "text/html":       ".html",
        "text/css":        ".css",
        "text/javascript": ".js",
        "application/json": ".json",
        "application/xml":  ".xml",
        "application/zip":  ".zip",
    }

    if ext, ok := mimeMap[mimeType]; ok {
        return ext
    }

    return ""
}

// extToMime 根据扩展名获取MIME类型
func extToMime(ext string) string {
    ext = strings.ToLower(ext)
    if !strings.HasPrefix(ext, ".") {
        ext = "." + ext
    }

    extMap := map[string]string{
        ".jpg":  "image/jpeg",
        ".jpeg": "image/jpeg",
        ".png":  "image/png",
        ".gif":  "image/gif",
        ".webp": "image/webp",
        ".svg":  "image/svg+xml",
        ".pdf":  "application/pdf",
        ".txt":  "text/plain",
        ".html": "text/html",
        ".css":  "text/css",
        ".js":   "text/javascript",
        ".json": "application/json",
        ".xml":  "application/xml",
        ".zip":  "application/zip",
    }

    if mime, ok := extMap[ext]; ok {
        return mime
    }

    return "application/octet-stream"
}
```

---

## 5. S3存储实现

### 5.1 S3存储实现

```go
// store/file/s3.go

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
    "strings"
    "time"
)

// S3Storage S3兼容存储
type S3Storage struct {
    endpoint    string
    region      string
    bucket      string
    accessKey   string
    secretKey   string
    useSSL      bool
    pathStyle   bool
    client      *http.Client
    signer      *S3Signer
    metadata    MetadataManager
    imageProc   ImageProcessor
}

// NewS3Storage 创建S3存储
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

// Put 上传文件
func (s *S3Storage) Put(ctx context.Context, opts PutOptions) (*File, error) {
    // 生成文件ID和路径
    fileID := generateID()
    ext := filepath.Ext(opts.FileName)
    if ext == "" && opts.ContentType != "" {
        ext = mimeToExt(opts.ContentType)
    }

    // S3路径：tenant_id/2026/02/05/uuid.ext
    now := time.Now()
    objectKey := path.Join(
        opts.TenantID,
        fmt.Sprintf("%d", now.Year()),
        fmt.Sprintf("%02d", now.Month()),
        fmt.Sprintf("%02d", now.Day()),
        fileID+ext,
    )

    // 读取内容并计算哈希
    buf := new(bytes.Buffer)
    hash := sha256.New()
    size, err := io.Copy(io.MultiWriter(buf, hash), opts.Reader)
    if err != nil {
        return nil, err
    }

    hashString := hex.EncodeToString(hash.Sum(nil))

    // 检查去重
    if s.metadata != nil {
        existing, err := s.metadata.GetByHash(ctx, hashString)
        if err == nil && existing != nil {
            return existing, nil
        }
    }

    // 构建请求URL
    reqURL := s.buildURL(objectKey)

    // 创建PUT请求
    req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewReader(buf.Bytes()))
    if err != nil {
        return nil, err
    }

    // 设置Headers
    req.Header.Set("Content-Type", opts.ContentType)
    req.ContentLength = size

    // 签名请求
    if err := s.signer.SignRequest(req, hashString); err != nil {
        return nil, err
    }

    // 发送请求
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

    // 构建文件元数据
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

    // 保存元数据
    if s.metadata != nil {
        if err := s.metadata.Save(ctx, file); err != nil {
            // 元数据保存失败，尝试删除已上传的文件
            s.Delete(ctx, objectKey)
            return nil, err
        }
    }

    return file, nil
}

// Get 获取文件
func (s *S3Storage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
    reqURL := s.buildURL(path)

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
    if err != nil {
        return nil, err
    }

    // 签名请求
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

// Delete 删除文件
func (s *S3Storage) Delete(ctx context.Context, path string) error {
    reqURL := s.buildURL(path)

    req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
    if err != nil {
        return err
    }

    // 签名请求
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

// Exists 检查文件是否存在
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

// Stat 获取文件信息
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

// List 列出文件
func (s *S3Storage) List(ctx context.Context, prefix string, limit int) ([]*FileStat, error) {
    reqURL := s.buildURL("")

    // 构建查询参数
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

    // 解析XML响应
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

// GetURL 获取预签名URL
func (s *S3Storage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
    if expiry <= 0 {
        expiry = 15 * time.Minute
    }

    reqURL := s.buildURL(path)

    // 创建预签名请求
    req, err := http.NewRequest(http.MethodGet, reqURL, nil)
    if err != nil {
        return "", err
    }

    // 生成预签名URL
    signedURL, err := s.signer.PresignRequest(req, expiry)
    if err != nil {
        return "", err
    }

    return signedURL, nil
}

// Copy 复制文件
func (s *S3Storage) Copy(ctx context.Context, srcPath, dstPath string) error {
    reqURL := s.buildURL(dstPath)

    req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, nil)
    if err != nil {
        return err
    }

    // 设置复制源
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

// buildURL 构建请求URL
func (s *S3Storage) buildURL(objectKey string) string {
    scheme := "https"
    if !s.useSSL {
        scheme = "http"
    }

    if s.pathStyle {
        // 路径样式：http://s3.amazonaws.com/bucket/key
        return fmt.Sprintf("%s://%s/%s/%s", scheme, s.endpoint, s.bucket, objectKey)
    }

    // 虚拟主机样式：http://bucket.s3.amazonaws.com/key
    return fmt.Sprintf("%s://%s.%s/%s", scheme, s.bucket, s.endpoint, objectKey)
}
```

### 5.2 AWS Signature V4 实现

```go
// store/file/s3_signer.go

package file

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "net/http"
    "net/url"
    "sort"
    "strings"
    "time"
)

// S3Signer AWS Signature V4签名器
type S3Signer struct {
    accessKey string
    secretKey string
    region    string
    service   string
}

// NewS3Signer 创建签名器
func NewS3Signer(accessKey, secretKey, region string) *S3Signer {
    return &S3Signer{
        accessKey: accessKey,
        secretKey: secretKey,
        region:    region,
        service:   "s3",
    }
}

// SignRequest 签名HTTP请求
func (s *S3Signer) SignRequest(req *http.Request, payloadHash string) error {
    // 设置时间戳
    now := time.Now().UTC()
    amzDate := now.Format("20060102T150405Z")
    dateStamp := now.Format("20060102")

    // 如果没有提供payload哈希，使用空哈希
    if payloadHash == "" {
        if req.Body != nil {
            payloadHash = "UNSIGNED-PAYLOAD"
        } else {
            payloadHash = emptyStringSHA256()
        }
    }

    // 设置必需的Headers
    req.Header.Set("x-amz-date", amzDate)
    req.Header.Set("x-amz-content-sha256", payloadHash)

    // 如果有Host，确保正确设置
    if req.Host == "" {
        req.Host = req.URL.Host
    }

    // 步骤1：创建规范请求
    canonicalRequest := s.buildCanonicalRequest(req, payloadHash)

    // 步骤2：创建待签名字符串
    credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, s.region, s.service)
    stringToSign := s.buildStringToSign(amzDate, credentialScope, canonicalRequest)

    // 步骤3：计算签名
    signature := s.calculateSignature(dateStamp, stringToSign)

    // 步骤4：添加签名到Authorization header
    authorization := s.buildAuthorizationHeader(signature, credentialScope, req)
    req.Header.Set("Authorization", authorization)

    return nil
}

// PresignRequest 生成预签名URL
func (s *S3Signer) PresignRequest(req *http.Request, expiry time.Duration) (string, error) {
    now := time.Now().UTC()
    amzDate := now.Format("20060102T150405Z")
    dateStamp := now.Format("20060102")

    expirySeconds := int(expiry.Seconds())
    if expirySeconds <= 0 || expirySeconds > 604800 { // 最大7天
        expirySeconds = 900 // 默认15分钟
    }

    // 构建查询参数
    query := req.URL.Query()
    query.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
    query.Set("X-Amz-Credential", fmt.Sprintf("%s/%s/%s/%s/aws4_request",
        s.accessKey, dateStamp, s.region, s.service))
    query.Set("X-Amz-Date", amzDate)
    query.Set("X-Amz-Expires", fmt.Sprintf("%d", expirySeconds))
    query.Set("X-Amz-SignedHeaders", "host")

    // 构建规范请求
    req.URL.RawQuery = query.Encode()
    canonicalRequest := s.buildCanonicalRequestForPresign(req)

    // 构建待签名字符串
    credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, s.region, s.service)
    stringToSign := s.buildStringToSign(amzDate, credentialScope, canonicalRequest)

    // 计算签名
    signature := s.calculateSignature(dateStamp, stringToSign)

    // 添加签名到URL
    query.Set("X-Amz-Signature", signature)
    req.URL.RawQuery = query.Encode()

    return req.URL.String(), nil
}

// buildCanonicalRequest 构建规范请求
func (s *S3Signer) buildCanonicalRequest(req *http.Request, payloadHash string) string {
    // 1. HTTP方法
    method := req.Method

    // 2. 规范URI
    canonicalURI := req.URL.Path
    if canonicalURI == "" {
        canonicalURI = "/"
    }

    // 3. 规范查询字符串
    canonicalQueryString := s.buildCanonicalQueryString(req.URL.Query())

    // 4. 规范Headers
    canonicalHeaders, signedHeaders := s.buildCanonicalHeaders(req.Header, req.Host)

    // 5. 组合规范请求
    canonicalRequest := strings.Join([]string{
        method,
        canonicalURI,
        canonicalQueryString,
        canonicalHeaders,
        signedHeaders,
        payloadHash,
    }, "\n")

    return canonicalRequest
}

// buildCanonicalRequestForPresign 为预签名构建规范请求
func (s *S3Signer) buildCanonicalRequestForPresign(req *http.Request) string {
    method := req.Method
    canonicalURI := req.URL.Path
    if canonicalURI == "" {
        canonicalURI = "/"
    }

    canonicalQueryString := s.buildCanonicalQueryString(req.URL.Query())
    canonicalHeaders := "host:" + req.Host + "\n"
    signedHeaders := "host"
    payloadHash := "UNSIGNED-PAYLOAD"

    canonicalRequest := strings.Join([]string{
        method,
        canonicalURI,
        canonicalQueryString,
        canonicalHeaders,
        signedHeaders,
        payloadHash,
    }, "\n")

    return canonicalRequest
}

// buildCanonicalQueryString 构建规范查询字符串
func (s *S3Signer) buildCanonicalQueryString(values url.Values) string {
    var keys []string
    for k := range values {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    var parts []string
    for _, k := range keys {
        for _, v := range values[k] {
            parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
        }
    }

    return strings.Join(parts, "&")
}

// buildCanonicalHeaders 构建规范Headers
func (s *S3Signer) buildCanonicalHeaders(headers http.Header, host string) (string, string) {
    // 需要签名的Headers
    var headerKeys []string
    headerMap := make(map[string]string)

    // 添加必需的host header
    headerMap["host"] = host
    headerKeys = append(headerKeys, "host")

    // 添加其他Headers
    for k, v := range headers {
        lowerKey := strings.ToLower(k)
        if strings.HasPrefix(lowerKey, "x-amz-") {
            headerMap[lowerKey] = strings.Join(v, ",")
            headerKeys = append(headerKeys, lowerKey)
        }
    }

    sort.Strings(headerKeys)

    // 构建规范Headers字符串
    var canonicalHeaders []string
    for _, k := range headerKeys {
        canonicalHeaders = append(canonicalHeaders, k+":"+headerMap[k])
    }

    canonicalHeadersStr := strings.Join(canonicalHeaders, "\n") + "\n"
    signedHeadersStr := strings.Join(headerKeys, ";")

    return canonicalHeadersStr, signedHeadersStr
}

// buildStringToSign 构建待签名字符串
func (s *S3Signer) buildStringToSign(amzDate, credentialScope, canonicalRequest string) string {
    hash := sha256.Sum256([]byte(canonicalRequest))
    return strings.Join([]string{
        "AWS4-HMAC-SHA256",
        amzDate,
        credentialScope,
        hex.EncodeToString(hash[:]),
    }, "\n")
}

// calculateSignature 计算签名
func (s *S3Signer) calculateSignature(dateStamp, stringToSign string) string {
    // 生成签名密钥
    kDate := hmacSHA256([]byte("AWS4"+s.secretKey), []byte(dateStamp))
    kRegion := hmacSHA256(kDate, []byte(s.region))
    kService := hmacSHA256(kRegion, []byte(s.service))
    kSigning := hmacSHA256(kService, []byte("aws4_request"))

    // 计算签名
    signature := hmacSHA256(kSigning, []byte(stringToSign))
    return hex.EncodeToString(signature)
}

// buildAuthorizationHeader 构建Authorization header
func (s *S3Signer) buildAuthorizationHeader(signature, credentialScope string, req *http.Request) string {
    _, signedHeaders := s.buildCanonicalHeaders(req.Header, req.Host)

    return fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
        s.accessKey,
        credentialScope,
        signedHeaders,
        signature,
    )
}

// hmacSHA256 计算HMAC-SHA256
func hmacSHA256(key, data []byte) []byte {
    h := hmac.New(sha256.New, key)
    h.Write(data)
    return h.Sum(nil)
}

// emptyStringSHA256 空字符串的SHA256哈希
func emptyStringSHA256() string {
    hash := sha256.Sum256([]byte{})
    return hex.EncodeToString(hash[:])
}
```

---

## 6. 图片处理

### 6.1 图片处理器实现

```go
// store/file/image.go

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

// ImageProcessor 图片处理器
type imageProcessor struct{}

// NewImageProcessor 创建图片处理器
func NewImageProcessor() ImageProcessor {
    return &imageProcessor{}
}

// Resize 调整图片大小
func (p *imageProcessor) Resize(src io.Reader, width, height int) (io.Reader, error) {
    // 解码图片
    img, format, err := image.Decode(src)
    if err != nil {
        return nil, &Error{
            Op:  "Resize",
            Err: err,
        }
    }

    // 使用简单的最近邻算法缩放
    resized := resizeImage(img, width, height)

    // 编码回图片
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

// Thumbnail 生成缩略图（保持宽高比）
func (p *imageProcessor) Thumbnail(src io.Reader, maxWidth, maxHeight int) (io.Reader, error) {
    // 解码图片
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

    // 计算缩放比例（保持宽高比）
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

    // 缩放
    resized := resizeImage(img, newWidth, newHeight)

    // 编码
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

// GetInfo 获取图片信息
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

// IsImage 判断是否为图片
func (p *imageProcessor) IsImage(mimeType string) bool {
    mimeType = strings.ToLower(mimeType)
    return mimeType == "image/jpeg" ||
        mimeType == "image/png" ||
        mimeType == "image/gif" ||
        mimeType == "image/webp"
}

// resizeImage 使用最近邻算法缩放图片（简单实现）
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
```

---

## 7. 文件元数据管理

### 7.1 数据库元数据管理器

```go
// store/file/metadata.go

package file

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"
)

// DBMetadataManager 基于数据库的元数据管理器
type DBMetadataManager struct {
    db *sql.DB
}

// NewDBMetadataManager 创建元数据管理器
func NewDBMetadataManager(db *sql.DB) MetadataManager {
    return &DBMetadataManager{db: db}
}

// Save 保存文件元数据
func (m *DBMetadataManager) Save(ctx context.Context, file *File) error {
    metadataJSON, err := json.Marshal(file.Metadata)
    if err != nil {
        return err
    }

    query := `
        INSERT INTO files
        (id, tenant_id, name, path, size, mime_type, extension, hash, width, height,
         thumbnail_path, storage_type, metadata, uploaded_by, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            size = EXCLUDED.size,
            updated_at = EXCLUDED.updated_at
    `

    _, err = m.db.ExecContext(ctx, query,
        file.ID, file.TenantID, file.Name, file.Path, file.Size,
        file.MimeType, file.Extension, file.Hash, file.Width, file.Height,
        file.ThumbnailPath, file.StorageType, metadataJSON,
        file.UploadedBy, file.CreatedAt, file.UpdatedAt,
    )

    return err
}

// Get 获取文件元数据
func (m *DBMetadataManager) Get(ctx context.Context, id string) (*File, error) {
    query := `
        SELECT id, tenant_id, name, path, size, mime_type, extension, hash,
               width, height, thumbnail_path, storage_type, metadata,
               uploaded_by, created_at, updated_at, last_access_at, deleted_at
        FROM files
        WHERE id = $1 AND deleted_at IS NULL
    `

    var file File
    var metadataJSON []byte

    err := m.db.QueryRowContext(ctx, query, id).Scan(
        &file.ID, &file.TenantID, &file.Name, &file.Path, &file.Size,
        &file.MimeType, &file.Extension, &file.Hash, &file.Width, &file.Height,
        &file.ThumbnailPath, &file.StorageType, &metadataJSON,
        &file.UploadedBy, &file.CreatedAt, &file.UpdatedAt,
        &file.LastAccessAt, &file.DeletedAt,
    )

    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, err
    }

    if len(metadataJSON) > 0 {
        json.Unmarshal(metadataJSON, &file.Metadata)
    }

    return &file, nil
}

// GetByPath 根据路径获取元数据
func (m *DBMetadataManager) GetByPath(ctx context.Context, path string) (*File, error) {
    query := `
        SELECT id, tenant_id, name, path, size, mime_type, extension, hash,
               width, height, thumbnail_path, storage_type, metadata,
               uploaded_by, created_at, updated_at, last_access_at, deleted_at
        FROM files
        WHERE path = $1 AND deleted_at IS NULL
    `

    var file File
    var metadataJSON []byte

    err := m.db.QueryRowContext(ctx, query, path).Scan(
        &file.ID, &file.TenantID, &file.Name, &file.Path, &file.Size,
        &file.MimeType, &file.Extension, &file.Hash, &file.Width, &file.Height,
        &file.ThumbnailPath, &file.StorageType, &metadataJSON,
        &file.UploadedBy, &file.CreatedAt, &file.UpdatedAt,
        &file.LastAccessAt, &file.DeletedAt,
    )

    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, err
    }

    if len(metadataJSON) > 0 {
        json.Unmarshal(metadataJSON, &file.Metadata)
    }

    return &file, nil
}

// GetByHash 根据哈希获取元数据（去重）
func (m *DBMetadataManager) GetByHash(ctx context.Context, hash string) (*File, error) {
    query := `
        SELECT id, tenant_id, name, path, size, mime_type, extension, hash,
               width, height, thumbnail_path, storage_type, metadata,
               uploaded_by, created_at, updated_at, last_access_at, deleted_at
        FROM files
        WHERE hash = $1 AND deleted_at IS NULL
        LIMIT 1
    `

    var file File
    var metadataJSON []byte

    err := m.db.QueryRowContext(ctx, query, hash).Scan(
        &file.ID, &file.TenantID, &file.Name, &file.Path, &file.Size,
        &file.MimeType, &file.Extension, &file.Hash, &file.Width, &file.Height,
        &file.ThumbnailPath, &file.StorageType, &metadataJSON,
        &file.UploadedBy, &file.CreatedAt, &file.UpdatedAt,
        &file.LastAccessAt, &file.DeletedAt,
    )

    if err == sql.ErrNoRows {
        return nil, nil // 没有找到不是错误
    }
    if err != nil {
        return nil, err
    }

    if len(metadataJSON) > 0 {
        json.Unmarshal(metadataJSON, &file.Metadata)
    }

    return &file, nil
}

// List 列出文件元数据
func (m *DBMetadataManager) List(ctx context.Context, query Query) ([]*File, int64, error) {
    // 构建WHERE条件
    conditions := []string{"deleted_at IS NULL"}
    args := []interface{}{}
    argIndex := 1

    if query.TenantID != "" {
        conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
        args = append(args, query.TenantID)
        argIndex++
    }

    if query.UploadedBy != "" {
        conditions = append(conditions, fmt.Sprintf("uploaded_by = $%d", argIndex))
        args = append(args, query.UploadedBy)
        argIndex++
    }

    if query.MimeType != "" {
        conditions = append(conditions, fmt.Sprintf("mime_type = $%d", argIndex))
        args = append(args, query.MimeType)
        argIndex++
    }

    if !query.StartTime.IsZero() {
        conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
        args = append(args, query.StartTime)
        argIndex++
    }

    if !query.EndTime.IsZero() {
        conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
        args = append(args, query.EndTime)
        argIndex++
    }

    whereClause := ""
    if len(conditions) > 0 {
        whereClause = "WHERE " + strings.Join(conditions, " AND ")
    }

    // 查询总数
    var total int64
    countQuery := fmt.Sprintf("SELECT COUNT(*) FROM files %s", whereClause)
    err := m.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
    if err != nil {
        return nil, 0, err
    }

    // 分页查询
    if query.Page < 1 {
        query.Page = 1
    }
    if query.PageSize < 1 {
        query.PageSize = 20
    }
    offset := (query.Page - 1) * query.PageSize

    orderBy := "created_at DESC"
    if query.OrderBy != "" {
        orderBy = query.OrderBy
    }

    listQuery := fmt.Sprintf(`
        SELECT id, tenant_id, name, path, size, mime_type, extension, hash,
               width, height, thumbnail_path, storage_type, metadata,
               uploaded_by, created_at, updated_at, last_access_at, deleted_at
        FROM files
        %s
        ORDER BY %s
        LIMIT $%d OFFSET $%d
    `, whereClause, orderBy, argIndex, argIndex+1)

    args = append(args, query.PageSize, offset)

    rows, err := m.db.QueryContext(ctx, listQuery, args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var results []*File
    for rows.Next() {
        var file File
        var metadataJSON []byte

        err := rows.Scan(
            &file.ID, &file.TenantID, &file.Name, &file.Path, &file.Size,
            &file.MimeType, &file.Extension, &file.Hash, &file.Width, &file.Height,
            &file.ThumbnailPath, &file.StorageType, &metadataJSON,
            &file.UploadedBy, &file.CreatedAt, &file.UpdatedAt,
            &file.LastAccessAt, &file.DeletedAt,
        )
        if err != nil {
            return nil, 0, err
        }

        if len(metadataJSON) > 0 {
            json.Unmarshal(metadataJSON, &file.Metadata)
        }

        results = append(results, &file)
    }

    return results, total, nil
}

// Delete 删除元数据（软删除）
func (m *DBMetadataManager) Delete(ctx context.Context, id string) error {
    query := `UPDATE files SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
    result, err := m.db.ExecContext(ctx, query, time.Now(), id)
    if err != nil {
        return err
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        return ErrNotFound
    }

    return nil
}

// UpdateAccessTime 更新访问时间
func (m *DBMetadataManager) UpdateAccessTime(ctx context.Context, id string) error {
    query := `UPDATE files SET last_access_at = $1 WHERE id = $2`
    _, err := m.db.ExecContext(ctx, query, time.Now(), id)
    return err
}
```

### 7.2 数据库Schema

```sql
-- 文件元数据表
CREATE TABLE files (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    path VARCHAR(512) NOT NULL,
    size BIGINT NOT NULL,
    mime_type VARCHAR(128),
    extension VARCHAR(32),
    hash VARCHAR(64) NOT NULL,
    width INT DEFAULT 0,
    height INT DEFAULT 0,
    thumbnail_path VARCHAR(512),
    storage_type VARCHAR(16) NOT NULL DEFAULT 'local',
    metadata JSONB,
    uploaded_by VARCHAR(64) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_access_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 索引
CREATE INDEX idx_files_tenant ON files(tenant_id);
CREATE INDEX idx_files_hash ON files(hash);
CREATE INDEX idx_files_uploaded_by ON files(uploaded_by);
CREATE INDEX idx_files_created_at ON files(created_at DESC);
CREATE INDEX idx_files_path ON files(path);

-- 唯一约束（同租户同哈希去重）
CREATE UNIQUE INDEX idx_files_tenant_hash ON files(tenant_id, hash) WHERE deleted_at IS NULL;
```

---

## 8. HTTP处理器

### 8.1 HTTP Handler实现

```go
// store/file/handler.go

package file

import (
    "fmt"
    "io"
    "net/http"
    "strconv"
    "strings"
    "time"
)

// Handler HTTP处理器
type Handler struct {
    storage  Storage
    metadata MetadataManager
}

// NewHandler 创建HTTP处理器
func NewHandler(storage Storage, metadata MetadataManager) *Handler {
    return &Handler{
        storage:  storage,
        metadata: metadata,
    }
}

// Upload 处理文件上传
// POST /files
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // 解析multipart form (最大32MB)
    if err := r.ParseMultipartForm(32 << 20); err != nil {
        http.Error(w, fmt.Sprintf("Parse form error: %v", err), http.StatusBadRequest)
        return
    }

    // 获取文件
    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, fmt.Sprintf("Get file error: %v", err), http.StatusBadRequest)
        return
    }
    defer file.Close()

    // 获取租户ID（从header或query）
    tenantID := r.Header.Get("X-Tenant-ID")
    if tenantID == "" {
        tenantID = r.URL.Query().Get("tenant_id")
    }
    if tenantID == "" {
        http.Error(w, "tenant_id required", http.StatusBadRequest)
        return
    }

    // 获取上传人
    uploadedBy := r.Header.Get("X-User-ID")
    if uploadedBy == "" {
        uploadedBy = "anonymous"
    }

    // 解析选项
    generateThumb := r.FormValue("generate_thumb") == "true"
    thumbWidth, _ := strconv.Atoi(r.FormValue("thumb_width"))
    thumbHeight, _ := strconv.Atoi(r.FormValue("thumb_height"))

    // 上传文件
    opts := PutOptions{
        TenantID:      tenantID,
        Reader:        file,
        FileName:      header.Filename,
        ContentType:   header.Header.Get("Content-Type"),
        Size:          header.Size,
        UploadedBy:    uploadedBy,
        GenerateThumb: generateThumb,
        ThumbWidth:    thumbWidth,
        ThumbHeight:   thumbHeight,
    }

    result, err := h.storage.Put(r.Context(), opts)
    if err != nil {
        http.Error(w, fmt.Sprintf("Upload error: %v", err), http.StatusInternalServerError)
        return
    }

    // 返回JSON
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    fmt.Fprintf(w, `{"id":"%s","path":"%s","size":%d}`, result.ID, result.Path, result.Size)
}

// Download 处理文件下载
// GET /files/:id
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // 从URL路径提取文件ID
    path := strings.TrimPrefix(r.URL.Path, "/files/")
    fileID := strings.Split(path, "/")[0]

    if fileID == "" {
        http.Error(w, "file_id required", http.StatusBadRequest)
        return
    }

    // 获取文件元数据
    fileInfo, err := h.metadata.Get(r.Context(), fileID)
    if err != nil {
        if err == ErrNotFound {
            http.Error(w, "File not found", http.StatusNotFound)
        } else {
            http.Error(w, fmt.Sprintf("Get metadata error: %v", err), http.StatusInternalServerError)
        }
        return
    }

    // 获取文件内容
    reader, err := h.storage.Get(r.Context(), fileInfo.Path)
    if err != nil {
        http.Error(w, fmt.Sprintf("Get file error: %v", err), http.StatusInternalServerError)
        return
    }
    defer reader.Close()

    // 设置响应头
    w.Header().Set("Content-Type", fileInfo.MimeType)
    w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size))
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name))

    // 流式传输文件
    io.Copy(w, reader)

    // 更新访问时间（异步）
    go h.metadata.UpdateAccessTime(r.Context(), fileID)
}

// GetInfo 获取文件信息
// GET /files/:id/info
func (h *Handler) GetInfo(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    path := strings.TrimPrefix(r.URL.Path, "/files/")
    fileID := strings.Split(path, "/")[0]

    if fileID == "" {
        http.Error(w, "file_id required", http.StatusBadRequest)
        return
    }

    fileInfo, err := h.metadata.Get(r.Context(), fileID)
    if err != nil {
        if err == ErrNotFound {
            http.Error(w, "File not found", http.StatusNotFound)
        } else {
            http.Error(w, fmt.Sprintf("Get metadata error: %v", err), http.StatusInternalServerError)
        }
        return
    }

    // 返回JSON
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, `{"id":"%s","name":"%s","path":"%s","size":%d,"mime_type":"%s","created_at":"%s"}`,
        fileInfo.ID, fileInfo.Name, fileInfo.Path, fileInfo.Size,
        fileInfo.MimeType, fileInfo.CreatedAt.Format(time.RFC3339))
}

// Delete 删除文件
// DELETE /files/:id
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    path := strings.TrimPrefix(r.URL.Path, "/files/")
    fileID := strings.Split(path, "/")[0]

    if fileID == "" {
        http.Error(w, "file_id required", http.StatusBadRequest)
        return
    }

    // 获取文件元数据
    fileInfo, err := h.metadata.Get(r.Context(), fileID)
    if err != nil {
        if err == ErrNotFound {
            http.Error(w, "File not found", http.StatusNotFound)
        } else {
            http.Error(w, fmt.Sprintf("Get metadata error: %v", err), http.StatusInternalServerError)
        }
        return
    }

    // 删除物理文件
    if err := h.storage.Delete(r.Context(), fileInfo.Path); err != nil {
        http.Error(w, fmt.Sprintf("Delete file error: %v", err), http.StatusInternalServerError)
        return
    }

    // 删除元数据（软删除）
    if err := h.metadata.Delete(r.Context(), fileID); err != nil {
        http.Error(w, fmt.Sprintf("Delete metadata error: %v", err), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// List 列出文件
// GET /files
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    query := Query{
        TenantID:   r.URL.Query().Get("tenant_id"),
        UploadedBy: r.URL.Query().Get("uploaded_by"),
        MimeType:   r.URL.Query().Get("mime_type"),
    }

    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
    if page < 1 {
        page = 1
    }
    if pageSize < 1 {
        pageSize = 20
    }
    query.Page = page
    query.PageSize = pageSize

    files, total, err := h.metadata.List(r.Context(), query)
    if err != nil {
        http.Error(w, fmt.Sprintf("List error: %v", err), http.StatusInternalServerError)
        return
    }

    // 返回JSON
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, `{"items":[`)
    for i, f := range files {
        if i > 0 {
            fmt.Fprintf(w, ",")
        }
        fmt.Fprintf(w, `{"id":"%s","name":"%s","size":%d}`, f.ID, f.Name, f.Size)
    }
    fmt.Fprintf(w, `],"total":%d,"page":%d,"page_size":%d}`, total, page, pageSize)
}
```

---

## 9. 使用示例

### 9.1 本地存储使用

```go
// examples/file/local_example.go

package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"

    _ "github.com/lib/pq"
    "github.com/spcent/plumego/store/file"
)

func main() {
    // 连接数据库
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/plumego?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // 创建元数据管理器
    metadata := file.NewDBMetadataManager(db)

    // 创建本地存储
    storage, err := file.NewLocalStorage("./uploads", "http://localhost:8080/uploads", metadata)
    if err != nil {
        log.Fatal(err)
    }

    // 上传文件示例
    uploadExample(storage)

    // 下载文件示例
    downloadExample(storage)

    // 启动HTTP服务器
    handler := file.NewHandler(storage, metadata)
    http.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodPost:
            handler.Upload(w, r)
        case http.MethodGet:
            handler.List(w, r)
        }
    })
    http.HandleFunc("/files/", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            handler.Download(w, r)
        case http.MethodDelete:
            handler.Delete(w, r)
        }
    })

    log.Println("Server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func uploadExample(storage file.Storage) {
    // 打开文件
    f, err := os.Open("test.jpg")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    // 上传
    ctx := context.Background()
    result, err := storage.Put(ctx, file.PutOptions{
        TenantID:      "tenant-123",
        Reader:        f,
        FileName:      "test.jpg",
        ContentType:   "image/jpeg",
        UploadedBy:    "user-456",
        GenerateThumb: true,
        ThumbWidth:    200,
        ThumbHeight:   200,
    })

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Uploaded: ID=%s, Path=%s, Size=%d\n", result.ID, result.Path, result.Size)
}

func downloadExample(storage file.Storage) {
    ctx := context.Background()

    // 获取文件
    reader, err := storage.Get(ctx, "path/to/file.jpg")
    if err != nil {
        log.Fatal(err)
    }
    defer reader.Close()

    // 保存到本地
    out, err := os.Create("downloaded.jpg")
    if err != nil {
        log.Fatal(err)
    }
    defer out.Close()

    io.Copy(out, reader)
    fmt.Println("Downloaded successfully")
}
```

### 9.2 S3存储使用

```go
// examples/file/s3_example.go

package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/lib/pq"
    "github.com/spcent/plumego/store/file"
)

func main() {
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/plumego?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    metadata := file.NewDBMetadataManager(db)

    // 创建S3存储
    config := file.StorageConfig{
        Type:         "s3",
        S3Endpoint:   "s3.amazonaws.com",
        S3Region:     "us-east-1",
        S3Bucket:     "my-bucket",
        S3AccessKey:  os.Getenv("AWS_ACCESS_KEY_ID"),
        S3SecretKey:  os.Getenv("AWS_SECRET_ACCESS_KEY"),
        S3UseSSL:     true,
        S3PathStyle:  false,
    }

    storage, err := file.NewS3Storage(config, metadata)
    if err != nil {
        log.Fatal(err)
    }

    // 上传文件
    f, err := os.Open("test.jpg")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    ctx := context.Background()
    result, err := storage.Put(ctx, file.PutOptions{
        TenantID:    "tenant-123",
        Reader:      f,
        FileName:    "test.jpg",
        ContentType: "image/jpeg",
        UploadedBy:  "user-456",
    })

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Uploaded to S3: ID=%s, Path=%s\n", result.ID, result.Path)

    // 生成预签名URL
    url, err := storage.GetURL(ctx, result.Path, 15*time.Minute)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Presigned URL: %s\n", url)
}
```

---

## 10. 测试

### 10.1 本地存储测试

```go
// store/file/local_test.go

package file_test

import (
    "bytes"
    "context"
    "os"
    "path/filepath"
    "testing"

    "github.com/spcent/plumego/store/file"
)

func TestLocalStorage_PutAndGet(t *testing.T) {
    // 创建临时目录
    tmpDir, err := os.MkdirTemp("", "file-test-*")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    // 创建存储
    storage, err := file.NewLocalStorage(tmpDir, "http://localhost", nil)
    if err != nil {
        t.Fatal(err)
    }

    // 测试数据
    content := []byte("Hello, World!")
    reader := bytes.NewReader(content)

    // 上传
    ctx := context.Background()
    result, err := storage.Put(ctx, file.PutOptions{
        TenantID:    "test-tenant",
        Reader:      reader,
        FileName:    "test.txt",
        ContentType: "text/plain",
        UploadedBy:  "test-user",
    })

    if err != nil {
        t.Fatalf("Put failed: %v", err)
    }

    if result.ID == "" {
        t.Error("Expected ID, got empty")
    }

    // 验证文件存在
    exists, err := storage.Exists(ctx, result.Path)
    if err != nil {
        t.Fatalf("Exists failed: %v", err)
    }
    if !exists {
        t.Error("File should exist")
    }

    // 下载
    downloaded, err := storage.Get(ctx, result.Path)
    if err != nil {
        t.Fatalf("Get failed: %v", err)
    }
    defer downloaded.Close()

    buf := new(bytes.Buffer)
    buf.ReadFrom(downloaded)

    if !bytes.Equal(buf.Bytes(), content) {
        t.Errorf("Expected %s, got %s", content, buf.Bytes())
    }

    // 删除
    err = storage.Delete(ctx, result.Path)
    if err != nil {
        t.Fatalf("Delete failed: %v", err)
    }

    // 验证文件已删除
    exists, err = storage.Exists(ctx, result.Path)
    if err != nil {
        t.Fatalf("Exists after delete failed: %v", err)
    }
    if exists {
        t.Error("File should not exist after delete")
    }
}

func TestLocalStorage_PathSafety(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "file-test-*")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    storage, err := file.NewLocalStorage(tmpDir, "http://localhost", nil)
    if err != nil {
        t.Fatal(err)
    }

    ctx := context.Background()

    // 测试路径穿越攻击
    tests := []string{
        "../etc/passwd",
        "../../secret",
        "/etc/passwd",
    }

    for _, path := range tests {
        _, err := storage.Get(ctx, path)
        if err == nil {
            t.Errorf("Expected error for dangerous path %s, got nil", path)
        }
    }
}
```

### 10.2 图片处理测试

```go
// store/file/image_test.go

package file_test

import (
    "bytes"
    "image"
    "image/color"
    "image/jpeg"
    "testing"

    "github.com/spcent/plumego/store/file"
)

func TestImageProcessor_Thumbnail(t *testing.T) {
    // 创建测试图片
    img := image.NewRGBA(image.Rect(0, 0, 800, 600))
    for y := 0; y < 600; y++ {
        for x := 0; x < 800; x++ {
            img.Set(x, y, color.RGBA{255, 0, 0, 255})
        }
    }

    buf := new(bytes.Buffer)
    jpeg.Encode(buf, img, nil)

    // 创建处理器
    processor := file.NewImageProcessor()

    // 生成缩略图
    thumb, err := processor.Thumbnail(bytes.NewReader(buf.Bytes()), 200, 200)
    if err != nil {
        t.Fatalf("Thumbnail failed: %v", err)
    }

    // 解码缩略图
    thumbImg, _, err := image.Decode(thumb)
    if err != nil {
        t.Fatalf("Decode thumbnail failed: %v", err)
    }

    bounds := thumbImg.Bounds()
    width := bounds.Dx()
    height := bounds.Dy()

    // 验证尺寸（应保持宽高比）
    if width > 200 || height > 200 {
        t.Errorf("Thumbnail too large: %dx%d", width, height)
    }

    // 验证宽高比
    ratio := float64(width) / float64(height)
    expected := 800.0 / 600.0
    if ratio < expected-0.1 || ratio > expected+0.1 {
        t.Errorf("Aspect ratio not preserved: got %f, expected %f", ratio, expected)
    }
}
```

---

## 11. 总结

### 11.1 模块特点

✅ **标准库实现** - 无第三方依赖，纯Go标准库
✅ **接口分离** - 清晰的接口定义，易于测试和扩展
✅ **多存储支持** - 统一接口支持本地和S3
✅ **AWS兼容** - 手动实现Signature V4，兼容所有S3 API
✅ **图片处理** - 内置缩略图生成和图片信息提取
✅ **文件去重** - 基于SHA256哈希自动去重
✅ **安全可靠** - 路径安全检查、软删除、访问控制

### 11.2 目录结构总结

```
store/file/
├── file.go          # Storage接口定义
├── types.go         # 数据类型
├── errors.go        # 错误定义
├── metadata.go      # 元数据管理器
├── local.go         # 本地存储实现
├── s3.go            # S3存储实现
├── s3_signer.go     # AWS Signature V4
├── image.go         # 图片处理
├── handler.go       # HTTP处理器
├── utils.go         # 工具函数
└── *_test.go        # 测试文件
```

### 11.3 性能优化

- 流式处理，避免大文件内存溢出
- SHA256哈希去重，节省存储空间
- 软删除机制，可恢复误删文件
- 访问时间记录，支持LRU清理策略
- HTTP预签名URL，避免代理流量

### 11.4 安全措施

- 路径穿越防护
- 文件类型验证
- 大小限制控制
- 租户数据隔离
- 访问权限控制

### 11.5 扩展建议

未来可扩展功能：
- 视频处理（ffmpeg集成）
- CDN分发支持
- 文件版本控制
- 流量统计分析
- 智能压缩优化

---

**文档版本**: v1.0.0
**最后更新**: 2026-02-05
**维护者**: Plumego Storage Team