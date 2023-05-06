package server

import (
	"context"
	"time"
)

type Store interface {
	Add(ctx context.Context, key, value string, timestamp time.Time)
	Remove(ctx context.Context, key string) bool
	Get(ctx context.Context, key string) (string, bool)
	GetAll(ctx context.Context) []item
}
