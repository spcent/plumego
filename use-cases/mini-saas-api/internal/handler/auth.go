package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"mini-saas-api/internal/domain/access"
	"mini-saas-api/internal/domain/audit"
	"mini-saas-api/internal/domain/session"
	"mini-saas-api/internal/domain/tenantspace"
	"mini-saas-api/internal/domain/user"
)

// UserService is the account dependency of the handlers.
type UserService interface {
	Register(ctx context.Context, email, name, password string) (user.User, error)
	Authenticate(ctx context.Context, email, password string) (user.User, error)
	ByID(ctx context.Context, id string) (user.User, error)
	ByEmail(ctx context.Context, email string) (user.User, error)
}

// WorkspaceService is the tenant dependency of the handlers.
// tenantspace.Service satisfies it.
type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, name, slug, ownerUserID string) (tenantspace.Tenant, tenantspace.Membership, error)
	Get(ctx context.Context, tenantID string) (tenantspace.Tenant, error)
	Update(ctx context.Context, tenantID, name, plan string) (tenantspace.Tenant, error)
	Members(ctx context.Context, tenantID string) ([]tenantspace.Membership, error)
	AddMember(ctx context.Context, tenantID, userID string, role access.Role) (tenantspace.Membership, error)
	ChangeRole(ctx context.Context, tenantID, membershipID string, role access.Role) (tenantspace.Membership, error)
	RemoveMember(ctx context.Context, tenantID, membershipID string) error
	PrimaryMembership(ctx context.Context, userID string) (tenantspace.Membership, error)
	MembershipForUser(ctx context.Context, tenantID, userID string) (tenantspace.Membership, error)
	MembershipBySlug(ctx context.Context, slug, userID string) (tenantspace.Membership, error)
}

// TokenIssuer is the token dependency of AuthHandler.
type TokenIssuer interface {
	IssueTokens(ctx context.Context, userID, tenantID string, role access.Role) (session.TokenSet, error)
	IssueAccess(ctx context.Context, userID, tenantID string, role access.Role) (string, int64, error)
	RotateRefresh(ctx context.Context, token string) (session.Record, string, error)
}

// AuditRecorder appends audit entries; failures must not block the request.
type AuditRecorder interface {
	Record(ctx context.Context, e audit.Entry) error
}

// AuthHandler serves signup, login, and refresh.
type AuthHandler struct {
	Users  UserService
	Spaces WorkspaceService
	Tokens TokenIssuer
	Audit  AuditRecorder
	Logger plumelog.StructuredLogger
}

type signupRequest struct {
	Email         string `json:"email"`
	Name          string `json:"name"`
	Password      string `json:"password"`
	WorkspaceName string `json:"workspace_name"`
	WorkspaceSlug string `json:"workspace_slug"`
}

type authResponse struct {
	User   user.User           `json:"user"`
	Tenant *tenantspace.Tenant `json:"tenant,omitempty"`
	Tokens session.TokenSet    `json:"tokens"`
}

// Signup creates an account plus its workspace and returns a token pair.
//
//	POST /api/v1/auth/signup → 201 {user, tenant, tokens}
func (h AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadJSON(w, r, h.Logger)
		return
	}
	u, err := h.Users.Register(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	tenant, membership, err := h.Spaces.CreateWorkspace(r.Context(), req.WorkspaceName, req.WorkspaceSlug, u.ID)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	tokens, err := h.Tokens.IssueTokens(r.Context(), u.ID, tenant.ID, membership.Role)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	recordAudit(r.Context(), h.Audit, h.Logger, tenant.ID, u.ID, "tenant.created", "tenant", tenant.ID, "workspace "+tenant.Slug+" created")
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusCreated, authResponse{
		User:   u,
		Tenant: &tenant,
		Tokens: tokens,
	}, nil))
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	// WorkspaceSlug optionally selects which workspace to log into; the
	// default is the user's oldest membership.
	WorkspaceSlug string `json:"workspace_slug"`
}

// Login verifies credentials and returns a token pair bound to the selected
// workspace (workspace_slug) or the user's primary (oldest) one.
//
//	POST /api/v1/auth/login → 200 {user, tokens}
func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadJSON(w, r, h.Logger)
		return
	}
	u, err := h.Users.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	var membership tenantspace.Membership
	if req.WorkspaceSlug != "" {
		membership, err = h.Spaces.MembershipBySlug(r.Context(), req.WorkspaceSlug, u.ID)
	} else {
		membership, err = h.Spaces.PrimaryMembership(r.Context(), u.ID)
	}
	if err != nil {
		// No usable membership: treat as invalid credentials rather than
		// leaking workspace existence or membership state.
		writeDomainError(w, r, h.Logger, user.ErrInvalidCredentials)
		return
	}
	tokens, err := h.Tokens.IssueTokens(r.Context(), u.ID, membership.TenantID, membership.Role)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	recordAudit(r.Context(), h.Audit, h.Logger, membership.TenantID, u.ID, "auth.login", "user", u.ID, "")
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, authResponse{User: u, Tokens: tokens}, nil))
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh rotates a refresh token and returns a fresh token pair. The role is
// re-read from the membership store so role changes propagate on refresh.
//
//	POST /api/v1/auth/refresh → 200 {tokens}
func (h AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadJSON(w, r, h.Logger)
		return
	}
	rec, nextRefresh, err := h.Tokens.RotateRefresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	membership, err := h.Spaces.MembershipForUser(r.Context(), rec.TenantID, rec.UserID)
	if err != nil {
		// Member was removed since the refresh token was issued: fail closed.
		writeDomainError(w, r, h.Logger, session.ErrInvalidToken)
		return
	}
	accessToken, expiresIn, err := h.Tokens.IssueAccess(r.Context(), rec.UserID, rec.TenantID, membership.Role)
	if err != nil {
		writeDomainError(w, r, h.Logger, err)
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, struct {
		Tokens session.TokenSet `json:"tokens"`
	}{session.TokenSet{
		AccessToken:  accessToken,
		RefreshToken: nextRefresh,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	}}, nil))
}
