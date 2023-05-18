package otium

import "errors"

var errBack = errors.New("go back (sentinel)")

type RunFn func(Bag) error

type Step struct {
	Title string
	Desc  string
	Run   RunFn
}
