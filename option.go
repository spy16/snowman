package snowman

import (
	"strings"
)

// Option can be provided to New() to customise Bot initialisation.
type Option func(bot *Bot)

// WithExecutor sets the executor to be used for processing the intents.
// If exec is nil, default Executor will be used.
func WithExecutor(exec Executor) Option {
	return func(bot *Bot) {
		bot.exec = exec
	}
}

// WithClassifier sets the intent classifier to be used by the bot.
// If classifier is nil, a default classifier which always returns
// UNKNOWN intent will be used.
func WithClassifier(classifier Classifier) Option {
	return func(bot *Bot) {
		if classifier == nil {
			re := &RegexClassifier{}
			_ = re.Register(SysIntentUnknown, ".*")
		}
		bot.cls = classifier
	}
}

// WithUI sets the UI to be used by the bot. WithUI can be used multiple
// times to use multiple interfaces.
func WithUI(ui UI) Option {
	return func(bot *Bot) {
		if ui == nil {
			return
		}
		bot.uis = append(bot.uis, ui)
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
		bot.self = User{
			ID:   strings.ToLower(name),
			Name: name,
		}
	}
}

// WithLogger sets the Logger be used by the bot. If lg is nil, logging
// will be disabled.
func WithLogger(lg Logger) Option {
	return func(bot *Bot) {
		if lg == nil {
			lg = NoOpLogger{}
		}
		bot.logger = lg
	}
}

func withDefaults(opts []Option) []Option {
	return append([]Option{
		WithName(defaultName),
		WithLogger(StdLogger{}),
		WithUI(nil),
		WithClassifier(nil),
		WithExecutor(nil),
	}, opts...)
}
