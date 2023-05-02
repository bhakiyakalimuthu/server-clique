package server

import (
	"context"
)

type Store interface {
	Add(ctx context.Context, key, value string)
	Remove(ctx context.Context, key string) bool
	Get(ctx context.Context, key string) (string, bool)
	GetAll(ctx context.Context) []pair
}
