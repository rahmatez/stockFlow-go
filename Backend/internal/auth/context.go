package auth

import (
	"context"

	"github.com/google/uuid"
)

const TenantContextKey contextKey = "tenant_id"

func WithTenantID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, TenantContextKey, id)
}

func TenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(TenantContextKey).(uuid.UUID)
	return id, ok && id != uuid.Nil
}
