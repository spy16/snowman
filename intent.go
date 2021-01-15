package snowman

import "context"

// Intent identifiers for system events.
const (
	IntentSysShutdown = "sys.shutdown"
)

var (
	_ Classifier = ClassifierFunc(nil)
	_ Processor  = ProcessorFunc(nil)
	_ Classifier = classifyUnknown
	_ Processor  = processUnknown
)

// Intent represents intention of a message sent by a user. Intent is used to
// generate action or response.
type Intent struct {
	ID  string                 `json:"id"`
	Msg Msg                    `json:"msg"`
	Ctx map[string]interface{} `json:"ctx"`
}

// Classifier is responsible for detecting user intent from the message.
// Intent is then used to generate action or response by Processor.
type Classifier interface {
	Classify(ctx context.Context, msg Msg) (Intent, error)
}

// ClassifierFunc is a Go func adaptor to implement Classifier.
type ClassifierFunc func(ctx context.Context, msg Msg) (Intent, error)

func (ic ClassifierFunc) Classify(ctx context.Context, msg Msg) (Intent, error) { return ic(ctx, msg) }

// Processor is responsible for generating a response or action based on the
// intent detected from the message.
type Processor interface {
	// Process can process the intent and generate a response that should be
	// sent to the user. If the msg has empty body, no message will be sent.
	Process(ctx context.Context, intent Intent) (Msg, error)
}

// ProcessorFunc is a Go func adaptor to implement Classifier.
type ProcessorFunc func(ctx context.Context, intent Intent) (Msg, error)

func (pf ProcessorFunc) Process(ctx context.Context, intent Intent) (Msg, error) {
	return pf(ctx, intent)
}

var (
	classifyUnknown = ClassifierFunc(func(ctx context.Context, msg Msg) (Intent, error) {
		return Intent{
			ID:  "UNKNOWN",
			Msg: msg,
			Ctx: nil,
		}, nil
	})

	processUnknown = ProcessorFunc(func(ctx context.Context, intent Intent) (Msg, error) {
		if intent.ID == IntentSysShutdown {
			return Msg{
				Body: "👋🙂 Bye!",
			}, nil
		}

		return Msg{
			Body: "I don't understand what you are saying 😐",
		}, nil
	})
)
