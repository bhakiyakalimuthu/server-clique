package client

import (
	"context"
	_ "embed"
	"encoding/json"

	"github.com/bhakiyakalimuthu/server-clique/queue"
	"github.com/bhakiyakalimuthu/server-clique/types"
	"go.uber.org/zap"
)

//go:embed input.json
var inputBytes []byte

type Client struct {
	logger *zap.Logger
	queue  queue.Queue
}

func New(logger *zap.Logger, queue queue.Queue) *Client {
	return &Client{
		logger: logger,
		queue:  queue,
	}
}

func (c *Client) Start(ctx context.Context) error {
	items := make([]*types.Message, 0)
	if err := json.Unmarshal(inputBytes, &items); err != nil {
		return err
	}
	for _, msg := range items {
		select {
		case <-ctx.Done():
			c.logger.Warn("context cancelled, publish operation aborted")
			return ctx.Err()
		default:
			if err := c.queue.Publish(msg); err != nil {
				c.logger.Error("failed to publish message", zap.Any("msg", msg), zap.Error(err))
				continue
			}
		}
	}
	return nil
}
