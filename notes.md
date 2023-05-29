- move timeout to makeExpectSend
- try to make a better expect with regexp or in any case no need to read byte by
  byte i think.
- careful with buffer handling: when giving copy to caller must extract correct
  window or beginning from the buffer; need tests here
- actually add timeout to MyExpect
- what to do when regexp does NOT match from the beginning? DO I trow the non
  matching away? Dropping seems the only practical approach, but before
  inventing my variant, what the original Tcl Expect does???
- can i do some of these tests without the need to go through a goroutine???

- add tests!!!
- set version string and replace the fixed one in the help
- added pre-commit hook; how can I share it easily?
- // FIXME HACK USE PROPER ErrQuit sentinel instead!!!
- Look for various FIXME and TODO
- add command "variables" to show the contents of the Bag
- uniform wrap all io.EOF and, simplify logic around all io.EOF and especially
  use errors.Is(err, io.ERR) to allow the wrapping
- replace fmt.Fprintln(pcd.Stdout, ...) with pcd.Println(...)
- look at all print functions and make them uniform
    - this should also add a prefix with the step name or similar? mhh too
      long...
- check that no two steps have the same Title
- probably remove kong to parse the REPL, keep it to parse the program itself?
-
support [checkpoints](https://en.wikipedia.org/wiki/Application_checkpointing)?
To allow re-running the procedure without starting from scratch, if something
in-between went wrong?
