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

	version = "dev"
)

// Internal errors and control flow.
var (
	errBack = errors.New("go back (sentinel)")
)
