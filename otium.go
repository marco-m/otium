// Package otium allows incremental automation of manual procedures
package otium

import "errors"

var (
	// ErrUnrecoverable tells the REPL to exit.
	//
	// From client code, use the %w verb of fmt.Errorf:
	// 	func(bag otium.Bag) error {
	//	    return fmt.Errorf("failed to X... %w", otium.ErrUnrecoverable)
	//	},
	ErrUnrecoverable = errors.New("(unrecoverable)")
)

// Internal errors and control flow.
var (
	errBack = errors.New("go back (sentinel)")
)

// RunFn is the function that automates a [Step]. When called, bag will contain
// all the key/value pairs set by the previous steps. A typical RunFn will use
// [Bag.Get] to get a k/v and [Bag.Put] to put a k/v.
type RunFn func(bag Bag) error

// Step is part of a [Procedure]. See [Procedure.Add].
type Step struct {
	// Title is the name of the step, shown in table of contents and at the
	// beginning of the step itself.
	Title string
	// Desc is the detailed description of what a human is supposed to do to
	// perform the step. If the step is automated, Desc should be shortened
	// and adapted to the change.
	Desc string
	// Run is the implementation of the step. If the step is completely manual,
	// Run must still be non nil. In this case, it can be empty, but probably
	// it should still ask for input. See [RunFn] for details.
	Run RunFn
}
