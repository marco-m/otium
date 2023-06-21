package otium

import (
	"errors"
	"fmt"
	"strings"
)

// RunFn is the function that automates a [Step]. When called, bag will contain
// all the key/value pairs set by the previous steps. A typical RunFn will use
// [Bag.Get] to get a k/v and [Bag.Put] to put a k/v.
type RunFn func(bag Bag) error

// Step is part of a [Procedure]. See [Procedure.Add].
type Step struct {
	// Text is the Markdown description of the step. The first line is the
	// title, must begin with "##" (markdown for section title) and must be
	// separated from the rest of the text by a blank line. The title will be
	// also shown in the table of contents.
	//
	// The text following the title should be the detailed description of what
	// a human is supposed to do to perform the step. If the step is automated,
	// this part should be shortened and adapted to the change.
	//
	// Text will be rendered as Markdown!
	Text string
	// Run is the implementation of the step. If the step is completely manual,
	// Run must still be non nil. In this case, it can be empty, but probably
	// it should still ask for input. See [RunFn] for details.
	Run RunFn
}

// validate checks that step is valid. Meant to be called by Procedure.Exec.
// It also takes care of trimming step.Text.
func (step *Step) validate(stepN int) error {
	var errs []error

	step.Text = strings.TrimSpace(step.Text)

	if step.Text == "" {
		errs = append(errs, fmt.Errorf("step (%d) has empty Text", stepN))
	}

	return errors.Join(errs...)
}
