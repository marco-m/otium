// Package otium allows incremental automation of manual procedures
package otium

import (
	"errors"
	"runtime/debug"
)

var (
	// ErrUnrecoverable tells the REPL to exit.
	//
	// From client code, use the %w verb of fmt.Errorf:
	// 	func(bag otium.Bag) error {
	//	    return fmt.Errorf("failed to X... %w", otium.ErrUnrecoverable)
	//	},
	ErrUnrecoverable = errors.New("(unrecoverable)")

	version = "something-went-wrong"
)

// Internal errors and control flow.
var (
	errBack = errors.New("go back (sentinel)")
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		version = "ReadBuildInfo-failed"
		return
	}
	version = findOtiumVersion(info)
}

func findOtiumVersion(info *debug.BuildInfo) string {
	// FIXME I should get this at runtime instead of hardcoding, but how?
	const otiumPath = "github.com/marco-m/otium"

	// If called by a program that uses otium as a dependency, then the otium
	// version is in one of the dependency modules.
	// Not robust if "replace" directive is used in go.mod
	for _, mod := range info.Deps {
		if mod.Path == otiumPath {
			return mod.Version
		}
	}

	// If called from one of the examples in the otium module, then the
	// otium version is in the Main module and will always be "(devel)",
	// due to a Go limitation.
	// Not robust if "replace" directive is used in go.mod
	if info.Main.Path == otiumPath {
		return info.Main.Version
	}

	return "findOtiumVersion-internal-error"
}
