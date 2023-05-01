package client

import (
	"context"

	"github.com/bhakiyakalimuthu/server-clique/queue"
	"go.uber.org/zap"
)

type Client struct {
	logger *zap.Logger
	queue  queue.Queue
}

func NewClient(logger *zap.Logger, queue queue.Queue) *Client {
	return &Client{
		logger: logger,
		queue:  queue,
	}
}

func (c *Client) Start(ctx context.Context) {

}
