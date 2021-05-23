package snowman

import (
	"context"
	"fmt"
	"time"
)

// Msg represents a message from the user/bot. Msg can contain additional
// context in terms of intents that may be used by handler to generate response.
type Msg struct {
	ctx context.Context

	At      time.Time `json:"at"`
	To      User      `json:"to"`
	From    User      `json:"from"`
	Body    string    `json:"body"`
	Intents []Intent  `json:"intents"`
}

// Context returns the context associated with the message.
func (m Msg) Context() context.Context { return m.ctx }

func (m Msg) String() string { return fmt.Sprintf("Msg<to=@%s,from=@%s>", m.To.ID, m.From.ID) }

// User represents a user that is interacting with snowman.
type User struct {
	ID      string                 `json:"id"`
	Name    string                 `json:"name"`
	Attribs map[string]interface{} `json:"attribs"`
}

func (u User) String() string { return fmt.Sprintf("User<ID=@%s>", u.ID) }
