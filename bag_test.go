package otium

import (
	"testing"

	"github.com/go-quicktest/qt"
)

func TestBag_GetNoAsk(t *testing.T) {
	sut := Bag{
		bag: map[string]string{"existing": "yes"},
	}

	val, err := sut.GetNoAsk("existing")
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(val, "yes"))

	_, err = sut.GetNoAsk("non-existing")
	qt.Assert(t, qt.ErrorMatches(err, `key not found.*`))
}
