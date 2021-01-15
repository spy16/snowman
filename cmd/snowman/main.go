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

	reClassifier := &snowman.RegexClassifier{}
	_ = reClassifier.Register("my name is (?P<name>.*)", "user.greets")

	if err := snowman.Run(ctx,
		snowman.WithName(*name),
		snowman.WithLogger(logger),
		snowman.WithUI(snowman.ConsoleUI{}),
		snowman.WithClassifier(reClassifier),
		snowman.WithProcessor(snowman.ProcessorFunc(func(ctx context.Context, intent snowman.Intent) (snowman.Msg, error) {
			switch intent.ID {
			case "user.greets":
				return snowman.Msg{
					Body: "👋 Hello " + intent.Ctx["name"].(string),
				}, nil
			default:
				return snowman.Msg{Body: "I don't understand 😐"}, nil
			}
		})),
	); err != nil {
		logger.Errorf("snowy exited with error: %v", err)
	}
}

func cancelOnInterrupt(cancel context.CancelFunc, logger snowman.Logger) {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logger.Infof("terminating (signal: %v)", sig)
	cancel()
}
