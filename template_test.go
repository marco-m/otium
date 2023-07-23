package otium

import (
	"bytes"
	"testing"

	"github.com/go-quicktest/qt"
)

func TestRender(t *testing.T) {
	type testCase struct {
		name string
		text string
		bag  map[string]Variable
		want string
	}

	run := func(t *testing.T, tc testCase) {
		var buf bytes.Buffer
		err := renderTemplate(&buf, tc.text, tc.bag)
		qt.Assert(t, qt.IsNil(err))
		qt.Assert(t, qt.Equals(buf.String(), tc.want))
	}

	testCases := []testCase{
		{
			name: "uppercase",
			text: "Hello {{.Name}}!",
			bag:  map[string]Variable{"Name": {val: "Joe"}},
			want: "Hello Joe!",
		},
		{
			name: "lowercase",
			text: "Hello {{.name}}!",
			bag:  map[string]Variable{"name": {val: "Joe"}},
			want: "Hello Joe!",
		},
		{
			// Decision point.
			// Default Go behavior is to keep going and just to write
			// <no value>. Is this OK or do we want to return an error?
			name: "no value",
			text: "Hello {{.Name}}!",
			bag:  map[string]Variable{},
			want: "Hello <no value>!",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
