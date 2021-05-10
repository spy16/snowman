package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/spy16/snowman"
	"github.com/spy16/snowman/slack"
)

var (
	name        = flag.String("name", "Snowy", "Name for the bot")
	noConsole   = flag.Bool("no-console", false, "Disable console UI")
	slackToken  = flag.String("slack", "", "Slack Bot Token")
	intentsFile = flag.String("intents", "", "Intent patterns JSON file")
)

func main() {
	flag.Parse()
	logger := logrus.New()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cls, err := getClassifier(*intentsFile)
	if err != nil {
		logger.Fatalf("failed to init classifier: %v", err)
	}

	opts := []snowman.Option{
		snowman.WithName(*name),
		snowman.WithLogger(logger),
		snowman.WithClassifier(cls),
	}

	if !*noConsole {
		opts = append(opts, snowman.WithUI(&snowman.ConsoleUI{}))
	}

	if *slackToken != "" {
		slackUI := &slack.UI{
			Token:  *slackToken,
			Logger: logger,
		}
		opts = append(opts, snowman.WithUI(slackUI))
	}

	if err := snowman.Run(ctx, opts...); err != nil {
		logger.Errorf("snowy exited with error: %v", err)
	}
}

func getClassifier(intentsFile string) (snowman.Classifier, error) {
	reClassifier := &snowman.RegexClassifier{}
	if intentsFile == "" {
		return reClassifier, nil
	}

	f, err := os.Open(intentsFile)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	if err := json.NewDecoder(f).Decode(&reClassifier); err != nil {
		return nil, err
	}

	return reClassifier, nil
}
