package expect_test

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-quicktest/qt"

	"github.com/marco-m/otium/expect"
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
		qt.Assert(t, qt.IsTrue(len(tc.want) > 0))
		sut := expect.Expect{
			Reader:   strings.NewReader(tc.input),
			MatchMax: tc.matchMax,
		}
		for _, want := range tc.want {
			t.Run("re "+want.re, func(t *testing.T) {
				have, err := sut.Expect(want.re)
				qt.Assert(t, qt.IsNil(err))
				qt.Assert(t, qt.Equals(have, want.out))
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
		qt.Assert(t, qt.ErrorIs(err, io.EOF))
		qt.Assert(t, qt.Equals(have, tc.wantOut))
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
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have, "678"))
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
	qt.Assert(t, qt.ErrorIs(err, io.EOF))
	qt.Assert(t, qt.Equals(have, "6789"))
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
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have, "1234567"))

	err = exp.Send("hello\n")
	qt.Assert(t, qt.IsNil(err))

	have, err = exp.Expect(`.*fa`)
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have, "890cafefa"))

	err = <-asyncErr
	qt.Assert(t, qt.IsNil(err))
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
	qt.Assert(t, qt.ErrorIs(err, expect.ErrTimeout))
	qt.Assert(t, qt.IsTrue(strings.HasPrefix(have, "12345")),
		qt.Commentf("have: %s", have))

	_, err = exp.Drain()
	qt.Assert(t, qt.IsNil(err))

	err = <-asyncErr
	qt.Assert(t, qt.IsNil(err))
}

func TestExpectSimulateSutOsPipeSuccess(t *testing.T) {
	exp, cleanup := expect.NewFilePipe(100*time.Millisecond,
		expect.MatchMaxDef)
	defer cleanup()

	// We simulate a sut that writes directly to os.Stdout and reads
	// directly from os.Stdin

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
		os.Stdout.Close()
		asyncErr <- err
	}()

	have, err := exp.Expect(`.*7`)
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have, "1234567"))

	err = exp.Send("hello\n")
	qt.Assert(t, qt.IsNil(err))

	have, err = exp.Expect(`.*fa`)
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have, "890cafefa"))

	err = <-asyncErr
	qt.Assert(t, qt.IsNil(err))
}

func TestExpectSimulateSutOsPipeTimeout(t *testing.T) {
	exp, cleanup := expect.NewFilePipe(100*time.Millisecond,
		expect.MatchMaxDef)
	defer cleanup()

	// We simulate a sut that writes directly to os.Stdout and reads
	// directly from os.Stdin

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
		os.Stdout.Close()
		asyncErr <- err
	}()

	have, err := exp.Expect(`.*HELLO`)
	qt.Assert(t, qt.ErrorIs(err, expect.ErrTimeout))
	qt.Assert(t, qt.IsTrue(strings.HasPrefix(have, "12345")),
		qt.Commentf("have: %s", have))

	_, err = exp.Drain()
	qt.Assert(t, qt.IsNil(err))

	err = <-asyncErr
	qt.Assert(t, qt.IsNil(err))
}
