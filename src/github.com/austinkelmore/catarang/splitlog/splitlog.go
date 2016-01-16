package splitlog

import (
	"bytes"
	"os/exec"
)

type CmdType int

const (
	CmdStdOut CmdType = 0
	CmdStdErr         = 1
)

type Buffer struct {
	Buf bytes.Buffer
	Src CmdType
}

type Writer struct {
	Lines *[]Buffer
	Src   CmdType
}

func (c *Writer) Write(p []byte) (n int, err error) {
	*c.Lines = append(*c.Lines, Buffer{Src: c.Src})
	line := &(*c.Lines)[len(*c.Lines)-1]
	return line.Buf.Write(p)
}

type CmdLog struct {
	Cmd *exec.Cmd
	Out Writer
	Err Writer
	Log []Buffer
}

func (c *CmdLog) Run() error {
	return c.Cmd.Run()
}

func (c *CmdLog) Bytes() []byte {
	var bytes []byte
	for _, log := range c.Log {
		bytes = append(bytes, log.Buf.Bytes()...)
	}
	return bytes
}

func (c *CmdLog) init(name string, arg ...string) {
	c.Out = Writer{Lines: &c.Log, Src: CmdStdOut}
	c.Err = Writer{Lines: &c.Log, Src: CmdStdErr}
	c.Cmd = exec.Command(name, arg...)
	c.Cmd.Stdout = &c.Out
	c.Cmd.Stderr = &c.Err
}

// todo: akelmore - figure out how to make it so that the commands only have access to the section they should
// instead of the entire command log
func New(cmds []CmdLog, name string, arg ...string) *CmdLog {
	cmds = append(cmds, CmdLog{})
	cmd := &cmds[len(cmds)-1]
	cmd.init(name, arg...)
	return cmd
}

type JobLog struct {
	Name string
	Cmds []CmdLog
}
