// Package connection manages saved database connection configurations.
package connection

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

// DriverType identifies the database driver.
// When adding a new driver: add its constant here, extend validateConnection in
// connections.go, add a DSN builder case in dbmanager/manager.go, and register
// handler routes in app/routes.go under a dedicated route group.
type DriverType string

const (
	DriverMySQL  DriverType = "mysql"
	DriverSQLite DriverType = "sqlite"

	DriverRedis DriverType = "redis" // supported: save/load config; driver not yet implemented

	// Planned drivers — not yet implemented:
	// DriverMongoDB       DriverType = "mongodb"
	// DriverElasticsearch DriverType = "elasticsearch"
)

var (
	ErrNotFound  = errors.New("connection: not found")
	ErrDuplicate = errors.New("connection: ID already exists")

	idKeyPrefix = "conn:"
	listKey     = "conn:__index__"
)

// Connection is a saved database connection configuration.
type Connection struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Driver   DriverType `json:"driver"`
	Host     string     `json:"host,omitempty"`
	Port     int        `json:"port,omitempty"`
	Database string     `json:"database,omitempty"`
	Username string     `json:"username,omitempty"`
	Password string     `json:"password,omitempty"`  // stored encrypted when SavePassword=true
	FilePath string     `json:"file_path,omitempty"` // for SQLite
	Options  string     `json:"options,omitempty"`   // extra DSN params
	// Redis-specific fields (used when Driver = "redis")
	RedisDBIndex     int       `json:"redis_db_index,omitempty"`    // logical DB index 0-15, default 0
	TLSEnabled       bool      `json:"tls_enabled,omitempty"`       // use TLS/SSL
	Readonly         bool      `json:"readonly,omitempty"`          // disallow all write operations
	SavePassword     bool      `json:"save_password,omitempty"`     // persist password to disk
	UploadedFile     bool      `json:"uploaded_file,omitempty"`     // file_path is a server-managed temp file
	OriginalFilename string    `json:"original_filename,omitempty"` // user's original upload filename
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Store persists connections in a KV store, encrypting passwords with AES-GCM.
type Store struct {
	kv  *kvstore.KVStore
	key []byte // 32-byte AES-GCM key; nil means no encryption
}

// NewStore creates a connection Store. encryptionKeyHex may be empty (disables encryption).
func NewStore(kv *kvstore.KVStore, encryptionKeyHex string) (*Store, error) {
	s := &Store{kv: kv}
	if encryptionKeyHex != "" {
		key, err := hex.DecodeString(encryptionKeyHex)
		if err != nil {
			return nil, fmt.Errorf("decode encryption key: %w", err)
		}
		if len(key) != 32 {
			return nil, fmt.Errorf("encryption key must be 32 bytes (64 hex chars), got %d", len(key))
		}
		s.key = key
	}
	return s, nil
}

// List returns all saved connections (passwords redacted).
func (s *Store) List() ([]*Connection, error) {
	ids, err := s.loadIndex()
	if err != nil {
		return nil, err
	}
	conns := make([]*Connection, 0, len(ids))
	for _, id := range ids {
		c, err := s.Get(id)
		if err != nil {
			continue
		}
		c.Password = ""
		conns = append(conns, c)
	}
	return conns, nil
}

// Get returns a connection by ID (password decrypted).
func (s *Store) Get(id string) (*Connection, error) {
	data, err := s.kv.Get(idKeyPrefix + id)
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) || errors.Is(err, kvstore.ErrKeyExpired) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get connection: %w", err)
	}
	var c Connection
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("unmarshal connection: %w", err)
	}
	if c.Password != "" && s.key != nil {
		plain, err := decrypt(s.key, c.Password)
		if err != nil {
			return nil, fmt.Errorf("decrypt password: %w", err)
		}
		c.Password = plain
	}
	return &c, nil
}

// Create saves a new connection and returns it.
func (s *Store) Create(c *Connection) error {
	if c.ID == "" {
		id, err := generateID()
		if err != nil {
			return err
		}
		c.ID = id
	}
	c.CreatedAt = time.Now().UTC()
	c.UpdatedAt = c.CreatedAt
	return s.save(c, true)
}

// Update replaces an existing connection.
func (s *Store) Update(c *Connection) error {
	existing, err := s.Get(c.ID)
	if err != nil {
		return err
	}
	if !c.SavePassword {
		// User opted out of saving password — clear any stored value.
		c.Password = ""
	} else if c.Password == "" {
		// SavePassword=true but no new password provided — keep the existing one.
		c.Password = existing.Password
	}
	c.CreatedAt = existing.CreatedAt
	c.UpdatedAt = time.Now().UTC()
	return s.save(c, false)
}

// Delete removes a connection by ID.
func (s *Store) Delete(id string) error {
	if err := s.kv.Delete(idKeyPrefix + id); err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("delete connection: %w", err)
	}
	ids, _ := s.loadIndex()
	updated := make([]string, 0, len(ids))
	for _, existing := range ids {
		if existing != id {
			updated = append(updated, existing)
		}
	}
	return s.saveIndex(updated)
}

func (s *Store) save(c *Connection, isNew bool) error {
	toStore := *c
	if !toStore.SavePassword {
		toStore.Password = "" // never persist when user opted out
	} else if toStore.Password != "" && s.key != nil {
		enc, err := encrypt(s.key, toStore.Password)
		if err != nil {
			return fmt.Errorf("encrypt password: %w", err)
		}
		toStore.Password = enc
	}
	data, err := json.Marshal(toStore)
	if err != nil {
		return fmt.Errorf("marshal connection: %w", err)
	}
	if err := s.kv.Set(idKeyPrefix+c.ID, data, 0); err != nil {
		return fmt.Errorf("persist connection: %w", err)
	}
	if isNew {
		ids, _ := s.loadIndex()
		ids = append(ids, c.ID)
		return s.saveIndex(ids)
	}
	return nil
}

func (s *Store) loadIndex() ([]string, error) {
	data, err := s.kv.Get(listKey)
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) || errors.Is(err, kvstore.ErrKeyExpired) {
			return nil, nil
		}
		return nil, err
	}
	var ids []string
	if err := json.Unmarshal(data, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func (s *Store) saveIndex(ids []string) error {
	data, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	return s.kv.Set(listKey, data, 0)
}

func generateID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func encrypt(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

func decrypt(key []byte, ciphertextHex string) (string, error) {
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
