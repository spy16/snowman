package regex

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/spy16/snowman"
)

// New returns a new regular-expressions based intent classifier.
func New(next snowman.Handler) *IntentClassifier {
	return &IntentClassifier{invoke: next}
}

// IntentClassifier implements snowman.Handler and provides regular expressions
// based intent tagging facilities.
type IntentClassifier struct {
	invoke  snowman.Handler
	entries []Entry
}

// Handle iterates through all the registered patterns and tags the message
// with the matching intents.
func (re *IntentClassifier) Handle(msg *snowman.Msg, di snowman.Dialogue) error {
	for _, pattern := range re.entries {
		intents := pattern.match(msg.Body)
		msg.Intents = append(msg.Intents, intents...)
	}
	return re.invoke.Handle(msg, di)
}

// Add adds pattern entries to the intent-classifier.
func (re *IntentClassifier) Add(entries ...Entry) error {
	for _, entry := range entries {
		if err := entry.init(); err != nil {
			return err
		}
		re.entries = append(re.entries, entry)
	}
	return nil
}

// Load loads patterns from yaml/json files found in the given directory.
func Load(re *IntentClassifier, intentsDir string) error {
	if walkErr := filepath.WalkDir(intentsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		var patterns []Entry
		if strings.HasSuffix(f.Name(), ".yaml") || strings.HasSuffix(f.Name(), ".yml") {
			if err := yaml.NewDecoder(f).Decode(&patterns); err != nil {
				return err
			}
		} else if strings.HasSuffix(f.Name(), ".json") {
			if err := json.NewDecoder(f).Decode(&patterns); err != nil {
				return err
			}
		} else {
			return nil
		}

		for i, intent := range patterns {
			if err := intent.init(); err != nil {
				return fmt.Errorf("error in '%s': %v", f.Name(), err)
			}
			patterns[i] = intent
		}

		re.entries = append(re.entries, patterns...)
		return nil
	}); walkErr != nil {
		return walkErr
	}

	return nil
}
