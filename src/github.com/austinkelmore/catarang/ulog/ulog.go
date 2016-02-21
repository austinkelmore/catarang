package ulog

import (
	"bytes"
	"os/exec"
)

type CmdType int

const (
	CmdTypeNone CmdType = 0
	CmdTypeOut          = 1
	CmdTypeErr          = 2
)

type WriteSection struct {
	Len int
	Src CmdType
}

type CmdWriter struct {
	cmd *Cmd
	Src CmdType
}

func (c *CmdWriter) Write(p []byte) (n int, err error) {
	var b bytes.Buffer
	n, err = b.Write(p)
	c.cmd.Str += b.String()

	sect := &c.cmd.Sect
	// if the source of the last write is the same as this one, increase the size
	// of the recorded write instead of creating a new one
	if len(*sect) > 0 && (*sect)[len(*sect)-1].Src == c.Src {
		(*sect)[len(*sect)-1].Len += n
	} else {
		*sect = append(*sect, WriteSection{Len: n, Src: c.Src})
	}

	return n, err
}

type Cmd struct {
	Cmd  *exec.Cmd
	Out  CmdWriter
	Err  CmdWriter
	Str  string
	Sect []WriteSection // this keeps track of what parts of Str were written to by stdout and stderr
}

func (c *Cmd) Run() error {
	return c.Cmd.Run()
}

type Commands struct {
	Cmds []Cmd
}

func (c *Commands) New(name string, arg ...string) *Cmd {
	c.Cmds = append(c.Cmds, Cmd{})
	cmd := &c.Cmds[len(c.Cmds)-1]
	cmd.Out = CmdWriter{cmd: cmd, Src: CmdTypeOut}
	cmd.Err = CmdWriter{cmd: cmd, Src: CmdTypeErr}
	cmd.Cmd = exec.Command(name, arg...)
	cmd.Cmd.Stdout = &cmd.Out
	cmd.Cmd.Stderr = &cmd.Err
	return cmd
}

type Job struct {
	Name string
	Cmds Commands
}

func NewJob(name string) *Job {
	j := Job{Name: name}
	return &j
}
