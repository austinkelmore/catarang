package multilog

import (
	"bytes"
	"os/exec"
)

type CmdType int

const (
	CmdStdOut CmdType = 0
	CmdStdErr         = 1
)

type CmdOutput struct {
	Log bytes.Buffer
	Src CmdType
}

type CmdLog struct {
	Cmd   *exec.Cmd
	Lines []CmdOutput
}

// todo: akelmore - make this fit with Write by passing through the function to two different
// objects that implement Write, but set the Out/Err appropriately
func (m *CmdLog) HandleStdOut(p []byte) (n int, err error) {
	m.Lines = append(m.Lines, CmdOutput{Src: CmdStdOut})
	line := &m.Lines[len(m.Lines)-1]
	return line.Log.Write(p)
}

func (m *CmdLog) HandleStdErr(p []byte) (n int, err error) {
	m.Lines = append(m.Lines, CmdOutput{Src: CmdStdErr})
	line := &m.Lines[len(m.Lines)-1]
	return line.Log.Write(p)
}

func New(logger []*CmdLog, name string, arg ...string) *CmdLog {
	cmd := new(CmdLog)
	cmd.Cmd = exec.Command(name, arg...)
	cmd.Cmd.Stdout = cmd.HandleStdOut
	cmd.Cmd.Stderr = cmd.HandleStdErr
	logger = append(logger, cmd)
}
