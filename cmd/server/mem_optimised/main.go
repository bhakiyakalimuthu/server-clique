package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bhakiyakalimuthu/server-clique/config"
	"github.com/bhakiyakalimuthu/server-clique/helper"
	"github.com/bhakiyakalimuthu/server-clique/queue"
	"github.com/bhakiyakalimuthu/server-clique/server"
	"github.com/bhakiyakalimuthu/server-clique/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	buildVersion string
	appName      string
)

const workerPoolSize = 5

func main() {
	l := newLogger(appName, buildVersion)
	cfg := config.NewConfig()

	// setup queue
	q, err := queue.New(l, cfg.QueueConnString, cfg.QueueName, appName)
	if err != nil {
		l.Fatal("failed to create new queue", zap.Error(err))
	}

	// setup store
	s := server.NewMemStoreOptimised(l)

	cChan := make(chan *types.Message, workerPoolSize)

	// output file writer
	f, err := os.OpenFile(cfg.OutputFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o664)
	if err != nil {
		l.Fatal("failed to open output file", zap.Error(err))
	}

	// file server for viewing output.json file, This is just a helper
	fileServer := helper.NewFileServer(l, cfg.OutputFileName, cfg.FileServerListenAddress)
	go fileServer.Start()

	server, err := server.New(l, f, q, s, cChan)
	if err != nil {
		l.Fatal("failed to start server", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)
	wg.Add(workerPoolSize)

	shutdown := make(chan os.Signal, 1)
	go func() {
		if err := server.Start(ctx); err != nil {
			// queue consume failed,exit the program
			shutdown <- syscall.SIGQUIT
		}
	}()

	// start worker and add worker pool
	for i := 0; i < workerPoolSize; i++ {
		go server.Process(ctx, wg)
	}

	signal.Notify(shutdown, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	// handle shut down
	<-shutdown
	l.Warn("Shutting down server")

	fileServer.Stop() // stop the file server
	cancel()          // cancel the context
	// even if cancellation received, current running job will not be interrupted until it completes
	// wait for all the workers to be completed
	wg.Wait()
}

func newLogger(appName, version string) *zap.Logger {
	logLevel := zap.DebugLevel
	var zapCore zapcore.Core
	level := zap.NewAtomicLevel()
	level.SetLevel(logLevel)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoder := zapcore.NewJSONEncoder(encoderCfg)
	zapCore = zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), level)

	logger := zap.New(zapCore, zap.AddCaller(), zap.ErrorOutput(zapcore.Lock(os.Stderr)))
	logger = logger.With(zap.String("app", appName), zap.String("buildVersion", version))
	return logger
}
