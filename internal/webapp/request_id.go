package webapp

import (
	"context"
	"github.com/google/uuid"
)

type requestIdCtxKey int

const requestIdKey requestIdCtxKey = 1

func NewRequestId() uuid.UUID {
	return uuid.New()
}

func WithRequestId(ctx context.Context, requestId uuid.UUID) context.Context {
	return context.WithValue(ctx, requestIdKey, requestId)
}

func RequestIdFromContext(ctx context.Context) uuid.UUID {
	requestId := ctx.Value(requestIdKey).(uuid.UUID)

	return requestId
}
