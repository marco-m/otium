package otium

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	steps   []*Step
	stepIdx int // Index into the step to execute.
	bag     Bag
	parser  *kong.Kong
	// Warning: term will be initialized by Execute(), not by NewProcedure().
	term *liner.State
}

// ProcedureOpts is used by [NewProcedure] to create a Procedure.
type ProcedureOpts struct {
	// Name is the name of the Procedure; by default it is the name of the executable.
	Name string
	// Title is the title of the Procedure, shown at the beginning of the program.
	Title string
	// Desc is the summary of what the procedure is about, shown at the beginning of
	// the program, just after the Title.
	Desc string
}

// NewProcedure creates a Procedure.
func NewProcedure(opts ProcedureOpts) *Procedure {
	return &Procedure{
		ProcedureOpts: opts,
		bag:           NewBag(),
		parser: kong.Must(&topcli{},
			kong.Name(""),
			kong.Exit(func(int) {}),
			kong.ConfigureHelp(kong.HelpOptions{
				// Must be disabled because it doesn't make sense in a REPL.
				NoAppSummary:   true,
				WrapUpperBound: 80,
			}),
			// Must be disabled because it doesn't make sense in a REPL.
			kong.NoDefaultHelp(),
		),
	}
}

// AddStep adds a [Step] to [Procedure].
func (pcd *Procedure) AddStep(step *Step) {
	pcd.steps = append(pcd.steps, step)
}

// Execute starts the [Procedure] by putting the user into a REPL.
// If it returns an error, the user program should print it and exit with a
// non-zero status code. See the examples for the suggested usage.
func (pcd *Procedure) Execute() error {
	var errs []error
	errs = append(errs, pcd.validate())
	for i, step := range pcd.steps {
		errs = append(errs, step.validate(i+1))
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}

	// Fill the bag with all the Vars from all the steps.
	// A duplicate variable is considered an error.
	for _, step := range pcd.steps {
		for _, variable := range step.Vars {
			if variable.Fn == nil {
				variable.Fn = func(string) error { return nil }
			}
			// Detect duplicates.
			if _, ok := pcd.bag.bag[variable.Name]; ok {
				err := fmt.Errorf("step %q: duplicate var %q", step.Title, variable.Name)
				errs = append(errs, err)
				continue
			}
			pcd.bag.bag[variable.Name] = variable
		}
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}

	// Add the Vars in the bag as CLI flags.
	for name, variable := range pcd.bag.bag {
		name, variable := name, variable // avoid loop capture :-(
		flag.Func(name, variable.Desc,
			func(val string) error {
				if err := variable.Fn(val); err != nil {
					return err
				}
				pcd.bag.Put(name, val)
				return nil
			})
	}

	// Parse the command-line.
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "%s: %s.\n", pcd.Name, pcd.Title)
		fmt.Fprintf(out,
			"This program is based on otium %s, a simple incremental automation system (https://github.com/marco-m/otium)\n",
			version)
		fmt.Fprintf(out, "\nUsage of %s:\n", pcd.Name)
		flag.PrintDefaults()
	}
	flag.Parse()

	// We cannot initialize liner before (say, in NewProcedure), because
	// NewLiner changes the terminal line discipline, so we must do this
	// _after_ having parsed the command-line.
	pcd.term = liner.NewLiner()
	// Restore terminal to previous mode, super important.
	defer pcd.term.Close()
	pcd.term.SetCtrlCAborts(true)

	//
	// Configure completer, part 1.
	// TODO think how to make it know deeper about the possible completions...
	//
	pcd.term.SetTabCompletionStyle(liner.TabPrints)
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

	fmt.Printf("# %s\n\n", pcd.Title)
	fmt.Printf("%s\n", pcd.Desc)
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
		pcd.term.SetCompleter(topCompleter)

		next := pcd.steps[pcd.stepIdx]
		fmt.Printf("\n(top)>> Next step: %d. %s %s\n",
			pcd.stepIdx+1, next.Icon(), next.Title)
		fmt.Printf("(top)>> Enter a command or '?' for help\n")
		var line string
		line, err := pcd.term.Prompt("(top)>> ")
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
		pcd.term.AppendHistory(line)

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

func (pcd *Procedure) Put(key, val string) {
	pcd.bag.Put(key, val)
}

func (pcd *Procedure) validate() error {
	var errs []error

	if pcd.Name == "" {
		_, pcd.Name = filepath.Split(os.Args[0])
	}
	pcd.Title = strings.TrimSpace(pcd.Title)
	pcd.Desc = strings.TrimSpace(pcd.Desc)

	if pcd.Title == "" {
		errs = append(errs,
			errors.New("procedure has empty title"))
	}
	if len(pcd.steps) == 0 {
		errs = append(errs,
			errors.New("procedure has zero steps; want at least one"))
	}

	return errors.Join(errs...)
}

// Table of contents
func printToc(pcd *Procedure) {
	fmt.Printf("\n## Table of contents\n\n")
	for i, step := range pcd.steps {
		var next string
		if i == pcd.stepIdx {
			next = "next->"
		}
		fmt.Printf("%6s %2d. %s %s\n", next, i+1, step.Icon(), step.Title)
	}
	fmt.Println()
}
