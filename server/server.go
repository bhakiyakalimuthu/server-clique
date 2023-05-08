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

func New(logger *zap.Logger, writer io.Writer, queue queue.Queue, store Store, cChan chan *types.Message) (*Server, error) {
	log.SetOutput(writer)
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmicroseconds)
	return &Server{
		logger: logger,
		queue:  queue,
		store:  store,
		cChan:  cChan,
	}, nil
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
			default:
			}
			s.cChan <- msg
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
		var val string
		var ok, isUnknown bool
		switch msg.Action {
		case types.AddItem:
			s.store.Add(ctx, msg.Key, msg.Value, msg.Timestamp)
		case types.RemoveItem:
			ok = s.store.Remove(ctx, msg.Key)
		case types.GetItem:
			val, ok = s.store.Get(ctx, msg.Key)
		case types.GetAll:
			lists := s.store.GetAll(ctx)
			ok = true
			log.Printf("worker id:%d performed action:%s items:%v itemsLength:%d\n", workerID, msg.Action.String(), lists, len(lists))
		default:
			isUnknown = true
		}
		s.output(workerID, msg, val, ok, isUnknown)
	}
}
func (s *Server) output(workerID int, msg *types.Message, result string, ok, isUnknown bool) {
	switch {
	case !ok:
		s.logger.Error("key not found", zap.Int("workerID", workerID), zap.String("action", msg.Action.String()), zap.String("key", msg.Key))
		return
	case isUnknown:
		s.logger.Error("unknown action", zap.Int("workerID", workerID), zap.String("action", msg.Action.String()))
		return
	default:
		log.Printf("worker id:%d performed action:%s key:%s value:%s\n", workerID, msg.Action.String(), msg.Key, result)
	}
}
