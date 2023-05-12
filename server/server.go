package server

import (
	"context"
	"io"
	"log"
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
	cChan  chan *types.Message // consumer channel
}

func New(logger *zap.Logger, writer io.Writer, queue queue.Queue, store Store, cChan chan *types.Message) *Server {
	log.SetOutput(writer)
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmicroseconds)
	return &Server{
		logger: logger,
		queue:  queue,
		store:  store,
		cChan:  cChan,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Init server and waiting.......")
	pChan, err := s.queue.Consume(ctx)
	if err != nil {
		s.logger.Error("failed to consume message", zap.Error(err))
		return err
	}
	defer func() {
		close(s.cChan)  // close consumer channel
		s.file.Close()  // close opened file
		s.queue.Close() // gracefully close the connection
	}()
	for {
		select {
		case msg := <-pChan:
			select {
			case <-ctx.Done():
				return nil
			case s.cChan <- msg:
			default:
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Server) Process(ctx context.Context, wg *sync.WaitGroup, workerID int) {
	defer func() {
		s.logger.Warn("worker exiting!!!", zap.Int("workerID", workerID))
		wg.Done()
	}()
	for msg := range s.cChan {
		if msg == nil {
			// handle edge case, when the connection is closed nil might get passed
			s.logger.Debug("received nil msg value")
			return
		}
		switch msg.Action {
		case types.AddItem:
			s.store.Add(ctx, msg.Key, msg.Value, msg.Timestamp)
			log.Printf("worker id:%d performed action:%s key:%s value:%s\n", workerID, msg.Action.String(), msg.Key, msg.Value)
		case types.RemoveItem:
			if ok := s.store.Remove(ctx, msg.Key); !ok {
				s.logger.Error("key not found", zap.Int("workerID", workerID), zap.String("action", msg.Action.String()), zap.String("key", msg.Key))
				continue
			}
			log.Printf("worker id:%d performed action:%s key:%s\n", workerID, msg.Action.String(), msg.Key)
		case types.GetItem:
			val, ok := s.store.Get(ctx, msg.Key)
			if !ok {
				s.logger.Error("key not found", zap.Int("workerID", workerID), zap.String("action", msg.Action.String()), zap.String("key", msg.Key))
				continue
			}
			log.Printf("worker id:%d performed action:%s key:%s value:%s\n", workerID, msg.Action.String(), msg.Key, val)
		case types.GetAll:
			lists := s.store.GetAll(ctx)
			log.Printf("worker id:%d performed action:%s items:%v itemsLength:%d\n", workerID, msg.Action.String(), lists, len(lists))
		default:
			s.logger.Error("unknown action", zap.Int("workerID", workerID), zap.String("action", msg.Action.String()))
		}
	}
}
