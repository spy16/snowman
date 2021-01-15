# ⛄️ Snowman

⛄ Snowman provides facilities for building chatbots in Go.

## Usage

### As Library

```go
package main

import "github.com/spy16/snowman"

func main() {
	logger := logurs.New()

	bot := snowman.New(
		snowman.WithName("Rick Sanchez"),
		snowman.WithClassifier(myIntentClassifer),
		snowman.WithProcessor(myIntentProcessor),
		snowman.WithLogger(logger),
	)
	if err := bot.Run(ctx); err != nil {
		logger.Errorf("bot exited with error: %v", err)
	}
}
```

### As Binary

1. Install using `go get -u -v github.com/spy16/snowman/cmd/snowman`
2. Run using `snowman --name=snowy`

## ffnet

`ffnet` package provides a generic Fully-Connected Feedforward neural network, and a Stochastic Gradient Descent based
trainer. Following snippet shows how to create a simple 2-10-10 network with ReLU for hidden and sigmoid for output
layer activations and how to make a prediction.

```go
net, _ := ffnet.New(2,
ffnet.Layer(10, ffnet.ReLU()),
ffnet.Layer(10, ffnet.Sigmoid()),
)
net.Predict(1, 1)
```

