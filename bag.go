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
	bag map[string]Variable
}

func NewBag() Bag {
	return Bag{bag: make(map[string]Variable)}
}

// Variable is an item of [Bag].
// Field Name is also the key in [Get]: if key "foo" exists, then:
// foo, _ := bag.Get("foo")
// foo.Name == "foo"
type Variable struct {
	Name string
	Desc string
	Fn   ValidatorFn
	val  string
	set  bool
}

// Get returns the value of key if key exists. If key doesn't exist, Get
// returns an error.
//
// In case of error, it means that you didn't set the Variables field of
// [Procedure.AddStep]. See the examples for clarification.
func (bag *Bag) Get(key string) (string, error) {
	variable := bag.bag[key]
	if !variable.set {
		return "", fmt.Errorf("key not found: %q", key)
	}
	return variable.val, nil
}

// Put adds key/val to bag, overwriting val if key already exists.
func (bag *Bag) Put(key, val string) {
	variable := bag.bag[key]
	variable.Name, variable.val = key, val
	variable.set = true
	bag.bag[key] = variable
}

// ValidatorFn is the optional function to validate a k/v pair. It is called
// either when parsing the command-line or when processing the Vars field of
// a [Step].
type ValidatorFn func(val string) error

// ask interactively asks the user for the value of key, prompting with desc.
// The key might already exist if it has been set as a command-line flag. In
// this case, ask returns the existing value.
// If the validator function fn(key, val) returns an error, ask shows the
// error to the user and keeps asking for a value. If fn returns no error,
// then ask stores the key/value in the bag.
func (bag *Bag) ask(key string, term *liner.State) (string, error) {
	variable := bag.bag[key]
	if variable.set {
		return variable.val, nil
	}

	term.SetCompleter(makeInputCompleter(key))

	for {
		fmt.Printf("(input)>> Enter %s (set %s <value>) or '?' for help\n",
			variable.Desc, key)
		line, err := term.PromptWithSuggestion(
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
			if variable.Fn != nil {
				if err := variable.Fn(val); err != nil {
					fmt.Println(err)
					continue
				}
			}
			bag.Put(key, val)
			return val, nil
		default:
			fmt.Printf("invalid: %q\n", line)
			continue
		}
	}
}

// makeInputCompleter returns a liner.Completer.
// We use a closure and a factory as an adapter, since this allows to pass the
// `key` parameter.
func makeInputCompleter(key string) liner.Completer {
	return func(line string) []string {
		commands := []string{"help", "?", "back", "set"}
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
	}
}
