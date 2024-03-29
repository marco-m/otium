# Otium CHANGELOG

https://github.com/marco-m/otium

## v0.1.7 2023-7-29

### New

- Support user context via `Step.Run(bag Bag, uctx any) error`
- Support pre-flight checks via `ProcedureOpts.PreFlight() (any, error)`
- When invoked with `--doc-only`, an otium procedure prints the documentation instead of running.

### Breaking

- `Procedure.Execute`: change signature to `Execute(args []string) error`, to ease testing. Normally you would use this as `pcd.Execute(os.Args)`.
- `Step.Run`: change signature to `Run func(bag Bag, uctx any) error`

## v0.1.6 2023-7-24

### New

- When invoked with -h, an otium procedure prints also the otium version.

## v0.1.5 2023-7-23

### New

- Bag: all variables are automatically settable also as command-line flags (see README and
  examples/cliflags).
- Procedure.AddStep: new field Vars, to declare all the k/v for the Bag. This replaces the
  majority of behaviors of Bag.
- Table of Contents: each step is marked 🤖 (bot) if automated or 🤠 (human) if manual.

### Breaking

- Bag.Get: modified from `Bag.Get(key string, desc string)` to `Bag.Get(key string)`,
  since now the description goes into the Vars field of Procedure.AddStep.
- Bag.Get: now it behaves as GetNoAsk: it is _never_ interactive.
- Bag.GetValidate: removed, since now the validator function goes into the Vars field of
  Procedure.AddStep.
- Bag.GetNoAsk: removed, since now the validator function goes into the Vars field of
  Procedure.AddStep.

### Changed

- Step: the Run user function is now optional.

## v0.1.4 2023-06-20

### New

- Bag: add GetNoAsk method

## v0.1.3 2023-06-18

### New

- New command `variables` to show the contents of the procedure bag

## Release v0.1.2 2023-06-17

### New

- Step.Description: render bag values with Go template
- Bag: add GetValidate method
- CI: add CI, build on https://cirrus-ci.org/
- Add ErrUnrecoverable (sentinel)
- Add CHANGELOG

## Release v0.1.1 2023-06-13

### New

- Bag.Get: suggest the key to complete

## Release v0.1.0 2023-06-13

First release.
