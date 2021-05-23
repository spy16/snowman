package regex

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spy16/snowman"
)

// Entry represents a regular-expression patterns to intent mapping entry.
type Entry struct {
	snowman.Intent `json:",inline" yaml:",inline"`

	Patterns []string `json:"patterns" yaml:"patterns"`
	regexps  []*regexp.Regexp
}

func (rp *Entry) match(text string) []snowman.Intent {
	var res []snowman.Intent

	for _, rex := range rp.regexps {
		matches := rex.FindStringSubmatch(strings.ToLower(text))
		if len(matches) == 0 {
			continue
		}

		argc := 0
		ctx := clone(rp.Context)
		names := rex.SubexpNames()
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
		intent := rp.Intent
		intent.Context = ctx
		intent.Confidence = 1
		res = append(res, intent)
	}

	return res
}

func (rp *Entry) init() error {
	for _, pattern := range rp.Patterns {
		var buf strings.Builder

		space := false
		for _, c := range pattern {
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

		rp.regexps = append(rp.regexps, rex)
	}
	return nil
}

func clone(m map[string]interface{}) map[string]interface{} {
	cloned := map[string]interface{}{}
	for k, v := range m {
		cloned[k] = v
	}
	return cloned
}
