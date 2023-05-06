package queue

import (
	"context"

	"github.com/bhakiyakalimuthu/server-clique/types"
)

// Queue is an abstract type, can be extended by different message queue implementations
type Queue interface {
	Publish(*types.Message) error
	Consume(context.Context) (<-chan *types.Message, error)
	Close() error
}
