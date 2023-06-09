package otium

import (
	"fmt"
	"io"
	"strings"

	"github.com/peterh/liner"
)

// Bag is passed to the [RunFn] of [Step]. It contains all the k/v pairs added
// by the various steps during the execution of the otium [Procedure].
type Bag struct {
	bag       map[string]string
	linenoise *liner.State
}

// Get returns the value of key if key exists. If key doesn't exist, Get
// interactively asks the user for it, prompting with desc, and stores the
// key/value in the bag.
//
// NOTE If key is not present, Get is interactive and blocking. See the
// examples for idiomatic usage.
func (b *Bag) Get(key, desc string) (string, error) {
	return b.GetValidate(key, desc, func(k, v string) error { return nil })
}

type ValidatorFn func(key string, desc string) error

// GetValidate returns the value of key if key exists. If key doesn't exist,
// GetValidate interactively asks the user for it, prompting with desc.
// If fn(key, val) returns an error, it shows the error to the user and keeps
// asking for a value. If fn returns no error, then GetValidate stores the
// key/value in the bag.
//
// NOTE If key is not present, Get is interactive and blocking. See the
// examples for idiomatic usage.
func (b *Bag) GetValidate(key, desc string, fn ValidatorFn) (string, error) {
	if val, ok := b.bag[key]; ok {
		return val, nil
	}
	//
	// Configure completer.
	//
	commands := []string{"help", "?", "back", "set"}
	b.linenoise.SetCompleter(func(line string) []string {
		completions := make([]string, 0, len(commands))
		line = strings.ToLower(line)
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, line) {
				completions = append(completions, cmd)
			}
		}
		// Basic dynamic completer on key name.
		if strings.HasPrefix(line, "set") {
			completions = []string{"set " + key + " "}
		}
		//fmt.Printf("completer: %q, completions: %q\n", line, completions)
		return completions
	})

	for {
		fmt.Printf("(input)>> Enter %s (set %s <value>) or '?' for help\n",
			desc, key)
		line, err := b.linenoise.PromptWithSuggestion(
			"(input)>> ", "set "+key+" ", -1)
		if err != nil {
			if err == io.EOF {
				return "", io.EOF
			}
			fmt.Println(err)
			continue
		}
		// TODO should actually also parse to required type here
		//  (string or int)

		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			continue
		}
		switch tokens[0] {
		case "help", "?":
			fmt.Print(`
  set <key> <value>    set <key> to <value>
  back                 go back to the top level REPL

`)
			continue
		case "back":
			return "", errBack
		case "set":
			if len(tokens) != 3 {
				fmt.Printf("want: set <key> <value>; have: %q\n", tokens)
				continue
			}
			name, val := tokens[1], tokens[2]
			if name != key {
				fmt.Printf("set: wrong key: have %q; want %q\n", name, key)
				continue
			}
			if fn != nil {
				if err := fn(key, val); err != nil {
					fmt.Println(err)
					continue
				}
			}
			b.Put(key, val)
			return val, nil
		default:
			fmt.Printf("invalid: %q\n", line)
			continue
		}
	}
}

// GetNoAsk returns the value of key if key exists. If key doesn't exist,
// GetNoAsk returns an error.
//
// NOTE If key is not present, GetNoAsk will not prompt the user.
func (b *Bag) GetNoAsk(key string) (string, error) {
	val, ok := b.bag[key]
	if !ok {
		return "", fmt.Errorf("key not found: %q", key)
	}
	return val, nil
}

// Put adds key/val to bag, overwriting val if key already exists.
func (b *Bag) Put(key, val string) {
	b.bag[key] = val
}
