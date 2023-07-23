package expect

import (
	"errors"
	"io"
	"os"
	"regexp"
	"time"
)

var ErrTimeout = errors.New("expect: timeout")

const MatchMaxDef = 2_000
const TimeoutDef = 60 * time.Second

// From the original Expect paper:
// > The exact string matched (or read but unmatched, if a timeout occurred)
// > is stored in the variable expect_match. Patterns must match the entire
// > output of the current process since the previous expect (hence the
// > reason most are surrounded by the * wildcard). However, more than 2000
// > bytes of output can force earlier bytes to be "forgotten". This may be
// > changed by setting the variable match_max.
type Expect struct {
	Reader   io.Reader
	Writer   io.Writer
	Timeout  time.Duration
	MatchMax int
	buf      []byte
	match    []byte
	offset   int
}

// New returns stdin as io.PipeReader, stdout as io.PipeWriter, and expect.
// Once stdin and stdout are replaced in the SUT, Expect and Send can be
// used by the test to communicate with the SUT.
// If the timeout expires, the next call to Expect or Send will fail the test.
// Note that communication is blocking, so the test must run the SUT in a
// goroutine.
func New(timeout time.Duration, matchMax int) (
	*io.PipeReader, *io.PipeWriter, *Expect,
) {
	stdinRd, stdinWr := io.Pipe()
	stdoutRd, stdoutWr := io.Pipe()

	exp := &Expect{
		Reader:   stdoutRd,
		Writer:   stdinWr,
		Timeout:  timeout,
		MatchMax: matchMax,
	}

	return stdinRd, stdoutWr, exp
}

// NewFilePipe swaps os.Stdin with stdinRd, os.Stdout with stdoutWr, where stdin
// and stdout are of type os.File. It returns expect and a cleanup  function
// to defer.
// Use this constructor when your target writes directly to os.Stdout
// and reads directly from os.Stdin (that is, it doesn't offer the possibility
// to inject a pair of io.Reader, io.Writer).
//
// WARNING: in case the underlying pipe syscall fails, this function will
// panic.
func NewFilePipe(timeout time.Duration, matchMax int) (*Expect, func()) {
	var err error
	stdinRd, stdinWr, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stdoutRd, stdoutWr, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	exp := &Expect{
		Reader:   stdoutRd,
		Writer:   stdinWr,
		Timeout:  timeout,
		MatchMax: matchMax,
	}

	oldStdin := os.Stdin
	os.Stdin = stdinRd
	oldStdout := os.Stdout
	os.Stdout = stdoutWr
	cleanup := func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}

	return exp, cleanup
}

func (e *Expect) Expect(re string) (string, error) {
	if e.MatchMax == 0 {
		e.MatchMax = MatchMaxDef
	}
	if e.Timeout == 0 {
		e.Timeout = TimeoutDef
	}
	if e.buf == nil {
		e.buf = make([]byte, e.MatchMax)
	}
	if e.match == nil {
		e.match = make([]byte, e.MatchMax)
	}

	reg, err := regexp.Compile(re)
	if err != nil {
		return "", err
	}

	// We use a goroutine to avoid a blocking read. This does not happen for
	// io.Pipe() but it does happen for os.Pipe()
	type readInfo struct {
		n   int
		err error
	}
	readCh := make(chan readInfo, 1)

	timer := time.NewTimer(e.Timeout)
	var n int
	// We handle in this way the error from the Read because in case of
	// error, Read can still return n > 0 useful bytes.
	for err == nil {
		go func() {
			n, err := e.Reader.Read(e.buf[e.offset:])
			readCh <- readInfo{n: n, err: err}
		}()
		select {
		case <-timer.C:
			// Timeout: return the buffer contents.
			// FIXME This leaks the goroutine and the channel :-(
			return string(e.buf[:e.offset]), ErrTimeout
		case rInfo := <-readCh:
			n, err = rInfo.n, rInfo.err
			e.offset += n
			if loc := reg.FindIndex(e.buf[:e.offset]); loc != nil {
				iL, iR := loc[0], loc[1]
				copy(e.match, e.buf[iL:iR])
				remain := e.offset - iR
				if remain > 0 {
					// Copy what is left after the match.
					copy(e.buf, e.buf[iR:e.offset])
					e.offset = remain
				} else {
					e.offset = 0
				}
				// Return the match.
				return string(e.match[:iR-iL]), nil
			}
			if e.offset == cap(e.buf) {
				// Reached the end of the buf. Copy half for overlap and start
				// another cycle.
				e.offset = cap(e.buf) / 2
				copy(e.buf, e.buf[cap(e.buf)-e.offset:])
			}
		}
	}
	// EOF or any other read error: return the buffer contents.
	return string(e.buf[:e.offset+n]), err
}

func (e *Expect) Send(msg string) error {
	_, err := e.Writer.Write([]byte(msg))
	return err
}

// Drain reads and discards the Expect input until EOF or error.
func (e *Expect) Drain() (int64, error) {
	return io.Copy(io.Discard, e.Reader)
}

// FixedReader reads from Reader always a fixed amount of bytes, N.
type FixedReader struct {
	Reader io.Reader // underlying reader
	N      int       // How many bytes to read (the "fixed" part)
}

func (fix *FixedReader) Read(p []byte) (n int, err error) {
	if fix.N <= 0 {
		return 0, io.EOF
	}
	if len(p) > fix.N {
		p = p[0:fix.N]
	}
	n, err = fix.Reader.Read(p)
	return
}
