package otium_test

import (
	"github.com/marco-m/otium/expect"
	"io"
	"os"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/marco-m/otium"
)

func TestProcedure_ExecuteWithZeroStepsFails(t *testing.T) {
	pcd := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Simple title",
		Desc:  `Simple description`,
	})

	err := pcd.Execute()

	assert.Error(t, err, "procedure has zero steps; want at least one")
}

func TestProcedure_ExecuteStepWithMissingFieldsFails(t *testing.T) {
	pcd := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Simple title",
		Desc:  `Simple description`,
	})
	pcd.AddStep(&otium.Step{
		Title: "",
		Run:   nil,
	})

	err := pcd.Execute()

	assert.Error(t, err,
		"step (1) has empty Title\nstep (1 ) misses Run function")
}

func TestProcedure_ExecuteOneStepWithEmptyRunSuccess(t *testing.T) {
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
		Run:   func(bag otium.Bag) error { return nil },
	})

	asyncErr := make(chan error)
	go func() {
		err := sut.Execute()
		stdout.Close()
		asyncErr <- err
	}()

	// (?s) = . matches also \n
	have, err := exp.Expect(`(?s).*\(top\)>>.*\(top\)>> `)
	assert.NilError(t, err, "buf:\n%q", have)
	want1 := `# Simple title

Simple description

## Table of contents

next->  1. step 1

(top)>> Enter a command or '?' for help
(top)>> `
	assert.Equal(t, have, want1)

	err = exp.Send("quit\n")
	assert.NilError(t, err)

	have, err = exp.Expect(`.*fa`)
	assert.ErrorIs(t, err, io.EOF, "buf:\n%q", have)

	err = <-asyncErr
	assert.ErrorIs(t, err, io.EOF)
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
	assert.NilError(t, err)

	err = exp.Send("next\n")
	assert.NilError(t, err)

	have, err := exp.Expect(`not existing`)
	assert.ErrorIs(t, err, io.EOF)
	assert.Equal(t, have, "")

	err = <-asyncErr
	assert.ErrorIs(t, err, otium.ErrUnrecoverable)
}
