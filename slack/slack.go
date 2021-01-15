package slack

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/slack-go/slack"

	"github.com/spy16/snowman"
)

const (
	msgNoHandler   = "I don't understand what you are saying. :neutral_face:"
	msgInternalErr = ":rotation_light: I am having an internal issue. Please contact developers."

	maxConnectAttempts = 10
)

func New(token string, logger logger) *Slack {
	if logger == nil {
		logger = noOpLogger{}
	}
	return &Slack{
		client: slack.New(token),
		logger: logger,
	}
}

// Slack implements snowman UI using Slack RTM API.
type Slack struct {
	logger
	ctx       context.Context
	cancel    func()
	self      slack.UserDetails
	connected bool
	client    *slack.Client
	rtm       *slack.RTM
}

func (sl *Slack) Listen(ctx context.Context) (<-chan snowman.Msg, error) {
	sl.ctx, sl.cancel = context.WithCancel(context.Background())
	defer sl.cancel()

	rtm := sl.client.NewRTM()
	sl.rtm = rtm
	go rtm.ManageConnection()

	out := make(chan snowman.Msg)
	go sl.listenForEvents(ctx, out)

	return out, nil
}

func (sl *Slack) Say(ctx context.Context, user snowman.User, msg snowman.Msg) error {
	return nil
}

func (sl *Slack) listenForEvents(ctx context.Context, out chan<- snowman.Msg) {
	defer close(out)

	for {
		select {
		case <-ctx.Done():
			return

		case ev := <-sl.rtm.IncomingEvents:
			switch e := ev.Data.(type) {
			case *slack.InvalidAuthEvent:
				sl.Errorf("authentication error, slack UI exiting")
				return
			case *slack.ConnectingEvent:
				sl.Infof("connecting [attempt=%d]...", e.Attempt)
				sl.connected = false
				if e.Attempt >= maxConnectAttempts {
					sl.Errorf("failed to connect even after %d attempts", e.Attempt)
					return
				}

			case *slack.ConnectedEvent:
				sl.connected = true
				sl.self = *e.Info.User
				sl.Infof("connected as '%s' (ID: %s)", sl.self.Name, sl.self.ID)

			default:
				sl.handleEvent(ev, out)
			}
		}
	}
}

func (sl *Slack) handleEvent(re slack.RTMEvent, out chan<- snowman.Msg) {
	sl.Debugf("event: %s [data=%#v]", re.Type, re.Data)

	msgEvent, ok := re.Data.(*slack.MessageEvent)
	if !ok {
		sl.Warnf("ignoring unknown event (type=%v)", reflect.TypeOf(re.Data))
		return
	}

	if msgEvent.User == sl.self.ID || msgEvent.Hidden {
		return
	}
	out <- snowman.Msg{
		From: snowman.User{
			ID:      "",
			Name:    "",
			Attribs: nil,
		},
		Body:    msgEvent.Text,
		Attribs: nil,
	}
}

// Bot represents an RTM connection based Slack Bot instance.
type Bot struct {
	logger

	// configurations
	Handler      Handler
	GroupSupport bool

	// internal state
	ctx       context.Context
	cancel    func()
	self      slack.UserDetails
	connected bool
	client    *slack.Client
	rtm       *slack.RTM
}

// Handler implementation can be set to the bot to handle a message event.
type Handler interface {
	Handle(bot *Bot, msg slack.MessageEvent, from slack.User) error
}

// Listen connects to slack and starts the RTM event connection. Blocks until
// an unrecoverable error occurs.
func (bot *Bot) Listen(ctx context.Context) error {
	bot.ctx, bot.cancel = context.WithCancel(context.Background())
	defer bot.cancel()

	rtm := bot.client.NewRTM()
	bot.rtm = rtm
	go rtm.ManageConnection()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case ev := <-rtm.IncomingEvents:
			if err := bot.handleEvent(ev); err != nil {
				return err
			}
		}
	}
}

