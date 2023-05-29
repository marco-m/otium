package expect

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"testing"
	"time"

	"gotest.tools/v3/assert"
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
// Once stdin and stdout are replaced in the SUT, expect and send can be
// used by the test to communicate with the SUT.
// If the timeout expires, the next expect or send will fail the test.
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

// NewFilePipe returns stdin as os.File, stdout as os.File, and expect.
// Use this constructor when your target writes directly to os.Stdout
// and reads directly from os.Stdin (that is, it doesn't offer the possibility
// to inject a pair of io.Reader, io.Writer).
//
// WARNING: in case the underlying pipe syscall fails, this function will
// panic.
func NewFilePipe(timeout time.Duration, matchMax int) (
	*os.File, *os.File, *Expect,
) {
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

	return stdinRd, stdoutWr, exp
}

// See
// https://github.com/golang/go/issues/28790#issuecomment-438992014
// https://go.dev/play/p/iGeuSFtebI-
func bufferedPipe() {
	var lines = []string{
		"This is a test.",
		"Or is it?",
		"No one knows...",
		"Hmmm...",
		"I think I need more text.",
		"Gosh darn it.",
	}

	rd, wr := io.Pipe()

	go func() {
		bufWr := bufio.NewWriterSize(wr, 64)
		defer wr.Close()
		defer bufWr.Flush()

		for _, line := range lines {
			fmt.Fprintln(bufWr, line)
			//fmt.Printf("Buffer size: %v\n", bufWr.Buffered())
		}
	}()

	s := bufio.NewScanner(rd)
	for s.Scan() {
		//fmt.Printf("%q\n", s.Text())
		time.Sleep(time.Second)
	}
	if err := s.Err(); err != nil {
		panic(err)
	}
}

func (e *Expect) ExpectT(t *testing.T, re string) string {
	t.Helper()
	buf, err := e.Expect(re)
	assert.NilError(t, err)
	return string(buf)
}

func (e *Expect) SendT(t *testing.T, msg string) {
	t.Helper()
	_, err := e.Writer.Write([]byte(msg))
	assert.NilError(t, err)
}

func (e *Expect) Expect(re string) ([]byte, error) {
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
		return nil, err
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
			return e.buf[:e.offset], ErrTimeout
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
				return e.match[:iR-iL], nil
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
	return e.buf[:e.offset+n], err
}

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
