package fileapi

import "context"

type userIDContextKey struct{}

// WithUserID attaches an uploader identity to context for file transport use.
func WithUserID(ctx context.Context, userID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

// UserIDFromContext extracts the uploader identity from context.
func UserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	userID, _ := ctx.Value(userIDContextKey{}).(string)
	return userID
}
