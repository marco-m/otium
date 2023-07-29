package otium

import (
	"testing"
	"time"

	"github.com/go-quicktest/qt"
	"github.com/peterh/liner"

	"github.com/marco-m/otium/expect"
)

func TestBag_Get_ExistingVar(t *testing.T) {
	const key = "existing"
	sut := NewBag()
	sut.bag[key] = Variable{val: "yes", set: true}

	val, err := sut.Get(key)

	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(val, "yes"))
}

func TestBag_Get_NonExistingVar(t *testing.T) {
	sut := NewBag()

	_, err := sut.Get("non-existing")

	qt.Assert(t, qt.ErrorMatches(err, `key not found.*`))
}

func TestBag_Get_NotSetVar(t *testing.T) {
	const key = "existing-but-not-yet-asked"
	sut := NewBag()
	sut.bag[key] = Variable{val: "yes"}

	_, err := sut.Get(key)

	qt.Assert(t, qt.ErrorMatches(err, `key not found.*`))
}

func TestBag_Put(t *testing.T) {
	const key = "fruit"
	sut := NewBag()

	sut.Put(key, "banana")

	fruit, err := sut.Get(key)
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(fruit, "banana"))
}

func TestBag_AskKeyAlreadySet(t *testing.T) {
	const key = "fruit"
	sut := NewBag()
	sut.Put(key, "banana")

	val, err := sut.ask(key, nil)

	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(val, "banana"))
}

func TestBag_AskKeyInteractive(t *testing.T) {
	exp, cleanup := expect.NewFilePipe(100*time.Millisecond, expect.MatchMaxDef)
	defer cleanup()
	const key = "fruit"
	sut := NewBag()
	sut.bag[key] = Variable{Name: "key", Desc: "Your fruit for breakfast"}

	term := liner.NewLiner()
	// Restore terminal to previous mode, super important.
	defer term.Close()

	var val string
	asyncErr := make(chan error)
	go func() {
		var err error
		val, err = sut.ask(key, term)
		//os.Stdout.Close()
		asyncErr <- err
	}()

	have, err := exp.Expect(`.*Enter Your fruit for breakfast \(set fruit <value>\) or '\?' for help\n\(input\)>> `)
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(have, "(input) Enter Your fruit for breakfast (set fruit <value>) or '?' for help\n(input)>> "))

	err = exp.Send("set fruit banana\n")
	qt.Assert(t, qt.IsNil(err))

	err = <-asyncErr
	qt.Assert(t, qt.IsNil(err))

	qt.Assert(t, qt.Equals(val, "banana"))
}

func TestInputCompleter(t *testing.T) {
	type testCase struct {
		name string
		key  string
		line string
		want []string
	}

	run := func(t *testing.T, tc testCase) {
		sut := makeInputCompleter(tc.key)

		have := sut(tc.line)

		qt.Assert(t, qt.DeepEquals(have, tc.want))
	}

	testCases := []testCase{
		{
			name: "empty line expands to all commands",
			key:  "fruit",
			want: []string{"help", "?", "back", "set"},
		},
		{
			name: "non-matching line expands to nothing",
			key:  "fruit",
			line: "x",
			want: []string{},
		},
		{
			name: "s expands to set",
			key:  "fruit",
			line: "s",
			want: []string{"set"},
		},
		{
			name: "set expands also to key",
			key:  "fruit",
			line: "set",
			want: []string{"set fruit "},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { run(t, tc) })
	}
}
