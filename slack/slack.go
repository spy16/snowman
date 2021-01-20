package slack

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/slack-go/slack"

	"github.com/spy16/snowman"
)

const maxConnectAttempts = 10

func New(token string, logger logger) *Slack {
	if logger == nil {
		logger = snowman.NoOpLogger{}
	}
	return &Slack{
		client: slack.New(token),
		logger: logger,
	}
}

// Slack implements snowman UI using Slack RTM API.
type Slack struct {
	logger

	GroupSupport bool

	ctx       context.Context
	cancel    func()
	self      slack.UserDetails
	connected bool
	client    *slack.Client
	rtm       *slack.RTM
}

// Listen starts an RTM connection and starts listening for events. Message events
// are pushed to the returned channel.
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
				sl.handleEvent(ctx, ev, out)
			}
		}
	}
}

func (sl *Slack) handleEvent(ctx context.Context, re slack.RTMEvent, out chan<- snowman.Msg) {
	sl.Debugf("event: %s [data=%#v]", re.Type, re.Data)

	ev, ok := re.Data.(*slack.MessageEvent)
	if !ok {
		sl.Warnf("ignoring unknown event (type=%v)", reflect.TypeOf(re.Data))
		return
	}

	_, err := sl.client.GetGroupInfo(ev.Channel)
	if err == nil && !sl.GroupSupport {
		if !sl.stripAtAddress(ev) {
			return
		}
	}

	user, err := sl.client.GetUserInfo(ev.User)
	if err != nil {
		sl.Errorf("GetUserInfo('%s'): %v", ev.User, err)
		return
	}

	if ev.User == sl.self.ID || ev.Hidden {
		return
	}

	snowMsg := snowman.Msg{
		From: snowman.User{
			ID:   user.ID,
			Name: user.RealName,
		},
		Body: ev.Text,
		Attribs: map[string]interface{}{
			"slack_msg":  ev.Msg,
			"slack_user": *user,
		},
	}

	select {
	case <-ctx.Done():
		return
	case out <- snowMsg:
	}
}

func (sl *Slack) stripAtAddress(ev *slack.MessageEvent) bool {
	var prefixes = []string{
		AddressUser(sl.self.ID, ""),
		AddressUser(sl.self.ID, sl.self.Name),
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

// SendMessage sends the text as message to the given channel on behalf of
// the bot instance.
func (sl *Slack) SendMessage(text string, responseTo slack.Msg) error {
	opts := []slack.MsgOption{
		slack.MsgOptionAsUser(true),
		slack.MsgOptionText(text, false),
	}

	if responseTo.ThreadTimestamp != "" {
		opts = append(opts, slack.MsgOptionTS(responseTo.ThreadTimestamp))
	} else if responseTo.Timestamp != "" {
		opts = append(opts, slack.MsgOptionTS(responseTo.Timestamp))
	}

	_, _, err := sl.rtm.PostMessage(responseTo.Channel, opts...)
	return err
}

// Self returns details about the currently connected bot user.
func (sl *Slack) Self() slack.UserDetails { return sl.self }

// Client returns the underlying Slack client instance.
func (sl *Slack) Client() *slack.Client { return sl.client }

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

func (n noOpLogger) Debugf(string, ...interface{}) {}
func (n noOpLogger) Infof(string, ...interface{})  {}
func (n noOpLogger) Warnf(string, ...interface{})  {}
func (n noOpLogger) Errorf(string, ...interface{}) {}
