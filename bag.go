package otium

import (
	"fmt"
	"io"
	"strings"

	"github.com/peterh/liner"
)

type Bag struct {
	bag       map[string]string
	linenoise *liner.State
}

func (b *Bag) Get(key, desc string) (string, error) {
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
		fmt.Printf("(input)>> Enter %s (set %s <value>) or '?' for help\n", desc, key)
		line, err := b.linenoise.Prompt("(input)>> ")
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
			fmt.Print(help)
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
			b.bag[key] = val
			return val, nil
		default:
			fmt.Printf("invalid: %q\n", line)
			continue
		}
	}
}

const help = `
  set <key> <value>    set <key> to <value>
  back                 go back to the top level REPL

`

func (b *Bag) Put(key, val string) {
	b.bag[key] = val
}
