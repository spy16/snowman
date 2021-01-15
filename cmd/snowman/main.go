package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/spy16/snowman"
)

var name = flag.String("name", "Snowy", "")

func main() {
	flag.Parse()

	logger := logrus.New()

	ctx, cancel := context.WithCancel(context.Background())
	go cancelOnInterrupt(cancel, logger)

	bot := snowman.New(
		snowman.WithName(*name),
	)
	if err := bot.Run(ctx); err != nil {
		logger.Errorf("bot exited with error: %v", err)
	}
}

func cancelOnInterrupt(cancel context.CancelFunc, logger snowman.Logger) {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logger.Infof("terminating (signal: %v)", sig)
	cancel()
}
