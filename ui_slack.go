package snowman

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

var _ UI = (*SlackUI)(nil)

const maxConnectAttempts = 5

// SlackUI implements snowman SlackUI using Slack RTM based API.
type SlackUI struct {
	Logger

	Token        string
	Options      []slack.Option
	EnableGroups bool

	client    *slack.Client
	slRTM     *slack.RTM
	self      slack.UserDetails
	connected bool
}

// Say sends a message to the user/channel on Slack identified using the UserID
// in the msg.To field.
func (sui *SlackUI) Say(ctx context.Context, msg Msg) error {
	channel, ok := msg.To.Attribs["slack_channel"].(string)
	if !ok {
		return errors.New("slack_channel attrib missing")
	}

	ts, ok := msg.To.Attribs["slack_ts"].(string)
	if !ok {
		return errors.New("slack_ts attrib missing")
	}

	_, _, err := sui.slRTM.PostMessageContext(ctx, channel, []slack.MsgOption{
		slack.MsgOptionAsUser(true),
		slack.MsgOptionText(msg.Body, false),
		slack.MsgOptionTS(ts),
	}...)
	return err
}

func (sui *SlackUI) Listen(ctx context.Context, handle func(msg Msg)) error {
	sui.client = slack.New(sui.Token, sui.Options...)
	sui.slRTM = sui.client.NewRTM()
	go sui.slRTM.ManageConnection()

	if sui.Logger == nil {
		sui.Logger = NoOpLogger{}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case ev, more := <-sui.slRTM.IncomingEvents:
			if !more {
				return nil
			}

			switch e := ev.Data.(type) {
			case *slack.InvalidAuthEvent:
				return errors.New("authentication error")

			case *slack.ConnectingEvent:
				sui.Infof("connecting [attempt=%d]...", e.Attempt)
				sui.connected = false
				if e.Attempt >= maxConnectAttempts {
					return fmt.Errorf("failed to connect even after %d attempts", e.Attempt)
				}

			case *slack.ConnectedEvent:
				sui.connected = true
				sui.self = *e.Info.User
				sui.Infof("connected as '%s' (ID: %s)", sui.self.Name, sui.self.ID)

			case *slack.MessageEvent:
				if e.Hidden {
					continue
				} else if e.User == sui.self.ID {
					continue
				}

				sui.handleMessageEvent(e, handle)

			default:
				sui.Debugf("unhandled event: %v", ev)
			}
		}
	}
}

func (sui *SlackUI) handleMessageEvent(e *slack.MessageEvent, handle func(msg Msg)) {
	from := User{
		ID:   e.User,
		Name: e.Username,
		Attribs: map[string]interface{}{
			"slack_channel": e.Channel,
		},
	}

	if e.ThreadTimestamp != "" {
		from.Attribs["slack_ts"] = e.ThreadTimestamp
	} else if e.EventTimestamp != "" {
		from.Attribs["slack_ts"] = e.EventTimestamp
	} else {
		from.Attribs["slack_ts"] = e.Timestamp
	}

	handle(Msg{
		At:   time.Now(),
		From: from,
		Body: e.Text,
	})
}

func (sui *SlackUI) stripAtAddress(ev *slack.MessageEvent) bool {
	var prefixes = []string{
		addressUser(sui.self.ID, ""),
		addressUser(sui.self.ID, sui.self.Name),
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

// addressUser creates the escape sequence for marking a user in a message.
func addressUser(userID string, userName string) string {
	if userName != "" {
		return fmt.Sprintf("<@%s|%s>:", userID, userName)
	}
	return fmt.Sprintf("<@%s>", userID)
}
