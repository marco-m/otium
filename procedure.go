package otium

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/google/shlex"
	"github.com/peterh/liner"
)

// Procedure is made of a sequence of [Step]. Create it with [NewProcedure],
// fill it with [Procedure.AddStep] and finally start it with
// [Procedure.Execute].
type Procedure struct {
	ProcedureOpts
	steps []*Step
	// Index into the step to execute.
	stepIdx int
	bag     Bag
	parser  *kong.Kong
	// Warning: will be initialized by Execute(), not by NewProcedure().
	linenoise *liner.State
}

// ProcedureOpts is used by [NewProcedure] to create a Procedure.
type ProcedureOpts struct {
	// Title is the name of the Procedure, shown at the beginning of the
	// program.
	Title string
	// Desc is the summary of what the procedure is about, shown at the
	// beginning of the program, just after the Title.
	Desc string
}

// NewProcedure creates a Procedure.
func NewProcedure(opts ProcedureOpts) *Procedure {
	pcd := &Procedure{
		ProcedureOpts: opts,
		bag:           Bag{bag: make(map[string]string)},
	}

	description := "otium v0.0.0"

	pcd.parser = kong.Must(&topcli{},
		kong.Name(""),
		kong.Description(
			description+" -- a simple incremental automation system (https://github.com/marco-m/otium)"),
		kong.Exit(func(int) {}),
		kong.ConfigureHelp(kong.HelpOptions{
			// Must be disabled because it doesn't make sense in a REPL.
			NoAppSummary:   true,
			WrapUpperBound: 80,
		}),
		// Must be disabled because it doesn't make sense in a REPL.
		kong.NoDefaultHelp(),
	)

	return pcd
}

// AddStep adds a [Step] to [Procedure].
func (pcd *Procedure) AddStep(step *Step) {
	pcd.steps = append(pcd.steps, step)
}

// Execute starts the [Procedure] by putting the user into a REPL.
// If it returns an error, the user program should print it and exit with a
// non-zero status code. See the examples for the suggested usage.
func (pcd *Procedure) Execute() error {
	if len(pcd.steps) == 0 {
		return errors.New("procedure has zero steps; want at least one")
	}

	var errs []error
	for i, step := range pcd.steps {
		if strings.TrimSpace(step.Title) == "" {
			errs = append(errs,
				fmt.Errorf("step (%d) has empty Title", i+1))
		}
		if step.Run == nil {
			errs = append(errs,
				fmt.Errorf("step (%d %s) misses Run function", i+1, step.Title))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	pcd.linenoise = liner.NewLiner()
	// Restore terminal to previous mode, super important.
	defer pcd.linenoise.Close()
	pcd.bag.linenoise = pcd.linenoise
	pcd.linenoise.SetCtrlCAborts(true)

	//
	// Configure completer, part 1.
	// TODO think how to make it know deeper about the possible completions...
	//
	pcd.linenoise.SetTabCompletionStyle(liner.TabPrints)
	var commands []string
	for _, node := range pcd.parser.Model.Children {
		commands = append(commands, node.Name)
	}
	topCompleter := func(line string) []string {
		completions := make([]string, 0, len(commands))
		line = strings.ToLower(line)
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, line) {
				completions = append(completions, cmd)
			}
		}
		//fmt.Printf("completer: %q, completions: %q\n", line, completions)
		return completions
	}

	fmt.Printf("# %s\n\n", strings.TrimSpace(pcd.Title))
	fmt.Printf("%s\n", strings.TrimSpace(pcd.Desc))
	printToc(pcd)

	//
	// Main loop.
	//
	var kongCtx *kong.Context
	for {
		if pcd.stepIdx == len(pcd.steps) {
			fmt.Printf("\n(top)>> Procedure terminated successfully\n")
			return nil
		}

		// We set the completer on each loop because the sub repl in bag.Get
		// sets its own.
		pcd.linenoise.SetCompleter(topCompleter)

		next := pcd.steps[pcd.stepIdx]
		fmt.Printf("\n(top)>> Next step: (%d) %s\n", pcd.stepIdx+1,
			strings.TrimSpace(next.Title))
		fmt.Printf("(top)>> Enter a command or '?' for help\n")
		var line string
		line, err := pcd.linenoise.Prompt("(top)>> ")
		// TODO if we receive EOF, we should return no error if the procedure
		//  has not started yet, and an error if the procedure has already
		//  started
		if err != nil {
			return err
		}

		//
		// Parse user input.
		//
		var args []string
		args, err = shlex.Split(line)
		if err != nil {
			pcd.parser.Errorf("%s", err)
			continue
		}
		kongCtx, err = pcd.parser.Parse(args)
		if err != nil {
			pcd.parser.Errorf("%s", err)
			continue
		}
		pcd.linenoise.AppendHistory(line)

		//
		// Execute user command.
		//
		err = kongCtx.Run(&bind{pcd: pcd})
		if errors.Is(err, io.EOF) || errors.Is(err, ErrUnrecoverable) {
			return err
		}
		if errors.Is(err, errBack) {
			continue
		}
		if err != nil {
			pcd.parser.Errorf("%s", err)
			continue
		}
	}
}

// Table of contents
func printToc(pcd *Procedure) {
	fmt.Printf("\n## Table of contents\n\n")
	for i, step := range pcd.steps {
		var next string
		if i == pcd.stepIdx {
			next = "next->"
		}
		fmt.Printf("%6s %2d. %s\n", next, i+1,
			strings.TrimSpace(step.Title))
	}
	fmt.Println()
}

func cmdNext(pcd *Procedure) error {
	if pcd.stepIdx >= len(pcd.steps) {
		return fmt.Errorf("next: internal error: step index > len(steps)")
	}
	step := pcd.steps[pcd.stepIdx]
	fmt.Printf("\n## (%d) %s\n\n", pcd.stepIdx+1, strings.TrimSpace(step.Title))

	if step.Desc != "" {
		if err := render(os.Stdout, strings.TrimSpace(step.Desc),
			pcd.bag); err != nil {
			return fmt.Errorf("%s %w", err, ErrUnrecoverable)
		}
		fmt.Printf("\n\n")
	}

	if err := step.Run(pcd.bag); err != nil {
		return fmt.Errorf("step %d: %w", pcd.stepIdx+1, err)
	}
	pcd.stepIdx++

	return nil
}
