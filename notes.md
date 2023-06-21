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
