package snowman

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// Handler represents the message Handler that can process the message
// to generate response or to append additional context to the message.
type Handler interface {
	Handle(msg *Msg, di Dialogue) error
}

// UI represents the interface used for interactions between the bot and
// the user.
type UI interface {
	// Say should display the message to the user as if the bot said it.
	// Target user that should receive the message should be identified
	// using the 'msg.To' value.
	Say(ctx context.Context, msg Msg) error

	// Listen should listen for messages from user and deliver them to the
	// Handler. Listen should block until UI reaches a terminal state or
	// until context is cancelled.
	Listen(ctx context.Context, receive func(msg Msg)) error
}

// Dialogue holds the conversational context of bot with a specific user.
type Dialogue interface {
	ID() string
	Self() User
	Say(ctx context.Context, body string) error
}

// Bot represents an instance of the bot. A bot runs continuously blocking the
// host goroutine and generates messages or actions in response to scheduled
// intents or user messages.
type Bot struct {
	UI      UI
	Self    User
	Logger  Logger
	Handler Handler

	diLock    sync.RWMutex
	dialogues map[string]*dialogueCtx
}

// Run starts all the workers. Run blocks the current goroutine until the ctx
// is cancelled or inputs is closed.
func (bot *Bot) Run(ctx context.Context) error {
	bot.init()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return bot.UI.Listen(ctx, func(msg Msg) {
		di := bot.allocDialogue(msg)
		if msg.ctx == nil {
			msg.ctx = ctx
		}

		if err := bot.Handler.Handle(&msg, di); err != nil {
			bot.Logger.Errorf("handler error for message from '%s': %v", msg.From, err)
		}
	})
}

func (bot *Bot) allocDialogue(msg Msg) *dialogueCtx {
	bot.diLock.Lock()
	defer bot.diLock.Unlock()

	if _, found := bot.dialogues[msg.From.String()]; !found {
		if bot.dialogues == nil {
			bot.dialogues = map[string]*dialogueCtx{}
		}
		bot.dialogues[msg.From.String()] = &dialogueCtx{
			ui:   bot.UI,
			self: bot.Self,
			with: msg.From,
		}
	}

	di := bot.dialogues[msg.From.String()]
	di.with = msg.From
	return di
}

func (bot *Bot) init() {
	if bot.UI == nil {
		bot.UI = &ConsoleUI{}
	}

	if bot.Logger == nil {
		bot.Logger = &StdLogger{}
	}

	if bot.Self.ID == "" {
		bot.Self = User{
			ID:      "snowy",
			Name:    "Snowy",
			Attribs: nil,
		}
	}

	if bot.Handler == nil {
		bot.Handler = Fn(func(msg *Msg, di Dialogue) error {
			return di.Say(msg.Context(), dumbResponse())
		})
	}
}

func dumbResponse() string {
	var responses = []string{
		"I don't understand what you are saying 😐",
		"Please rephrase what you just said 🙏",
		"I am not able to understand what you just said 😔",
	}

	return responses[rand.Intn(len(responses))]
}

// Fn is an adaptor to implement Handler using simple Go func values.
type Fn func(msg *Msg, di Dialogue) error

// Handle simply dispatches the args to the wrapped function.
func (fn Fn) Handle(msg *Msg, di Dialogue) error { return fn(msg, di) }

// dialogueCtx represents the context of a dialogue.
type dialogueCtx struct {
	ui   UI
	self User
	with User
}

func (di *dialogueCtx) ID() string { return di.with.String() }
func (di *dialogueCtx) Self() User { return di.self }
func (di *dialogueCtx) Say(ctx context.Context, body string) error {
	return di.ui.Say(ctx, Msg{
		At:   time.Now(),
		To:   di.with,
		From: di.self,
		Body: body,
	})
}
