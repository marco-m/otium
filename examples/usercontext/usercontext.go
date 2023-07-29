package main

import (
	"fmt"
	"os"

	"github.com/marco-m/otium"
	"github.com/marco-m/otium/examples/usercontext/foo"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func run() error {
	pcd := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Example showing usage of the user context",
		Desc: `
Sometimes you need to initialize and then pass to each step a user context,
for example a struct representing an API client.

In this example the user context is initialized:
1. after the CLI parsing, so that asking for --help works no matter what, AND
2. before the first step, so that in case of error we can stop the procedure
   before entering the REPL.`,
		PreFlight: func() (any, error) {
			if os.Getenv("FRUIT") == "" {
				return nil, fmt.Errorf("missing environment variable FRUIT")
			}
			return foo.NewClient(), nil
		},
	})

	pcd.AddStep(&otium.Step{
		Title: "Using the context for the first time",
		Desc: `
This step retrieves and modifies the context.`,
		Run: func(bag otium.Bag, uctx any) error {
			// Type assertion (https://go.dev/tour/methods/15)
			fooClient := uctx.(*foo.Client)

			fmt.Println(fooClient)
			fooClient.Something()
			return nil
		},
	})

	pcd.AddStep(&otium.Step{
		Title: "Using the context for the second time",
		Desc: `
This step retrieves the modified context.`,
		Run: func(bag otium.Bag, uctx any) error {
			fooClient := uctx.(*foo.Client)
			fmt.Println(fooClient)
			return nil
		},
	})

	return pcd.Execute(os.Args)
}
