package expect_test

import (
	"fmt"
	"github.com/marco-m/otium/expect"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestExpectSuccess(t *testing.T) {
	type want struct {
		re  string
		out string
	}
	type testCase struct {
		name     string
		input    string
		matchMax int
		want     []want
	}

	run := func(t *testing.T, tc testCase) {
		assert.Assert(t, len(tc.want) > 0)
		sut := expect.Expect{
			Reader:   strings.NewReader(tc.input),
			MatchMax: tc.matchMax,
		}
		for _, want := range tc.want {
			t.Run("re "+want.re, func(t *testing.T) {
				have, err := sut.Expect(want.re)
				assert.NilError(t, err)
				assert.Equal(t, have, want.out)
			})
		}
	}

	testCases := []testCase{
		{
			name:  "one expect, complete match",
			input: "0123456789",
			want:  []want{{re: `.*89`, out: "0123456789"}},
		},
		{
			name:  "one expect, left partial match",
			input: "0123456789",
			want:  []want{{re: `45.*89`, out: "456789"}},
		},
		{
			name:  "one expect, right partial match",
			input: "0123456789",
			want:  []want{{re: `.*56`, out: "0123456"}},
		},
		{
			name:     "one expect, input gt maxMatch odd",
			input:    "0123456789",
			matchMax: 5,
			want:     []want{{re: `.*78`, out: "678"}},
		},
		{
			name:     "one expect, input gt maxMatch even",
			input:    "0123456789",
			matchMax: 6,
			want:     []want{{re: `.*78`, out: "345678"}},
		},
		{
			name:  "two expects, complete match",
			input: "0123456789",
			want: []want{
				{re: `.*45`, out: "012345"},
				{re: `.*89`, out: "6789"},
			},
		},
		{
			name:  "two expects, hole in the middle",
			input: "0123456789",
			want: []want{
				{re: `.*23`, out: "0123"},
				{re: `6.*9`, out: "6789"},
			},
		},
		{
			name:     "two expects, input gt maxMatch odd",
			input:    "1234567890ABCDE",
			matchMax: 5,
			want: []want{
				{re: `.*7`, out: "4567"},
				{re: `.*C`, out: "ABC"},
			},
		},
		{
			name:     "two expects, input gt maxMatch even",
			input:    "1234567890ABCDE",
			matchMax: 6,
			want: []want{
				{re: `.*7`, out: "4567"},
				{re: `.*C`, out: "890ABC"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestExpectNoMatchEOF(t *testing.T) {
	type testCase struct {
		name     string
		input    string
		matchMax int
		re       string
		wantOut  string
	}

	run := func(t *testing.T, tc testCase) {
		sut := expect.Expect{
			Reader:   strings.NewReader(tc.input),
			MatchMax: tc.matchMax,
		}
		have, err := sut.Expect(tc.re)
		assert.ErrorIs(t, err, io.EOF)
		assert.Equal(t, have, tc.wantOut)
	}

	testCases := []testCase{
		{
			name:    "empty input",
			input:   "",
			re:      `.*9`,
			wantOut: "",
		},
		{
			name:    "buffer gt input",
			input:   "123",
			re:      `.*a`,
			wantOut: "123",
		},
		{
			name:     "buffer lt input odd",
			input:    "123456789",
			matchMax: 5,
			re:       `.*a`,
			wantOut:  "789",
		},
		{
			name:     "buffer lt input even",
			input:    "123456789",
			matchMax: 6,
			re:       `.*a`,
			wantOut:  "789",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestExpectShortReadsMatch(t *testing.T) {
	fixed := &expect.FixedReader{
		Reader: strings.NewReader("0123456789"),
		N:      1,
	}
	sut := expect.Expect{
		Reader:   fixed,
		MatchMax: 5,
	}
	have, err := sut.Expect(`.*8`)
	assert.NilError(t, err, "buf: %s", have)
	assert.Equal(t, have, "678")
}

func TestExpectShortReadsNoMatch(t *testing.T) {
	fixed := &expect.FixedReader{
		Reader: strings.NewReader("0123456789"),
		N:      1,
	}

	sut := expect.Expect{
		Reader:   fixed,
		MatchMax: 5,
	}
	have, err := sut.Expect(`.*A`)
	assert.ErrorIs(t, err, io.EOF, "buf: %s", have)
	assert.Equal(t, have, "6789")
}

func TestExpectSimulateSutSuccess(t *testing.T) {
	stdin, stdout, exp := expect.New(100*time.Millisecond, expect.MatchMaxDef)

	target := func(stdin io.Reader, stdout io.Writer) error {
		fmt.Fprint(stdout, "1234567890")
		var line string
		if _, err := fmt.Fscan(stdin, &line); err != nil {
			return err
		}
		want := "hello"
		if line != want {
			return fmt.Errorf("stdin: have: %q, want: %q", line, want)
		}
		fmt.Fprint(stdout, "cafefade")
		return nil
	}

	asyncErr := make(chan error)
	go func() {
		err := target(stdin, stdout)
		stdout.Close()
		asyncErr <- err
	}()

	have, err := exp.Expect(`.*7`)
	assert.NilError(t, err)
	assert.Equal(t, have, "1234567")

	err = exp.Send("hello\n")
	assert.NilError(t, err)

	have, err = exp.Expect(`.*fa`)
	assert.NilError(t, err)
	assert.Equal(t, have, "890cafefa")

	err = <-asyncErr
	assert.NilError(t, err)
}

func TestExpectSimulateSutTimeout(t *testing.T) {
	stdin, stdout, exp := expect.New(100*time.Millisecond, expect.MatchMaxDef)

	target := func(stdin io.Reader, stdout io.Writer) error {
		for i := 1; i < 20; i++ {
			time.Sleep(10 * time.Millisecond)
			fmt.Fprint(stdout, i)
		}
		return nil
	}

	asyncErr := make(chan error)
	go func() {
		err := target(stdin, stdout)
		stdout.Close()
		asyncErr <- err
	}()

	have, err := exp.Expect(`.*HELLO`)
	assert.ErrorIs(t, err, expect.ErrTimeout)
	assert.Assert(t, strings.HasPrefix(have, "12345"), "have: %s", have)

	_, err = exp.Drain()
	assert.NilError(t, err)

	err = <-asyncErr
	assert.NilError(t, err)
}

func TestExpectSimulateSutOsPipeSuccess(t *testing.T) {
	stdin, stdout, exp := expect.NewFilePipe(100*time.Millisecond,
		expect.MatchMaxDef)

	// We simulate a sut that writes directly to os.Stdout and reads
	// directly from os.Stdin

	oldStdin := os.Stdin
	os.Stdin = stdin
	oldStdout := os.Stdout
	os.Stdout = stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	target := func() error {
		fmt.Print("1234567890")
		var line string
		if _, err := fmt.Scan(&line); err != nil {
			return err
		}
		want := "hello"
		if line != want {
			return fmt.Errorf("stdin: have: %q, want: %q", line, want)
		}
		fmt.Print("cafefade")
		return nil
	}

	asyncErr := make(chan error)
	go func() {
		err := target()
		stdout.Close()
		asyncErr <- err
	}()

	have, err := exp.Expect(`.*7`)
	assert.NilError(t, err)
	assert.Equal(t, have, "1234567")

	err = exp.Send("hello\n")
	assert.NilError(t, err)

	have, err = exp.Expect(`.*fa`)
	assert.NilError(t, err)
	assert.Equal(t, have, "890cafefa")

	err = <-asyncErr
	assert.NilError(t, err)
}

func TestExpectSimulateSutOsPipeTimeout(t *testing.T) {
	stdin, stdout, exp := expect.NewFilePipe(100*time.Millisecond,
		expect.MatchMaxDef)

	// We simulate a sut that writes directly to os.Stdout and reads
	// directly from os.Stdin

	oldStdin := os.Stdin
	os.Stdin = stdin
	oldStdout := os.Stdout
	os.Stdout = stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	target := func() error {
		for i := 1; i < 20; i++ {
			time.Sleep(10 * time.Millisecond)
			fmt.Print(i)
		}
		return nil
	}

	asyncErr := make(chan error)
	go func() {
		err := target()
		stdout.Close()
		asyncErr <- err
	}()

	have, err := exp.Expect(`.*HELLO`)
	assert.ErrorIs(t, err, expect.ErrTimeout)
	assert.Assert(t, strings.HasPrefix(have, "12345"), "have: %s", have)

	_, err = exp.Drain()
	assert.NilError(t, err)

	err = <-asyncErr
	assert.NilError(t, err)
}
