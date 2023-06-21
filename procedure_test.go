package otium_test

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/go-quicktest/qt"

	"github.com/marco-m/otium"
	"github.com/marco-m/otium/expect"
)

func TestProcedure_ExecuteWithZeroStepsFails(t *testing.T) {
	pcd := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Simple title",
		Desc:  `Simple description`,
	})

	err := pcd.Execute()

	qt.Assert(t, qt.ErrorMatches(err,
		"procedure has zero steps; want at least one"))
}

func TestProcedure_ExecuteStepWithMissingFieldsFails(t *testing.T) {
	pcd := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Simple title",
		Desc:  `Simple description`,
	})
	pcd.AddStep(&otium.Step{
		Title: "",
	})

	err := pcd.Execute()

	qt.Assert(t, qt.ErrorMatches(err, `step \(1\) has empty Title`))
}

func TestProcedure_ExecuteOneStepRunSuccess(t *testing.T) {
	stdin, stdout, exp := expect.NewFilePipe(100*time.Millisecond,
		expect.MatchMaxDef)

	// Due to the REPL library, we must swap os.Stdin and os.Stdout

	oldStdin := os.Stdin
	os.Stdin = stdin
	oldStdout := os.Stdout
	os.Stdout = stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	sut := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Simple title",
		Desc:  `Simple description`,
	})
	sut.AddStep(&otium.Step{
		Title: "step 1",
		Run: func(bag otium.Bag) error {
			fmt.Println("hello from step 1")
			return nil
		},
	})

	asyncErr := make(chan error)
	go func() {
		err := sut.Execute()
		stdout.Close()
		asyncErr <- err
	}()

	want1 := `# Simple title

Simple description

## Table of contents

next->  1. step 1


(top)>> Next step: (1) step 1
(top)>> Enter a command or '?' for help
(top)>> `
	// Flag (?s) means that . matches also \n
	have, err := exp.Expect(`(?s).*(\(top\)>>.*\n){2}\(top\)>> `)
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have, want1))

	// FIXME just found a bug!!!
	//have, err = exp.Expect(`not existing`)
	//qt.Assert(t, qt.ErrorIs())(t, err, expect.ErrTimeout)
	//qt.Assert(t, qt.Equal())(t, have, "")

	err = exp.Send("next\n")
	qt.Assert(t, qt.IsNil(err))

	have, err = exp.Expect(`.*hello from step 1`)
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have, "hello from step 1"))

	err = exp.Send("quit\n")
	qt.Assert(t, qt.IsNil(err))

	_, err = exp.Expect(`not existing`)
	qt.Assert(t, qt.ErrorIs(err, io.EOF))
	// FIXME TEST BROKEN qt.Assert(t, qt.Equals(have, ""))

	err = <-asyncErr
	qt.Assert(t, qt.IsNil(err))
}

func TestProcedure_ExecuteOneStepRunFailure(t *testing.T) {
	stdin, stdout, exp := expect.NewFilePipe(100*time.Millisecond,
		expect.MatchMaxDef)

	// Due to the REPL library, we must swap os.Stdin and os.Stdout

	oldStdin := os.Stdin
	os.Stdin = stdin
	oldStdout := os.Stdout
	os.Stdout = stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	sut := otium.NewProcedure(otium.ProcedureOpts{Title: "Simple title"})
	sut.AddStep(&otium.Step{
		Title: "step 1",
		Run: func(bag otium.Bag) error {
			return fmt.Errorf("flatlined %w", otium.ErrUnrecoverable)
		},
	})

	asyncErr := make(chan error)
	go func() {
		err := sut.Execute()
		stdout.Close()
		asyncErr <- err
	}()

	// Flag (?s) means that . matches also \n
	_, err := exp.Expect(`(?s).*\(top\)>>.*\(top\)>> `)
	qt.Assert(t, qt.IsNil(err))

	err = exp.Send("next\n")
	qt.Assert(t, qt.IsNil(err))

	_, err = exp.Expect(`not existing`)
	qt.Assert(t, qt.ErrorIs(err, io.EOF))
	// FIXME :-( qt.Assert(t, qt.Equals(have, ""))

	err = <-asyncErr
	qt.Assert(t, qt.ErrorIs(err, otium.ErrUnrecoverable))
}
