package snowman

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Intent identifiers for system events.
const (
	SysIntentUnknown  = "sys.unknown"
	SysIntentShutdown = "sys.shutdown"
)

var _ Classifier = (*RegexClassifier)(nil)

// Classifier is responsible for extracting all different intents of a message.
type Classifier interface {
	Classify(ctx context.Context, dialogue Dialogue, msg Msg) ([]Intent, error)
}

// Intent represents the intent of a message along with context extracted
// from the message.
type Intent struct {
	ID         string                 `json:"id"`
	Context    map[string]interface{} `json:"context"`
	Confidence float64                `json:"confidence"`
}

func (intent Intent) String() string {
	var args []string
	for k, v := range intent.Context {
		args = append(args, fmt.Sprintf("%s=%v", k, v))
	}
	return fmt.Sprintf("%s(%s)", intent.ID, strings.Join(args, ", "))
}

// RegexClassifier implements Classifier using simple regular expression matching
// technique. This type can also be loaded from JSON string.
type RegexClassifier struct {
	Patterns []rePattern `json:"patterns"`
}

// Register adds multiple regex patterns to the classifier to generate given
// intent ID.
func (rc *RegexClassifier) Register(intentID string, patterns ...string) error {
	for _, pattern := range patterns {
		var rp rePattern
		if err := rp.Init(pattern, intentID); err != nil {
			return err
		}
		rc.Patterns = append(rc.Patterns, rp)
	}
	return nil
}

// Classify uses the registered regex patterns to find all the intents that
// match the given message body.
func (rc *RegexClassifier) Classify(ctx context.Context, dialogue Dialogue, msg Msg) ([]Intent, error) {
	var result []Intent
	for _, rp := range rc.Patterns {
		if intent := rp.Match(msg); intent != nil {
			result = append(result, *intent)
		}

		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.Canceled) {
				break // context cancelled, return what we have.
			}
			return nil, err
		}
	}

	return result, nil
}

type rePattern struct {
	RegEx  *regexp.Regexp
	Intent string
}

func (p *rePattern) Init(pat, intent string) error {
	var buf strings.Builder

	space := false
	for _, c := range pat {
		if c == ' ' || c == '\t' {
			space = true
			continue
		} else if space {
			buf.WriteString("\\s+")
			space = false
		}

		buf.WriteRune(c)
	}

	rex, err := regexp.Compile(buf.String())
	if err != nil {
		return err
	}

	p.RegEx = rex
	p.Intent = intent
	return nil
}

func (p *rePattern) Match(msg Msg) *Intent {
	confidence := 1.0
	matches := p.RegEx.FindStringSubmatch(msg.Body)
	if len(matches) == 0 {
		confidence = 0.5
		matches = p.RegEx.FindStringSubmatch(strings.ToLower(msg.Body))
	}

	if len(matches) == 0 {
		return nil
	}

	argc := 0
	ctx := map[string]interface{}{}
	names := p.RegEx.SubexpNames()
	for i, match := range matches {
		if i == 0 {
			continue
		}
		name := strings.TrimSpace(names[i])
		if name == "" {
			name = fmt.Sprintf("arg%d", argc)
			argc++
		}
		ctx[name] = match
	}

	return &Intent{
		ID:         p.Intent,
		Context:    ctx,
		Confidence: confidence,
	}
}

func (p rePattern) MarshalJSON() ([]byte, error) {
	return json.Marshal([2]string{
		p.Intent,
		p.RegEx.String(),
	})
}

func (p *rePattern) UnmarshalJSON(bytes []byte) error {
	var val [2]string
	if err := json.Unmarshal(bytes, &val); err != nil {
		return err
	}
	return p.Init(val[0], val[1])
}
