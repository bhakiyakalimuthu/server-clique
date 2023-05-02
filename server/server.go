package server

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/bhakiyakalimuthu/server-clique/queue"
	"github.com/bhakiyakalimuthu/server-clique/types"
	"go.uber.org/zap"
)

type Server struct {
	logger *zap.Logger
	queue  queue.Queue
	store  Store
	file   *os.File
	cChan  chan *types.Message
}

func New(logger *zap.Logger, queue queue.Queue, store Store, fileName string, cChan chan *types.Message) (*Server, error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o664)
	if err != nil {
		return nil, err
	}
	return &Server{
		logger: logger,
		queue:  queue,
		store:  store,
		file:   file,
		cChan:  cChan,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	pChan, err := s.queue.Consume(ctx)
	if err != nil {
		s.logger.Error("failed to consume message", zap.Error(err))
		return err
	}
	defer func() {
		close(s.cChan)
		s.file.Close()
	}()
	for {
		select {
		case msg := <-pChan:
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			s.cChan <- msg
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Server) Process(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for msg := range s.cChan {
		var str string
		switch msg.Action {
		case types.AddItem:
			s.store.Add(ctx, msg.Key, msg.Value)
			str = fmt.Sprintf("performed action:%s key:%s value:%s\n", msg.Action.String(), msg.Key, msg.Value)
		case types.RemoveItem:
			if ok := s.store.Remove(ctx, msg.Key); !ok {
				s.logger.Error("key not found", zap.String("action", msg.Action.String()), zap.String("key", msg.Key))
				continue
			}
			str = fmt.Sprintf("performed action:%s key:%s\n", msg.Action.String(), msg.Key)
		case types.GetItem:
			val, ok := s.store.Get(ctx, msg.Key)
			if !ok {
				s.logger.Error("key not found", zap.String("action", msg.Action.String()), zap.String("key", msg.Key))
				continue
			}
			str = fmt.Sprintf("performed action:%s key:%s value:%s\n", msg.Action.String(), msg.Key, val)
		case types.GetAll:
			lists := s.store.GetAll(ctx)
			str = fmt.Sprintf("performed action:%s items:%v itemsLength:%d\n", msg.Action.String(), lists, len(lists))
		default:
			s.logger.Error("unknown action", zap.String("action", msg.Action.String()))
		}
		fmt.Fprint(s.file, str)
	}
}
