package otium

import (
	"errors"
	"fmt"
	"strings"
)

// Step is part of a [Procedure]. See [Procedure.Add].
type Step struct {
	// Title is the name of the step, shown in table of contents and at the
	// beginning of the step itself.
	Title string
	// Desc is the detailed description of what a human is supposed to do to
	// perform the step. If the step is automated, Desc should be shortened
	// and adapted to the change.
	Desc string
	// Vars are the new variables needed by the step.
	Vars []Variable
	// Run is the optional automation of the step. If the step is manual,
	// leave Run unset. When called, bag will contain all the key/value pairs
	// set by the previous steps and uctx, if not nil, will point to the user
	// context.
	// A typical Run will use [Bag.Get] to get a k/v, [Bag.Put] to put a k/v and
	// will type assert uctx to the type returned by [ProcedureOpts.PreFlight].
	// For the user context, see also [ProcedureOpts.PreFlight] and
	// examples/usercontext.
	Run func(bag Bag, uctx any) error
}

// validate checks that step is valid. Meant to be called by Procedure.Exec.
func (step *Step) validate(stepN int) error {
	var errs []error

	step.Title = strings.TrimSpace(step.Title)
	step.Desc = strings.TrimSpace(step.Desc)

	if step.Title == "" {
		errs = append(errs, fmt.Errorf("step (%d) has empty Title", stepN))
	}

	return errors.Join(errs...)
}

func (step *Step) Icon() string {
	if step.Run != nil {
		return "🤖"
	}
	return "🤠"
}
