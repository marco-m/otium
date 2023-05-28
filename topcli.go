package otium

import (
	"fmt"
	"io"

	"github.com/alecthomas/kong"
)

type bind struct {
	pcd *Procedure
}

// Ensure the topcli compiles.
var _ = kong.Must(&topcli{})

// Top-level REPL commands.
type topcli struct {
	Help     helpCmd `cmd:"" help:"Show help."`
	Question helpCmd `cmd:"" hidden:"" name:"?" help:"Show help."`
	Repl     replCmd `cmd:"" help:"Show help for the REPL."`
	List     listCmd `cmd:"" help:"Show the list of steps."`
	Next     nextCmd `cmd:"" help:"Run the next step."`
	Quit     quitCmd `cmd:"" help:"Quit the program"`
}

type helpCmd struct {
	Command []string `arg:"" optional:"" help:"Show help on command."`
}

func (h *helpCmd) Run(realCtx *kong.Context) error {
	ctx, err := kong.Trace(realCtx.Kong, h.Command)
	if err != nil {
		return err
	}
	if ctx.Error != nil {
		return ctx.Error
	}
	err = ctx.PrintUsage(false)
	if err != nil {
		return err
	}
	fmt.Fprintln(realCtx.Stdout)
	return nil
}

type replCmd struct{}

func (hr *replCmd) Run(ctx *kong.Context) error {
	ctx.Printf(`
The REPL is based on https://github.com/peterh/liner; see there for the full
keystroke list.

Subset of the supported keystrokes:

Keystroke    Action
-----------  ------
Tab          List completions
Tab Tab      List all commands (if line is empty)
Up arrow     Previous match from history, like Fish shell
Down arrow   Next match from history, like Fish shell
Ctrl-R       Reverse Search history, like Bash shell (Esc cancel)
Ctrl-D       (if line is empty) End of File - usually quits application
`)
	return nil
}

type listCmd struct{}

func (l *listCmd) Run(bind *bind) error {
	printToc(bind.pcd)
	return nil
}

type nextCmd struct{}

func (n *nextCmd) Run(bind *bind) error {
	return cmdNext(bind.pcd)
}

type quitCmd struct{}

func (q *quitCmd) Run(bind *bind) error {
	// FIXME HACK USE PROPER ErrQuit sentinel instead!!!
	return io.EOF
}
