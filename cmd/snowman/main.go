package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/spy16/snowman"
	"github.com/spy16/snowman/regex"
)

var (
	name       = flag.String("name", "Snowy", "Name for the bot")
	slackToken = flag.String("slack", "", "Slack Bot Token")
	intentsDir = flag.String("intents", "./samples", "Intent files directory")
)

func main() {
	flag.Parse()
	logger := logrus.New()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	handler := regex.New(simpleHandler())
	if err := regex.Load(handler, *intentsDir); err != nil {
		log.Fatalf("failed to load intents from '%s': %v", *intentsDir, err)
	}

	snowy := snowman.Bot{
		Logger:  logger,
		Handler: handler,
		Self: snowman.User{
			ID:   *name,
			Name: *name,
		},
		UI: &snowman.SlackUI{
			Token:  *slackToken,
			Logger: logger,
		},
	}
	if err := snowy.Run(ctx); err != nil {
		log.Fatalf("snowy exited: %v", err)
	}
}

func simpleHandler() snowman.Fn {
	return func(msg *snowman.Msg, di snowman.Dialogue) error {
		if len(msg.Intents) == 0 {
			return di.Say(msg.Context(), "I could not understand what you just said 😐")
		}

		return di.Say(msg.Context(), msg.Intents[0].Response)
	}
}
