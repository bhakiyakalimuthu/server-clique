package server

import (
	"context"

	"github.com/bhakiyakalimuthu/server-clique/types"
)

type Store interface {
	Add(ctx context.Context, key, value string)
	Remove(ctx context.Context, key string) bool
	Get(ctx context.Context, key string) (string, bool)
	GetAll(ctx context.Context) []types.Payload
}
