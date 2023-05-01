package queue

import (
	"context"

	"github.com/bhakiyakalimuthu/server-clique/types"
)

type Queue interface {
	Publish(*types.Message) error
	Consume(context.Context) (<-chan *types.Message, error)
}
