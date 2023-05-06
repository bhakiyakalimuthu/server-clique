package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/bhakiyakalimuthu/server-clique/client"
	"github.com/bhakiyakalimuthu/server-clique/config"
	"github.com/bhakiyakalimuthu/server-clique/queue"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	buildVersion string
	appName      string
)

func main() {
	l := newLogger(buildVersion, appName)
	cfg := config.NewConfig()
	q, err := queue.New(l, cfg.QueueConnString, cfg.QueueName, appName)
	if err != nil {
		l.Fatal("failed to create new queue", zap.Error(err))
	}
	client := client.New(l, q)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
		<-shutdown
		l.Warn("Shutting down client")
		cancel() // cancel the context
	}()
	if err := client.Start(ctx); err != nil {
		l.Fatal("failed to start client", zap.Error(err))
	}
	l.Warn("client sending data completed")
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
