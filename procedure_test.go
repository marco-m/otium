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

var osArgs = []string{"exe.name"}

func TestProcedure_ExecuteWithZeroStepsFails(t *testing.T) {
	pcd := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Simple title",
		Desc:  `Simple description`,
	})

	err := pcd.Execute(osArgs)

	qt.Assert(t, qt.ErrorMatches(err,
		"procedure has zero steps; want at least one"))
}

func TestProcedure_ExecuteStepWithMissingFieldsFails(t *testing.T) {
	pcd := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Simple title",
		Desc:  `Simple description`,
	})
	pcd.AddStep(&otium.Step{Title: ""})

	err := pcd.Execute(osArgs)

	qt.Assert(t, qt.ErrorMatches(err, `step \(1\) has empty Title`))
}

func TestProcedure_ExecuteDuplicateVarsInSameStepFail(t *testing.T) {
	pcd := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Simple title",
		Desc:  `Simple description`,
	})
	const key = "fruit"
	pcd.AddStep(&otium.Step{
		Title: "Step A",
		Vars: []otium.Variable{
			{Name: key},
			{Name: key},
		},
	})

	err := pcd.Execute(osArgs)

	qt.Assert(t, qt.ErrorMatches(err, `step "Step A": duplicate var "fruit"`))
}

func TestProcedure_ExecuteDuplicateVarsInDifferentStepsFail(t *testing.T) {
	pcd := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Simple title",
		Desc:  `Simple description`,
	})
	pcd.AddStep(&otium.Step{
		Title: "Step A",
		Vars:  []otium.Variable{{Name: "fruit"}},
	})
	pcd.AddStep(&otium.Step{
		Title: "Step B",
		Vars:  []otium.Variable{{Name: "fruit"}},
	})

	err := pcd.Execute(osArgs)

	qt.Assert(t, qt.ErrorMatches(err, `step "Step B": duplicate var "fruit"`))
}

func TestProcedure_ExecuteOneStepNoRunWithVarFromCLI(t *testing.T) {
	exp, cleanup := expect.NewFilePipe(100*time.Millisecond, expect.MatchMaxDef)
	defer cleanup()

	sut := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Simple title",
		Desc:  `Simple description`,
	})
	sut.AddStep(&otium.Step{
		Title: "step 1",
		Vars:  []otium.Variable{{Name: "fruit"}},
	})

	asyncErr := make(chan error)
	go func() {
		err := sut.Execute([]string{"exe.name", "--fruit=mango"})
		os.Stdout.Close()
		asyncErr <- err
	}()

	want1 := `# Simple title

Simple description

## Table of contents

next->  1. 🤠 step 1


(top) Next step: 1. 🤠 step 1
(top) Enter a command or '?' for help
(top)>> `
	// Flag (?s) means that . matches also \n
	have, err := exp.Expect(`(?s).*\(top\)>> `)
	qt.Check(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have, want1))

	err = exp.Send("next\n")
	qt.Assert(t, qt.IsNil(err))

	have, err = exp.Expect(`.*terminated successfully`)
	qt.Check(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have, "(top) Procedure terminated successfully"))

	err = <-asyncErr
	qt.Assert(t, qt.IsNil(err))
}

func TestProcedure_ExecuteOneStepWithRunSuccess(t *testing.T) {
	exp, cleanup := expect.NewFilePipe(100*time.Millisecond, expect.MatchMaxDef)
	defer cleanup()

	sut := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Simple title",
		Desc:  `Simple description`,
	})
	sut.AddStep(&otium.Step{
		Title: "step 1",
		Run: func(bag otium.Bag, uctx any) error {
			fmt.Println("hello from step 1")
			return nil
		},
	})

	asyncErr := make(chan error)
	go func() {
		err := sut.Execute(osArgs)
		os.Stdout.Close()
		asyncErr <- err
	}()

	want1 := `# Simple title

Simple description

## Table of contents

next->  1. 🤖 step 1


(top) Next step: 1. 🤖 step 1
(top) Enter a command or '?' for help
(top)>> `
	// Flag (?s) means that . matches also \n
	have, err := exp.Expect(`(?s).*\(top\)>> `)
	qt.Check(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have, want1))

	// FIXME just found a bug!!!
	//have, err = exp.Expect(`not existing`)
	//qt.Check(t, qt.ErrorIs())(t, err, expect.ErrTimeout)
	//qt.Assert(t, qt.Equal())(t, have, "")

	err = exp.Send("next\n")
	qt.Assert(t, qt.IsNil(err))

	have, err = exp.Expect(`.*hello from step 1`)
	qt.Check(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have, "hello from step 1"))

	err = exp.Send("quit\n")
	qt.Assert(t, qt.IsNil(err))

	_, err = exp.Expect(`not existing`)
	qt.Assert(t, qt.ErrorIs(err, io.EOF))
	// FIXME TEST BROKEN qt.Assert(t, qt.Equals(have, ""))

	err = <-asyncErr
	qt.Assert(t, qt.IsNil(err))
}

