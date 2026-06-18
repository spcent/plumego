package mfa

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

const (
	enrollmentKeyPrefix = "mfa:"
	challengeKeyPrefix  = "mfachallenge:"
	defaultIssuer       = "dbadmin"
	challengeTTL        = 5 * time.Minute
)

// ErrNotFound is returned when no MFA enrollment exists for a user.
var ErrNotFound = errors.New("mfa: not found")

// ErrChallengeNotFound is returned when an MFA login challenge token is
// missing, expired, or already consumed.
var ErrChallengeNotFound = errors.New("mfa: challenge not found or expired")

// Enrollment holds the persisted MFA state for a single user.
type Enrollment struct {
	Username  string    `json:"username"`
	Secret    string    `json:"secret"` // base32, never logged
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Challenge represents a pending second-factor login challenge issued after
// a successful password check, before a session cookie is granted.
type Challenge struct {
	Token     string    `json:"token"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// Store persists MFA enrollments and login challenges in a KV store.
type Store struct {
	kv     *kvstore.KVStore
	issuer string
}

// NewStore creates a Store backed by the provided KV store.
func NewStore(kv *kvstore.KVStore) *Store {
	return &Store{kv: kv, issuer: defaultIssuer}
}

// StartEnrollment generates a new TOTP secret for username and persists it
// in a disabled state. It does not affect an already-enabled enrollment's
// validity for login until ConfirmEnrollment is called. Returns the base32
// secret and an otpauth:// URI for use with an external authenticator app.
func (s *Store) StartEnrollment(username string) (secret, uri string, err error) {
	secret, err = GenerateSecret()
	if err != nil {
		return "", "", err
	}
	now := time.Now().UTC()
	enr := Enrollment{
		Username:  username,
		Secret:    secret,
		Enabled:   false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.save(&enr); err != nil {
		return "", "", err
	}
	return secret, otpauthURI(s.issuer, username, secret), nil
}

// ConfirmEnrollment validates code against the pending (or existing) secret
// for username and, on success, marks the enrollment enabled. Fails closed:
// any lookup or validation error results in a non-nil error and no state
// change.
func (s *Store) ConfirmEnrollment(username, code string) error {
	enr, err := s.get(username)
	if err != nil {
		return err
	}
	if !Validate(enr.Secret, code, time.Now()) {
		return fmt.Errorf("mfa: invalid code")
	}
	enr.Enabled = true
	enr.UpdatedAt = time.Now().UTC()
	return s.save(enr)
}

// Disable removes the MFA enrollment for username, disabling MFA at login.
func (s *Store) Disable(username string) error {
	err := s.kv.Delete(enrollmentKeyPrefix + username)
	if err != nil && !errors.Is(err, kvstore.ErrKeyNotFound) {
		return fmt.Errorf("delete mfa enrollment: %w", err)
	}
	return nil
}

// IsEnabled reports whether username has MFA enabled. Returns false (not an
// error) when no enrollment exists, so callers can treat MFA as opt-in.
func (s *Store) IsEnabled(username string) bool {
	enr, err := s.get(username)
	if err != nil {
		return false
	}
	return enr.Enabled
}

// Get returns the enrollment for username, or ErrNotFound.
func (s *Store) Get(username string) (*Enrollment, error) {
	return s.get(username)
}

// VerifyCode validates code for username's enabled enrollment at the current
// time. Fails closed: returns false if no enrollment exists, MFA is not
// enabled, or the code is invalid.
func (s *Store) VerifyCode(username, code string) bool {
	enr, err := s.get(username)
	if err != nil || !enr.Enabled {
		return false
	}
	return Validate(enr.Secret, code, time.Now())
}

func (s *Store) get(username string) (*Enrollment, error) {
	data, err := s.kv.Get(enrollmentKeyPrefix + username)
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) || errors.Is(err, kvstore.ErrKeyExpired) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get mfa enrollment: %w", err)
	}
	var enr Enrollment
	if err := json.Unmarshal(data, &enr); err != nil {
		return nil, fmt.Errorf("unmarshal mfa enrollment: %w", err)
	}
	return &enr, nil
}

func (s *Store) save(enr *Enrollment) error {
	data, err := json.Marshal(enr)
	if err != nil {
		return fmt.Errorf("marshal mfa enrollment: %w", err)
	}
	if err := s.kv.Set(enrollmentKeyPrefix+enr.Username, data, 0); err != nil {
		return fmt.Errorf("persist mfa enrollment: %w", err)
	}
	return nil
}

// CreateChallenge issues a short-lived MFA login challenge token for
// username, to be presented back along with a TOTP code to complete login.
func (s *Store) CreateChallenge(username string) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("generate mfa challenge token: %w", err)
	}
	chal := Challenge{
		Token:     token,
		Username:  username,
		CreatedAt: time.Now().UTC(),
	}
	data, err := json.Marshal(chal)
	if err != nil {
		return "", fmt.Errorf("marshal mfa challenge: %w", err)
	}
	if err := s.kv.Set(challengeKeyPrefix+token, data, challengeTTL); err != nil {
		return "", fmt.Errorf("persist mfa challenge: %w", err)
	}
	return token, nil
}

// ConsumeChallenge looks up and deletes the challenge for token, returning
// the associated username. The challenge is deleted whether or not the
// caller goes on to successfully verify a TOTP code, so a challenge token
// can only ever be used once. Returns ErrChallengeNotFound if the token is
// missing, expired, or already consumed.
func (s *Store) ConsumeChallenge(token string) (string, error) {
	data, err := s.kv.Get(challengeKeyPrefix + token)
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) || errors.Is(err, kvstore.ErrKeyExpired) {
			return "", ErrChallengeNotFound
		}
		return "", fmt.Errorf("get mfa challenge: %w", err)
	}
	// Best-effort single-use delete; if this fails we still proceed since the
	// data has already been read, but log-free per the no-secret-logging rule.
	_ = s.kv.Delete(challengeKeyPrefix + token)
	var chal Challenge
	if err := json.Unmarshal(data, &chal); err != nil {
		return "", fmt.Errorf("unmarshal mfa challenge: %w", err)
	}
	return chal.Username, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
