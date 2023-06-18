# otium: incremental automation of manual procedures

Otium attempts to solve the dilemma of procedures that are tedious,
time-consuming and error-prone to perform manually but that are impractical to
automate right now: maybe you don't have time, cannot yet see the justification,
don't know yet the best approach.

The idea is to automate incrementally, starting from an empty code skeleton that
substitutes the text document that you would normally write.

Instead of reading the document, you run the skeleton, which guides you
step-by-step through the whole procedure. It is a simple REPL that gathers
inputs from you and passes them to the subsequent steps.

When you have time and enough experience from proceeding manually, you automate
one step. When you have some more time, you automate another step, thus
incrementally automating the whole procedure over time.

## Scope

Otium aims to stay as simple as possible, because it is threading a fine line
between manual procedures guided by a document and full automation with code.

In particular, there is no plan to support conditional logic, because I have the
feeling that when the complexity of the procedure requires conditional logic, it
is time to "just" write a "normal" Go program instead, with the full power of
the language (we do not want to invent yet another DSL).

## Status

- Pre v1: assume API will change in a non-backward compatible way until v1.
- Before a breaking change a version tag will bet set.

## Example: WRITEME

## Examples in the source code

See directory [examples](examples).

## Usage

Write your own main and import this package.

Run the program. It has a REPL with command completion and history, thanks
to [peterh/liner]. Enter `?` or `<TAB>` twice to get started:

```
(top)>> Enter a command or '?' for help
(top)>> ?

Commands:
  help [<command> ...]
    Show help.

  repl
    Show help for the REPL.

  list
    Show the list of steps.

  next
    Run the next step.

  quit
    Quit the program.

  variables
    List the variables.

(top)>>
```

## Rendering bag values in the step description with Go template

Assuming that the procedure bag contains the k/v `Name: Joe`, then

```go
pcd.AddStep(&otium.Step{
    Desc: `Hello {{.Name}}!`
})
````

will be rendered as:

```
Hello Joe!
```

This feature is inspired by [danslimmon/donothing].

## Returning an error from a step

Sometimes an error is recoverable within the same execution, sometimes it is
unrecoverable.

If the error is recoverable, return an error as you please (wrapped or not),
for example:

```go
pcd.AddStep(&otium.Step{
    Run: func (bag otium.Bag) error {
        ...
        return fmt.Errorf("failed to X... %s", err)
    },
})
```

If the error is unrecoverable, use the `w` verb of `fmt.Errorf` (or equivalent):

```go
pcd.AddStep(&otium.Step{
    Run: func (bag otium.Bag) error {
        ...
        return fmt.Errorf("failed to X... %w", otium.ErrUnrecoverable)
    },
})
```

## Design decisions

- To test the interactive behavior, I wrote a minimal `expect` package, inspired
  by the great [Expect Tcl].
- Cannot inject `io.Reader`, `io.Writer` to ease testing because the REPL
  library has hardcoded os.Stdin and os.Stdout. Instead, we use
  `expect.NewFilePipe()` and swap `os.Stdin`, `os.Stdout` in each test.

## Credits

The idea of `otium` comes from [danslimmon/donothing] at v0.2.0, which had the
rationale that impressed me and the majority of preliminary code, but was still
missing support for running the user-provided automation functions.

## License

[MIT](LICENSE).

[danslimmon/donothing]: https://github.com/danslimmon/donothing

[peterh/liner]: https://github.com/peterh/liner

[Expect Tcl]: https://tsapps.nist.gov/publication/get_pdf.cfm?pub_id=821311
