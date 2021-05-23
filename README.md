# ⛄️ Snowman

> WIP

_Snowman_ provides facilities for building chat-bots in Go.

A bot consists of `UI` to interact with users, and `Handler` to understand the intent, figure out the action and finally
generate response. Handler follows the familiar `http.Handler` approach and allows for all the standard patterns used (
e.g., middlewares).

## Usage

### As Library

Refer [cmd/snowman/main.go](./cmd/snowman/main.go) for usage example.

### As Binary

1. Install using `go get -u -v github.com/spy16/snowman/cmd/snowman`
2. Run using `snowman --name=snowy --slack=<token>`

## TODO

- [x] Support for channels / independent conversations.
- [x] Slack RTM
- [ ] Socket UI implementations.
- [ ] NLP with [prose](https://github.com/jdkato/prose)
- [ ] An intent classifier with FFNet.
