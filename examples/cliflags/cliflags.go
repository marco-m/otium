package main

import (
	"fmt"
	"os"

	"golang.org/x/exp/slices"

	"github.com/marco-m/otium"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func run() error {
	pcd := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Example showing command-line flags",
		Desc: `
Sometimes you know beforehand some of the variables that the procedure steps
will ask. In those cases, it can be simpler to pass them as command-line flags,
instead of waiting to be prompted for them.
`})

	pcd.AddStep(&otium.Step{
		Title: "Introduction",
		Desc: `
- To see all the variables directly settable from the command-line, re-invoke
  this program as:

    cliflags -h

- To see at any point in time the contents of the _bag_ (that is, all the
  variables available to this program, type this command from the (top) REPL:

    variables
`,
	})

	pcd.AddStep(&otium.Step{
		Title: "Two variables",
		Desc: `
- This step asks for two variables: 'fruit' and 'amount'.
  To set them as command-line flags, just invoke the program as:

    cliflags --fruit banana --amount 42

- The variable 'fruit' has a validator, that limits the acceptable inputs.
  Try it.
`,
		Vars: []otium.Variable{
			{
				Name: "fruit",
				Desc: "Fruit for breakfast",
				Fn: func(val string) error {
					basket := []string{"banana", "mango"}
					if !slices.Contains(basket, val) {
						return fmt.Errorf("we only have %s", basket)
					}
					return nil
				},
			},
			{Name: "amount", Desc: "How many pieces of fruit"},
		},
		Run: func(bag otium.Bag) error {
			fruit, err := bag.Get("fruit")
			if err != nil {
				return err
			}
			amount, err := bag.Get("amount")
			if err != nil {
				return err
			}

			fmt.Println("The variables are:")
			fmt.Println("fruit", fruit)
			fmt.Println("amount", amount)

			return nil
		},
	})

	return pcd.Execute()
}
