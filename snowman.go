package snowman

import (
	"context"
	"strings"
)

const defaultName = "Snowy"

// Run sets up a snowman instance and runs until context is cancelled or the UI
// signals a by closing its listener channel.
func Run(ctx context.Context, opts ...Option) error {
	return New(opts...).Run(ctx)
}

// New returns an instance of Bot with given name and classifier. If the classifier
// is nil, a default dumb classifier will be used.
func New(opts ...Option) *Bot {
	var bot Bot
	for _, opt := range withDefaults(opts) {
		opt(&bot)
	}
	return &bot
}

// Bot represents an instance of the bot. A bot runs continuously blocking the
// host goroutines and generates messages or actions in response to scheduled
// intents or user messages.
type Bot struct {
	name   string
	ui     UI
	cls    Classifier
	proc   Processor
	logger Logger
}

// Run starts all the workers. Run blocks the current goroutine until the ctx
// is cancelled or inputs is closed.
func (bot *Bot) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	inputs, err := bot.ui.Listen(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			bot.process(ctx, Intent{ID: SysIntentShutdown})
			return nil

		case msg, ok := <-inputs:
			if !ok {
				return nil
			}
			intent, err := bot.classify(ctx, msg)
			if err != nil || intent == nil {
				bot.handleErr(err)
				continue
			}
			bot.process(ctx, *intent)
		}
	}
}

func (bot *Bot) classify(ctx context.Context, msg Msg) (*Intent, error) {
	msg.Body = strings.TrimSpace(msg.Body)
	if msg.Body == "" {
		return nil, nil
	}

	intent, err := bot.cls.Classify(ctx, msg)
	if err != nil {
		bot.handleErr(err)
		return nil, err
	}
	intent.Msg = msg

	return &intent, nil
}

func (bot *Bot) process(ctx context.Context, intent Intent) {
	bot.logger.Infof("processing intent: %s", intent.ID)

	botMsg, err := bot.proc.Process(ctx, intent)
	if err != nil {
		bot.handleErr(err)
		return
	}
	botMsg.From = User{ID: bot.name, Name: bot.name}

	if err := bot.ui.Say(ctx, intent.Msg.From, botMsg); err != nil {
		bot.handleErr(err)
		return
	}
}

func (bot *Bot) handleErr(err error) {
	if err == nil {
		return
	}
	bot.logger.Warnf("ignoring error: %v", err)
}

// Msg represents a message from the user. Msg can contain additional context
// in terms of attributes that may be used by the intent classifier or action
// & response generation.
type Msg struct {
	From    User                   `json:"from"`
	Body    string                 `json:"body"`
	Attribs map[string]interface{} `json:"attribs"`
}

// User represents a user that is interacting with snowman.
type User struct {
	ID      string                 `json:"id"`
	Name    string                 `json:"name"`
	Attribs map[string]interface{} `json:"attribs"`
}
