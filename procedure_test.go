package otium_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/marco-m/otium"
)

func TestProcedure_ExecuteWithZeroStepsFails(t *testing.T) {
	pcd := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Simple title",
		Desc:  `Simple description`,
	})

	err := pcd.Execute()

	assert.Error(t, err, "procedure has zero steps; want at least one")
}

func TestProcedure_ExecuteStepWithMissingFieldsFails(t *testing.T) {
	pcd := otium.NewProcedure(otium.ProcedureOpts{
		Title: "Simple title",
		Desc:  `Simple description`,
	})
	pcd.AddStep(&otium.Step{
		Title: "",
		Run:   nil,
	})

	err := pcd.Execute()

	assert.Error(t, err,
		"step (1) has empty Title\nstep (1 ) misses Run function")
}
