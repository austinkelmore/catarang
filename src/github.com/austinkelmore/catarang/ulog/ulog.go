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
	c.cmd.Buf += b.String()

	sect := &c.cmd.Sect
	// if the source of the last write is the same as this one, increase the size
	// of the recorded write instead of creating a new one
	if len(*sect) > 0 && (*sect)[len(*sect)-1].Src == c.Src {
		(*sect)[len(*sect)-1].Len += n
	} else {
		*sect = append(*sect, WriteSection{Len: n, Src: c.Src})
	}

	for _, callback := range *c.cmd.callbacks {
		callback.CmdLog(p)
	}

	return n, err
}

type CmdLogger interface {
	CmdLog(p []byte)
}

type Cmd struct {
	Cmd  *exec.Cmd
	Out  CmdWriter
	Err  CmdWriter
	Buf  string
	Sect []WriteSection // this keeps track of what parts of Buf were written to by stdout and stderr

	// todo: akelmore - is this the best way to do cmd logging with callbacks? i don't think so
	callbacks *[]CmdLogger
}

func (c *Cmd) Run() error {
	return c.Cmd.Run()
}

func (c *Cmd) init(callbacks *[]CmdLogger, name string, arg ...string) {
	c.Out = CmdWriter{cmd: c, Src: CmdTypeOut}
	c.Err = CmdWriter{cmd: c, Src: CmdTypeErr}
	c.Cmd = exec.Command(name, arg...)
	c.Cmd.Stdout = &c.Out
	c.Cmd.Stderr = &c.Err
	c.callbacks = callbacks
}

type Commands struct {
	Cmds []Cmd
	// todo: akelmore - don't store callbacks, make it a messaging system
	callbacks []CmdLogger
}

func (c *Commands) New(name string, arg ...string) *Cmd {
	c.Cmds = append(c.Cmds, Cmd{})
	cmd := &c.Cmds[len(c.Cmds)-1]
	cmd.init(&c.callbacks, name, arg...)
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

func (j *Job) AddCallback(logger CmdLogger) {
	j.Cmds.callbacks = append(j.Cmds.callbacks, logger)
}
