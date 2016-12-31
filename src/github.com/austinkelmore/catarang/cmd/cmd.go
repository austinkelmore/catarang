package cmd

import (
	"bytes"
	"os/exec"
)

// cmdType is which target was used for the cmd's output (either StdErr or StdOut)
type cmdType int

const (
	// cmdTypeNone is the default value, but should be changed to Out or Err
	cmdTypeNone cmdType = 0
	// cmdTypeOut is used when the cmd would output to StdOut
	cmdTypeOut = 1
	// cmdTypeErr is used when the cmd would output to StdErr
	cmdTypeErr = 2
)

// writeSection is a section of a log that knows how long it is and what type it is
type writeSection struct {
	Len int
	Src cmdType
}

// writer writes the cmd's output to the log and remembers what type of output it was
type writer struct {
	cmd *Cmd
	Src cmdType
}

// Write writes the bytes to the log
func (c *writer) Write(p []byte) (n int, err error) {
	var b bytes.Buffer
	n, err = b.Write(p)
	c.cmd.Str += b.String()

	sect := &c.cmd.Sect
	// if the source of the last write is the same as this one, increase the size
	// of the recorded write instead of creating a new one
	if len(*sect) > 0 && (*sect)[len(*sect)-1].Src == c.Src {
		(*sect)[len(*sect)-1].Len += n
	} else {
		*sect = append(*sect, writeSection{Len: n, Src: c.Src})
	}

	return n, err
}

// Cmd holds all of the information needed to record the output of a cmd's process in the format that is needed
type Cmd struct {
	Cmd  *exec.Cmd
	Out  writer
	Err  writer
	Str  string
	Sect []writeSection // this keeps track of what parts of Str were written to by stdout and stderr
}

// Run is the entry point for starting the cmd's process
func (c *Cmd) Run() error {
	return c.Cmd.Run()
}

// Log contains an array of commands and metadata about where to run the commands as well as the output from the cmds
type Log struct {
	Name       string
	WorkingDir string
	Cmds       []Cmd
}

// New creates a new log and returns the cmd structure
func (log *Log) New(name string, arg ...string) *Cmd {
	log.Cmds = append(log.Cmds, Cmd{})
	cmd := &log.Cmds[len(log.Cmds)-1]
	cmd.Out = writer{cmd: cmd, Src: cmdTypeOut}
	cmd.Err = writer{cmd: cmd, Src: cmdTypeErr}
	cmd.Cmd = exec.Command(name, arg...)
	cmd.Cmd.Dir = log.WorkingDir
	cmd.Cmd.Stdout = &cmd.Out
	cmd.Cmd.Stderr = &cmd.Err

	return cmd
}
