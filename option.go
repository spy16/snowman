package snowman

import (
	"strings"
)

// Option can be provided to New() to customise Bot initialisation.
type Option func(bot *Bot)

// WithClassifier sets the intent classifier to be used by the bot.
// If classifier is nil, a default classifier which always returns
// UNKNOWN intent will be used.
func WithClassifier(classifier Classifier) Option {
	return func(bot *Bot) {
		if classifier == nil {
			classifier = classifyUnknown
		}
		bot.cls = classifier
	}
}

// WithProcessor sets the intent processor to be used by the bot. If
// processor is nil, a default processor with constant response will
// be used.
func WithProcessor(processor Processor) Option {
	return func(bot *Bot) {
		if processor == nil {
			processor = processUnknown
		}
		bot.proc = processor
	}
}

// WithUI sets the UI to be used by the bot. If ui is nil, a console
// UI will be used.
func WithUI(ui UI) Option {
	return func(bot *Bot) {
		if ui == nil {
			ui = &ConsoleUI{Prompt: "user > "}
		}
		bot.ui = ui
	}
}

// WithName sets the name of the bot. If name is empty string, default
// name will be used.
func WithName(name string) Option {
	return func(bot *Bot) {
		name = strings.TrimSpace(name)
		if name == "" {
			name = defaultName
		}
		bot.name = name
	}
}

// WithLogger sets the Logger be used by the bot. If lg is nil, logging
// will be disabled.
func WithLogger(lg Logger) Option {
	return func(bot *Bot) {
		if lg == nil {
			lg = noOpLogger{}
		}
		bot.logger = lg
	}
}

func withDefaults(opts []Option) []Option {
	return append([]Option{
		WithName(defaultName),
		WithLogger(stdLogger{}),
		WithUI(nil),
		WithClassifier(nil),
		WithProcessor(nil),
	}, opts...)
}
