package user

import (
	"context"
	"strings"
	"time"

	"github.com/spcent/plumego/security/password"
	"mini-saas-api/internal/domain/ident"
)

// Service implements account registration and credential verification.
type Service struct {
	repo     Repository
	strength password.PasswordStrengthConfig
}

// NewService constructs a Service over the given repository.
func NewService(repo Repository) *Service {
	return &Service{
		repo:     repo,
		strength: password.DefaultPasswordStrengthConfig(),
	}
}

// Register creates a new account. The email is lowercased for uniqueness;
// the password is strength-checked and stored only as a bcrypt hash.
func (s *Service) Register(ctx context.Context, email, name, plain string) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") {
		return User{}, ErrInvalidCredentials
	}
	if !password.ValidatePasswordStrength(plain, s.strength) {
		return User{}, ErrWeakPassword
	}
	hash, err := password.HashPassword(plain)
	if err != nil {
		return User{}, err
	}
	id, err := ident.New()
	if err != nil {
		return User{}, err
	}
	u := User{
		ID:           id,
		Email:        email,
		Name:         strings.TrimSpace(name),
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return User{}, err
	}
	return u, nil
}

// Authenticate verifies email + password and returns the account.
// Unknown email and wrong password both return ErrInvalidCredentials so the
// response does not reveal which check failed.
func (s *Service) Authenticate(ctx context.Context, email, plain string) (User, error) {
	u, ok, err := s.repo.ByEmail(ctx, email)
	if err != nil {
		return User{}, err
	}
	if !ok {
		// Burn comparable time so unknown-email and wrong-password responses
		// are not distinguishable by latency.
		_ = password.CheckPassword("$2a$10$invalidsaltinvalidsaltinvalidsaltinvalids", plain)
		return User{}, ErrInvalidCredentials
	}
	if err := password.CheckPassword(u.PasswordHash, plain); err != nil {
		return User{}, ErrInvalidCredentials
	}
	return u, nil
}

// ByID returns the account for id.
func (s *Service) ByID(ctx context.Context, id string) (User, error) {
	u, ok, err := s.repo.ByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}
