package snowman

import (
	"context"
	"log"
	"strings"
)

const defaultName = "Snowy"

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
			return nil

		case msg, ok := <-inputs:
			if !ok {
				return nil
			}
			bot.handle(ctx, msg)
		}
	}
}

func (bot *Bot) handle(ctx context.Context, msg Msg) {
	msg.Body = strings.TrimSpace(msg.Body)
	if msg.Body == "" {
		return
	}

	intent, err := bot.cls.Classify(ctx, msg)
	if err != nil {
		bot.handleErr(err)
		return
	}
	intent.Msg = msg

	botMsg, err := bot.proc.Process(ctx, intent)
	if err != nil {
		bot.handleErr(err)
		return
	}
	botMsg.From = bot.name

	if err := bot.ui.Say(ctx, msg.From, botMsg); err != nil {
		bot.handleErr(err)
		return
	}
}

func (bot *Bot) handleErr(err error) {
	if err == nil {
		return
	}
	log.Printf("error: %v", err)
}

// Msg represents a message from the user. Msg can contain additional context
// in terms of attributes that may be used by the intent classifier or action
// & response generation.
type Msg struct {
	From    string                 `json:"from"`
	Body    string                 `json:"body"`
	Attribs map[string]interface{} `json:"attribs"`
}
