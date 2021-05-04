package snowman

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

var _ UI = (*ConsoleUI)(nil)

// UI represents the interface used for interactions between the bot and
// the user.
type UI interface {
	// Say should display the message to the user as if the bot said it.
	// Target user that should receive the message should be identified
	// using the 'user' value.
	Say(ctx context.Context, msg Msg) error

	// Listen should listen for messages from user and deliver them to the
	// handler. Listen should block until UI reaches a terminal state or
	// until context is cancelled.
	Listen(ctx context.Context, handle func(msg Msg)) error
}

// ConsoleUI implements a console based UI. Stdin is used for reading
// input from the user and Stdout is used for output.
type ConsoleUI struct {
	Prompt string

	once    sync.Once
	scanner *bufio.Scanner
}

func (cui *ConsoleUI) Listen(ctx context.Context, handle func(msg Msg)) error {
	cui.once.Do(func() {
		if cui.Prompt == "" {
			cui.Prompt = ">> "
		}
		cui.scanner = bufio.NewScanner(os.Stdin)
	})

	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil
			}
			return ctx.Err()

		default:
			fmt.Print(cui.Prompt)
			if !cui.scanner.Scan() {
				return nil
			}
			handle(Msg{
				From: User{ID: "user"},
				Body: cui.scanner.Text(),
			})
		}
	}
}

func (cui *ConsoleUI) Say(_ context.Context, msg Msg) error {
	fmt.Print("\r" + strings.Repeat(" ", len(cui.Prompt)+5))
	fmt.Println("\r" + msg.Body)
	return nil
}
