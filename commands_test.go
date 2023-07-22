package otium

import (
	"io"
	"os"
	"testing"

	"github.com/go-quicktest/qt"
)

func setupTestCmdVariables(t *testing.T) (*os.File, func()) {
	stdoutRd, stdoutWr, err := os.Pipe()
	qt.Assert(t, qt.IsNil(err))
	oldStdout := os.Stdout
	os.Stdout = stdoutWr
	cleanup := func() { os.Stdout = oldStdout }
	return stdoutRd, cleanup
}

func TestCmdVariablesEmpty(t *testing.T) {
	pcd := NewProcedure(ProcedureOpts{})
	stdoutRd, cleanup := setupTestCmdVariables(t)
	defer cleanup()

	cmdVariables(pcd)

	os.Stdout.Close()
	buf, err := io.ReadAll(stdoutRd)
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(string(buf), "no variables\n"))
}

func TestCmdVariablesNotEmpty(t *testing.T) {
	pcd := NewProcedure(ProcedureOpts{})
	pcd.Put("fruit", "mango")
	pcd.Put("amount", "100")
	stdoutRd, cleanup := setupTestCmdVariables(t)
	defer cleanup()

	cmdVariables(pcd)

	os.Stdout.Close()
	buf, err := io.ReadAll(stdoutRd)
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(string(buf), "amount (): 100\nfruit (): mango\n"))
}
