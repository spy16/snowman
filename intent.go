package snowman

import (
	"fmt"
	"strings"
)

// Intent represents the intent of a message along with context extracted
// from the message.
type Intent struct {
	Tag        string                 `json:"tag" yaml:"tag"`
	Context    map[string]interface{} `json:"context" yaml:"context"`
	Response   string                 `json:"response" yaml:"response"`
	Confidence float64                `json:"confidence" yaml:"confidence"`
}

// Clone returns a clone of the intent with the given context maps merged into
// the base context from the intent. Given maps act as overrides on the base.
func (intent Intent) Clone(mergeCtx ...map[string]interface{}) Intent {
	cloned := intent
	cloned.Context = cloneMerge(intent.Context, cloneMerge(mergeCtx...))
	return cloned
}

func (intent Intent) String() string {
	var args []string
	for k, v := range intent.Context {
		args = append(args, fmt.Sprintf("%s=%v", k, v))
	}
	return fmt.Sprintf("%s(%s)", intent.Tag, strings.Join(args, ", "))
}

func cloneMerge(maps ...map[string]interface{}) map[string]interface{} {
	merged := map[string]interface{}{}
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}
