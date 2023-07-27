- support command-line flags
  - [x] detect if two steps try to set the same Variable in field Vars and fail with error
  - [ ] add field HideCli; if present, the variable will not be settable as CLI flag
  - [ ] decide if we support default values, because if we support them, it means they can
    be overridden only as flags, unless we still do the autocompletion of the "set"
    command, and if a default value is present, we also autocomplete it, so that if the
    user just hits Enter he accepts the default, and to override he has to backspace to
    delete it? At this point, the same autocompletion of the value should be done also for
    all vars that have been set at the command-line, there is no difference with the
    default value...
  - [ ] support procedure-specific help
  - [ ] add callback AfterCliFn(bag) error, to otium.ProcedureOpts, that can be used to
    validate the results of the cli parsing (for example, mandatory flags)

- [ ] simplify all regexp of Expect by changing the prompt!!!
- [ ] add dry-run and generate document...

- [ ] add build version!

- [ ] add color to the titles

- support pre-flight checks, for example to check for the presence of aws-vault, in a
  cleaner way than what I need to do currently...

- I begin to think that rendering markdown is somehow too complicated for
  nothing, in addition I would not vene have a way to render in red the
  error text...
  - instead, i could use fatih/color as i do in timeit
  - and understand how glamour detects if the terminal is light or dark
```
	if style == "auto" {
		if termenv.HasDarkBackground() {
			return &DarkStyleConfig, nil
		}
		return &LightStyleConfig, nil
	}
```
  where termenv is https://github.com/muesli/termenv, which actually supports
  also colors and bold, so maybe I can use just this one?
  - very nice termenv has also integration with Go template!!!
  - actually i could contribute that logic to fathit/color! it is actually a
    mess...
  - ah I had already starred also https://github.com/jwalton/gchalk
- breaking: Step: merge fields Title and Desc into field Text
- does it make sense to simplify the API by removing the Title field and
  replacing it by the first line of the Desc? It might even be more readable
  when reading the code!
    - Maybe should even require to prefix the title with the markdown "##" ???
    - this is the real question :-/
- update the README when mentioning the Go template if needed
- WOW WOW could even render the description as markdown!!! in the supported
  terminals? There is already a tool that does that I think?
    - more deps: https://github.com/charmbracelet/glamour
- update version!!!!	description := "otium v0.0.0"
- does it make sense to have another prompt level, to which you can only
  answer "proceed (execute the step)" or "skip"? This somehow makes the flow
  less surprising when the step is _automated_, but risks also to become more
  annoying overall?
- this is similar to olivier's idea of generating a markdown document.
- somehow allow to skip steps? maybe from the Toc actually?
    - for example: new command: skip 3 4 15
- Find a way to pre-populate a subset of the bag with flags passed from the
  command-line! This would be cool but not so easy since the code that populates
  the bag is in the user-provided Run function :-(, I cannot pre-run it!
    - On the other hand, maybe I can split the user-provided function in two:
    - Inputs()  this one sets the bag??? - so that I can pre-run it ???
    - Run()
- the tests
    - TestProcedure_ExecuteOneStepRunSuccess
    - TestProcedure_ExecuteOneStepRunFailure need to be fixed/reconsided. I am
      running next at all?
- // FIXME just found a bug!!! in procedure_test :-(
- The goroutine used by Expect is racy, no protection for offset :-(
- add more tests!!!
- set version string and replace the fixed one in the help
- added pre-commit hook; how can I share it easily?
- // FIXME HACK USE PROPER ErrQuit sentinel instead!!!
- Look for various FIXME and TODO
- add command "variables" to show the contents of the Bag
- uniform wrap all io.EOF and, simplify logic around all io.EOF and especially
  use errors.Is(err, io.ERR) to allow the wrapping
- check that no two steps have the same Title
- probably remove kong to parse the REPL, keep it to parse the program itself?
-

support [checkpoints](https://en.wikipedia.org/wiki/Application_checkpointing)?
To allow re-running the procedure without starting from scratch, if something
in-between went wrong? This would be cool.
