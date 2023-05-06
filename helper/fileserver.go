package helper

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"
)

type FileServer struct {
	logger   *zap.Logger
	fileName string
	server   *http.Server
}

func NewFileServer(logger *zap.Logger, fileName string) *FileServer {
	return &FileServer{
		logger:   logger,
		fileName: fileName,
	}
}

func (f *FileServer) Start() {
	path := "./"
	f.server = &http.Server{
		Addr:    "localhost:8080",
		Handler: http.FileServer(http.Dir(path)),
	}
	f.logger.Info("Starting file server...", zap.String("listeningAddress", "localhost:8080"), zap.String("path", path))
	err := f.server.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		return
	}
	f.logger.Error("Failed to start file server", zap.Error(err))
}

func (f *FileServer) Stop() {
	if f.server != nil {
		_ = f.server.Shutdown(context.Background())
	}
}
