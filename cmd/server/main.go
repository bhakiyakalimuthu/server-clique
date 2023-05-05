package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bhakiyakalimuthu/server-clique/config"
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

const workerPoolSize = 1

func main() {
	l := newLogger(buildVersion, appName)
	cfg := config.NewConfig()
	q, err := queue.New(l, cfg.QueueConnString, cfg.QueueName, appName)
	if err != nil {
		l.Fatal("failed to create new queue", zap.Error(err))
	}
	s := server.NewMemStore(l)

	cChan := make(chan *types.Message, workerPoolSize)

	server, err := server.New(l, q, s, cfg.OutputFileName, cChan)
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
	cancel()
	// even if cancellation received, current running job will be not be interrupted until it completes
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
