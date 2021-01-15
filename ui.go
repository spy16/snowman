package snowman

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

var _ UI = (*ConsoleUI)(nil)

// UI represents the interface used for interactions between the bot and
// the user.
type UI interface {
	// Listen should a start a goroutine that writes the messages sent
	// by the user to the returned channel. GoRoutine should exit when
	// the context is cancelled.
	Listen(ctx context.Context) (<-chan Msg, error)

	// Say should display the message to the user as if the bot said it.
	// Target user that should receive the message should be identified
	// using the 'user' value.
	Say(ctx context.Context, user string, msg Msg) error
}

// ConsoleUI implements a console based UI. Stdin is used for reading
// input from the user and Stdout is used for output.
type ConsoleUI struct{ Prompt string }

func (cui ConsoleUI) Listen(ctx context.Context) (<-chan Msg, error) {
	if cui.Prompt == "" {
		cui.Prompt = ">> "
	}

	out := make(chan Msg)
	go func() {
		defer close(out)

		sc := bufio.NewScanner(os.Stdin)
		for ctx.Err() == nil {
			fmt.Print(cui.Prompt)
			if !sc.Scan() {
				break
			}

			msg := Msg{
				From:    "user",
				Body:    sc.Text(),
				Attribs: nil,
			}

			select {
			case <-ctx.Done():
				return
			case out <- msg:
			}
		}
	}()

	return out, nil
}

func (cui ConsoleUI) Say(_ context.Context, _ string, msg Msg) error {
	fmt.Print("\r" + strings.Repeat(" ", len(cui.Prompt)+5))
	fmt.Println("\r" + msg.Body)
	fmt.Print(cui.Prompt)
	return nil
}
