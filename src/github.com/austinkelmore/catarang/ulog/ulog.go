package ulog

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

type CmdWriter struct {
	Lines *[]Buffer
	Src   CmdType
}

func (c *CmdWriter) Write(p []byte) (n int, err error) {
	*c.Lines = append(*c.Lines, Buffer{Src: c.Src})
	line := &(*c.Lines)[len(*c.Lines)-1]
	return line.Buf.Write(p)
}

type Cmd struct {
	Cmd *exec.Cmd
	Out CmdWriter
	Err CmdWriter
	Log []Buffer
}

func (c *Cmd) Run() error {
	return c.Cmd.Run()
}

func (c *Cmd) Bytes() []byte {
	var bytes []byte
	for _, log := range c.Log {
		bytes = append(bytes, log.Buf.Bytes()...)
	}
	return bytes
}

func (c *Cmd) init(name string, arg ...string) {
	c.Out = CmdWriter{Lines: &c.Log, Src: CmdStdOut}
	c.Err = CmdWriter{Lines: &c.Log, Src: CmdStdErr}
	c.Cmd = exec.Command(name, arg...)
	c.Cmd.Stdout = &c.Out
	c.Cmd.Stderr = &c.Err
}

// todo: akelmore - figure out how to make it so that the commands only have access to the section they should
// instead of the entire command log
func New(cmds *[]Cmd, name string, arg ...string) *Cmd {
	*cmds = append(*cmds, Cmd{})
	cmd := &(*cmds)[len(*cmds)-1]
	cmd.init(name, arg...)
	return cmd
}

type Job struct {
	Name string
	Cmds []Cmd
}