func TestProcedure_ExecuteOneStepWithRunFailure(t *testing.T) {
	exp, cleanup := expect.NewFilePipe(100*time.Millisecond,
		expect.MatchMaxDef)
	defer cleanup()

	sut := otium.NewProcedure(otium.ProcedureOpts{Title: "Simple title"})
	sut.AddStep(&otium.Step{
		Title: "step 1",
		Run: func(bag otium.Bag, uctx any) error {
			return fmt.Errorf("flatlined %w", otium.ErrUnrecoverable)
		},
	})

	asyncErr := make(chan error)
	go func() {
		err := sut.Execute(osArgs)
		os.Stdout.Close()
		asyncErr <- err
	}()

	_, err := exp.Expect(`\(top\)>> `)
	qt.Assert(t, qt.IsNil(err))

	err = exp.Send("next\n")
	qt.Assert(t, qt.IsNil(err))

	_, err = exp.Expect(`not existing`)
	qt.Assert(t, qt.ErrorIs(err, io.EOF))
	// FIXME :-( qt.Assert(t, qt.Equals(have, ""))

	err = <-asyncErr
	qt.Assert(t, qt.ErrorIs(err, otium.ErrUnrecoverable))
}

func TestGenerateDocNoVars(t *testing.T) {
	exp, cleanup := expect.NewFilePipe(100*time.Millisecond,
		expect.MatchMaxDef)
	defer cleanup()

	sut := otium.NewProcedure(otium.ProcedureOpts{
		Title: "My preferred fruits",
		Desc:  `An overview of fabulous fruits.`,
	})
	sut.AddStep(&otium.Step{
		Title: "Red fruits",
		Desc:  `Watermelon`,
	})
	sut.AddStep(&otium.Step{
		Title: "Blue fruits",
		Desc:  `Blueberry`,
	})

	asyncErr := make(chan error)
	go func() {
		err := sut.Execute([]string{"exe.name", "--doc-only"})
		asyncErr <- err
	}()

	want := `# My preferred fruits

An overview of fabulous fruits.

## Table of contents

next->  1. 🤠 Red fruits
        2. 🤠 Blue fruits


## 1. 🤠 Red fruits

Watermelon


## 2. 🤠 Blue fruits

Blueberry
`

	// Flag (?s) means that . matches also \n
	have, err := exp.Expect(`(?s).*Blueberry\n`)
	qt.Check(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have, want))

	err = <-asyncErr
	qt.Assert(t, qt.IsNil(err))
}

func TestGenerateDocWithVars(t *testing.T) {
	exp, cleanup := expect.NewFilePipe(100*time.Millisecond,
		expect.MatchMaxDef)
	defer cleanup()

	sut := otium.NewProcedure(otium.ProcedureOpts{
		Title: "My preferred fruits",
		Desc:  `An overview of fabulous fruits.`,
	})
	sut.AddStep(&otium.Step{
		Title: "Red fruits",
		Desc:  `Watermelon`,
		Vars: []otium.Variable{{
			Name: "fruit",
			Desc: "Your preferred fruit",
		}},
	})

	asyncErr := make(chan error)
	go func() {
		err := sut.Execute([]string{"exe.name", "--doc-only"})
		asyncErr <- err
	}()

	want := `# My preferred fruits

An overview of fabulous fruits.

## Table of contents

next->  1. 🤠 Red fruits


## 1. 🤠 Red fruits

Watermelon
`

	// Flag (?s) means that . matches also \n
	have, err := exp.Expect(`(?s).*Watermelon\n`)
	qt.Check(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have, want))

	err = <-asyncErr
	qt.Assert(t, qt.IsNil(err))
}

func TestPreFlight(t *testing.T) {
	exp, cleanup := expect.NewFilePipe(100*time.Millisecond,
		expect.MatchMaxDef)
	defer cleanup()

	type usercontext struct {
		answer int
	}

	sut := otium.NewProcedure(otium.ProcedureOpts{
		Title: "The title",
		PreFlight: func() (any, error) {
			return &usercontext{answer: 42}, nil
		},
	})
	sut.AddStep(&otium.Step{
		Title: "Modify user context",
		Run: func(bag otium.Bag, uctx any) error {
			fooClient := uctx.(*usercontext)
			qt.Check(t, qt.Equals(fooClient.answer, 42))
			fooClient.answer = 99
			return nil
		},
	})
	sut.AddStep(&otium.Step{
		Title: "Read modified user context",
		Run: func(bag otium.Bag, uctx any) error {
			fooClient := uctx.(*usercontext)
			qt.Check(t, qt.Equals(fooClient.answer, 99))
			fmt.Println("end of test")
			return nil
		},
	})

	asyncErr := make(chan error)
	go func() {
		err := sut.Execute(osArgs)
		asyncErr <- err
	}()

	want1 := `# The title



## Table of contents

next->  1. 🤖 Modify user context
        2. 🤖 Read modified user context


(top) Next step: 1. 🤖 Modify user context
(top) Enter a command or '?' for help
(top)>> `
	// Flag (?s) means that . matches also \n
	have1, err := exp.Expect(`(?s).*\(top\)>> `)
	qt.Check(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have1, want1))

	err = exp.Send("next\n")
	qt.Assert(t, qt.IsNil(err))

	want2 := `
## 1. 🤖 Modify user context


(top) Next step: 2. 🤖 Read modified user context
(top) Enter a command or '?' for help
(top)>> `
	// Flag (?s) means that . matches also \n
	have2, err := exp.Expect(`(?s).*\(top\)>> `)
	qt.Check(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have2, want2))

	err = exp.Send("next\n")
	qt.Assert(t, qt.IsNil(err))

	_, err = exp.Expect(`end of test\n`)
	qt.Check(t, qt.IsNil(err))

	err = <-asyncErr
	qt.Assert(t, qt.IsNil(err))
}
