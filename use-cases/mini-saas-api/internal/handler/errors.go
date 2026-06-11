package handler

import (
	"errors"
	"net/http"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"mini-saas-api/internal/domain/project"
	"mini-saas-api/internal/domain/session"
	"mini-saas-api/internal/domain/tenantspace"
	"mini-saas-api/internal/domain/user"
)

// writeDomainError maps domain sentinel errors to the canonical contract error
// envelope. Unknown errors become 500 without leaking internals.
func writeDomainError(w http.ResponseWriter, r *http.Request, logger plumelog.StructuredLogger, err error) {
	eb := contract.NewErrorBuilder()
	switch {
	case errors.Is(err, user.ErrInvalidCredentials):
		eb = eb.Type(contract.TypeUnauthorized).Code("auth.invalid_credentials").Message("invalid email or password")
	case errors.Is(err, user.ErrEmailTaken):
		eb = eb.Type(contract.TypeAlreadyExists).Code("user.email_taken").Message("email is already registered")
	case errors.Is(err, user.ErrWeakPassword):
		eb = eb.Type(contract.TypeValidation).Code("user.weak_password").Message("password must be at least 8 characters with upper, lower, and digit")
	case errors.Is(err, user.ErrNotFound):
		eb = eb.Type(contract.TypeNotFound).Code("user.not_found").Message("user not found")
	case errors.Is(err, session.ErrReused):
		eb = eb.Type(contract.TypeUnauthorized).Code("auth.refresh_reused").Message("refresh token reuse detected; session revoked, log in again")
	case errors.Is(err, session.ErrInvalidToken):
		eb = eb.Type(contract.TypeUnauthorized).Code("auth.refresh_invalid").Message("invalid or expired refresh token")
	case errors.Is(err, tenantspace.ErrSlugTaken):
		eb = eb.Type(contract.TypeAlreadyExists).Code("tenant.slug_taken").Message("workspace slug is already taken")
	case errors.Is(err, tenantspace.ErrInvalidSlug):
		eb = eb.Type(contract.TypeValidation).Code("tenant.invalid_slug").Message("slug must be 3-64 lowercase letters, digits, or hyphens")
	case errors.Is(err, tenantspace.ErrAlreadyMember):
		eb = eb.Type(contract.TypeAlreadyExists).Code("tenant.already_member").Message("user is already a member of this workspace")
	case errors.Is(err, tenantspace.ErrLastOwner):
		eb = eb.Type(contract.TypeConflict).Code("tenant.last_owner").Message("cannot demote or remove the last owner")
	case errors.Is(err, tenantspace.ErrInvalidRole):
		eb = eb.Type(contract.TypeValidation).Code("tenant.invalid_role").Message("role must be owner, admin, or member")
	case errors.Is(err, tenantspace.ErrNotFound):
		eb = eb.Type(contract.TypeNotFound).Code("tenant.not_found").Message("not found")
	case errors.Is(err, project.ErrNameRequired):
		eb = eb.Type(contract.TypeRequired).Code("project.name_required").Message("project name is required")
	case errors.Is(err, project.ErrInvalidStatus):
		eb = eb.Type(contract.TypeValidation).Code("project.invalid_status").Message("status must be active, paused, or archived")
	case errors.Is(err, project.ErrLimitReached):
		eb = eb.Type(contract.TypeRateLimited).Code("project.limit_reached").Message("plan project limit reached")
	case errors.Is(err, project.ErrNotFound):
		eb = eb.Type(contract.TypeNotFound).Code("project.not_found").Message("not found")
	default:
		if logger != nil {
			logger.Error("internal error", plumelog.Fields{"error": err.Error()})
		}
		eb = eb.Type(contract.TypeInternal).Code("internal").Message("internal error")
	}
	logWriteErr(logger, contract.WriteError(w, r, eb.Build()))
}

// writeBadJSON responds 400 for undecodable request bodies.
func writeBadJSON(w http.ResponseWriter, r *http.Request, logger plumelog.StructuredLogger) {
	logWriteErr(logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeBadRequest).
		Code("request.invalid_json").
		Message("request body must be valid JSON").
		Build()))
}