// SendMessage sends the text as message to the given channel on behalf of
// the bot instance.
func (bot *Bot) SendMessage(text string, responseTo slack.Msg) error {
	opts := []slack.MsgOption{
		slack.MsgOptionAsUser(true),
		slack.MsgOptionText(text, false),
	}

	if responseTo.ThreadTimestamp != "" {
		opts = append(opts, slack.MsgOptionTS(responseTo.ThreadTimestamp))
	} else if responseTo.Timestamp != "" {
		opts = append(opts, slack.MsgOptionTS(responseTo.Timestamp))
	}

	_, _, err := bot.rtm.PostMessage(responseTo.Channel, opts...)
	return err
}

// Self returns details about the currently connected bot user.
func (bot *Bot) Self() slack.UserDetails { return bot.self }

// Client returns the underlying Slack client instance.
func (bot *Bot) Client() *slack.Client { return bot.client }

func (bot *Bot) handleEvent(re slack.RTMEvent) error {
	bot.Debugf("event: %s [data=%#v]", re.Type, re.Data)

	switch ev := re.Data.(type) {
	case *slack.ConnectingEvent:
		bot.Infof("connecting [attempt=%d]...", ev.Attempt)
		bot.connected = false
		if ev.Attempt >= maxConnectAttempts {
			return fmt.Errorf("failed to connect even after %d attempts", ev.Attempt)
		}

	case *slack.ConnectedEvent:
		bot.connected = true
		bot.self = *ev.Info.User
		bot.Infof("connected as '%s' (ID: %s)", bot.self.Name, bot.self.ID)

	case *slack.InvalidAuthEvent:
		return errors.New("authentication failed")

	case *slack.MessageEvent:
		if ev.User == bot.self.ID || ev.Hidden {
			return nil
		}
		return bot.respond(ev)
	}

	return nil
}

func (bot *Bot) respond(ev *slack.MessageEvent) (err error) {
	defer func() {
		if v := recover(); v != nil {
			bot.Errorf("recovered from panic: %+v", v)
			err = bot.SendMessage(msgInternalErr, ev.Msg)
		}
	}()

	_, err = bot.client.GetGroupInfo(ev.Channel)
	if err == nil && !bot.GroupSupport {
		if !bot.stripAtAddress(ev) {
			return nil
		}
	}

	// TODO: cache user info with TTL instead of fetching everytime.
	user, err := bot.client.GetUserInfo(ev.User)
	if err != nil {
		bot.Errorf("GetUserInfo('%s'): %v", ev.User, err)
		return nil
	}

	if bot.Handler == nil {
		return bot.SendMessage(msgNoHandler, ev.Msg)
	}
	return bot.Handler.Handle(bot, *ev, *user)
}

func (bot *Bot) stripAtAddress(ev *slack.MessageEvent) bool {
	var prefixes = []string{
		AddressUser(bot.self.ID, ""),
		AddressUser(bot.self.ID, bot.self.Name),
	}

	msgText := ev.Msg.Text
	for _, prefix := range prefixes {
		if strings.HasPrefix(ev.Msg.Text, prefix) {
			msgText = strings.TrimSpace(strings.Replace(ev.Msg.Text, prefix, "", -1))
			ev.Text = msgText
			return true
		}
	}

	return false
}

// AddressUser creates the escape sequence for marking a user in a message.
func AddressUser(userID string, userName string) string {
	if userName != "" {
		return fmt.Sprintf("<@%s|%s>:", userID, userName)
	}

	return fmt.Sprintf("<@%s>", userID)
}

type logger interface {
	Debugf(msg string, args ...interface{})
	Infof(msg string, args ...interface{})
	Warnf(msg string, args ...interface{})
	Errorf(msg string, args ...interface{})
}

type noOpLogger struct{}

func (n noOpLogger) Debugf(msg string, args ...interface{}) {}
func (n noOpLogger) Infof(msg string, args ...interface{})  {}
func (n noOpLogger) Warnf(msg string, args ...interface{})  {}
func (n noOpLogger) Errorf(msg string, args ...interface{}) {}
