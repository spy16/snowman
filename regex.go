package snowman

import (
	"context"
	"regexp"
)

// RegexClassifier implements a simple intent classifier using regular expression
// patterns. Zero value is safe for use.
type RegexClassifier struct {
	patterns []struct {
		re *regexp.Regexp
		id string
	}
}

// Classify iterates through each registered pattern and tries to match the msg body
// with it. If a match is identified, all the named expressions are inserted into the
// intent context and returned. If not match is found, returns SysIntentUnknown.
func (rc *RegexClassifier) Classify(_ context.Context, msg Msg) (Intent, error) {
	for _, p := range rc.patterns {
		matches := p.re.FindStringSubmatch(msg.Body)
		if len(matches) > 0 {
			in := Intent{ID: p.id, Ctx: map[string]interface{}{}}
			names := p.re.SubexpNames()
			for i, match := range matches {
				in.Ctx[names[i]] = match
			}
			return in, nil
		}
	}
	return Intent{ID: SysIntentUnknown}, nil
}

// Register registers the pattern and intent ID. When a message that matches the given
// pattern is received, returned intent will have the intentID provided here.
func (rc *RegexClassifier) Register(pattern string, intentID string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	rc.patterns = append(rc.patterns, struct {
		re *regexp.Regexp
		id string
	}{re: re, id: intentID})
	return nil
}
