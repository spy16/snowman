# ⛄️ Snowman

> WIP

_Snowman_ provides facilities for building chat-bots in Go.

A bot consists of `UI` to interact with users, an Intent `Classifier` to detect the intent behind a `Msg` from a user
and a `Processor` to respond to the user based on the intent detected.

## Usage

### As Library

```go
package main

import (
	"context"

	"github.com/spy16/snowman"
)

func main() {
	logger := logurs.New()

	if err := snowman.Run(context.Background(),
		snowman.WithName("Snowy"),
		snowman.WithUI(snowman.ConsoleUI{}),
		snowman.WithClassifier(myIntentClassifer),
		snowman.WithProcessor(myIntentProcessor),
		snowman.WithLogger(logger),
	); err != nil {
		logger.Errorf("snowy exited with error: %v", err)
	}
}
```

### As Binary

1. Install using `go get -u -v github.com/spy16/snowman/cmd/snowman`
2. Run using `snowman --name=snowy`

## TODO

- [ ] Support for channels / independent conversations.
- [ ] `Processor` with `Module` mechanism for intent classes.
- [ ] Slack RTM & Socket UI implementations.
- [ ] NLP with [prose](https://github.com/jdkato/prose)
- [ ] An intent classifier with FFNet.
