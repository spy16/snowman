package snowman

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"time"
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
	self   User
	cls    Classifier
	uis    []UI
	logger Logger

	diLock    sync.RWMutex
	dialogues map[string]*Dialogue
}

// Run starts all the workers. Run blocks the current goroutine until the ctx
// is cancelled or inputs is closed.
func (bot *Bot) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if len(bot.uis) == 0 {
		bot.uis = append(bot.uis, &ConsoleUI{Prompt: "user >"})
	}

	wg := &sync.WaitGroup{}
	for _, ui := range bot.uis {
		wg.Add(1)
		go func(ui UI) {
			defer wg.Done()

			err := ui.Listen(ctx, func(msg Msg) { bot.handle(ctx, msg, ui) })
			if err != nil {
				log.Printf("conversation over '%s' exited with error: %v", reflect.TypeOf(ui), err)
			}
		}(ui)
	}
	wg.Wait()

	return nil
}

func (bot *Bot) handle(ctx context.Context, msg Msg, ui UI) {
	di := bot.allocDialogue(msg.From)

	intents, err := bot.cls.Classify(ctx, *di, msg)
	if err != nil {
		bot.handleErr("Classifier.Classify", err)
		return
	}

	resp, err := bot.execute(ctx, *di, msg, intents)
	if err != nil {
		bot.handleErr("Bot.execute", err)
		return
	} else if resp == nil {
		return // nothing to say.
	}
	resp.From = bot.self
	if resp.To.ID == "" {
		resp.To = msg.From
	}

	if err := ui.Say(context.Background(), *resp); err != nil {
		bot.handleErr("UI.Say", err)
	}
}

func (bot *Bot) execute(ctx context.Context, di Dialogue, msg Msg, intents []Intent) (*Msg, error) {
	sort.SliceStable(intents, func(i, j int) bool {
		return intents[i].Confidence < intents[j].Confidence
	})

	if len(intents) == 0 || intents[0].ID == "sys.dumb" {
		return &Msg{
			At:   time.Now(),
			To:   msg.From,
			From: msg.To,
			Body: sysDumbResponse(),
		}, nil
	}

	return &Msg{
		At:   time.Now(),
		To:   msg.From,
		From: msg.To,
		Body: intents[0].String(),
	}, nil
}

func (bot *Bot) allocDialogue(with User) *Dialogue {
	bot.diLock.Lock()
	defer bot.diLock.Unlock()

	if _, exists := bot.dialogues[with.String()]; !exists {
		if bot.dialogues == nil {
			bot.dialogues = map[string]*Dialogue{}
		}

		bot.dialogues[with.String()] = &Dialogue{
			Ctx:  nil,
			with: with,
		}
	}

	return bot.dialogues[with.String()]
}

func (bot *Bot) handleErr(op string, err error) {
	if err == nil {
		return
	}
	bot.logger.Warnf("ignoring error during '%s': %v", op, err)
}

// Msg represents a message from the user. Msg can contain additional context
// in terms of attributes that may be used by the intent classifier or action
// & response generation.
type Msg struct {
	At   time.Time `json:"at"`
	To   User      `json:"to"`
	From User      `json:"from"`
	Body string    `json:"body"`
}

func (m Msg) String() string { return fmt.Sprintf("Msg<to=@%s,from=@%s>", m.To.ID, m.From.ID) }

// User represents a user that is interacting with snowman.
type User struct {
	ID      string                 `json:"id"`
	Name    string                 `json:"name"`
	Attribs map[string]interface{} `json:"attribs"`
}

func (u User) String() string { return fmt.Sprintf("User<ID=@%s>", u.ID) }

// Dialogue represents a conversation between the bot and a user and holds
// conversational context.
type Dialogue struct {
	Ctx  map[string]interface{}
	with User
}

func sysDumbResponse() string {
	var responses = []string{
		"I don't understand what you are saying 😐",
		"Please rephrase what you just said 🙏",
		"I am not able to understand what you just said 😔",
	}

	return responses[rand.Intn(len(responses))]
}
