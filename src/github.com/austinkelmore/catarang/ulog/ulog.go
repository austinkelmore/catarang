package ulog

import (
	"bytes"
	"os/exec"
)

// CmdType is which target was used for the cmd's output (either StdErr or StdOut)
type CmdType int

const (
	// CmdTypeNone is the default value, but should be changed to Out or Err
	CmdTypeNone CmdType = 0
	// CmdTypeOut is used when the cmd would output to StdOut
	CmdTypeOut = 1
	// CmdTypeErr is used when the cmd would output to StdErr
	CmdTypeErr = 2
)

// WriteSection is a section of a log that knows how long it is and what type it is
type WriteSection struct {
	Len int
	Src CmdType
}

// CmdWriter writes the cmd's output to the log and remembers what type of output it was
type CmdWriter struct {
	cmd *Cmd
	Src CmdType
}

// Write writes the bytes to the log
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

// Cmd holds all of the information needed to record the output of a cmd's process in the format that is needed
type Cmd struct {
	Cmd  *exec.Cmd
	Out  CmdWriter
	Err  CmdWriter
	Str  string
	Sect []WriteSection // this keeps track of what parts of Str were written to by stdout and stderr
}

// Run is the entry point for starting the cmd's process
func (c *Cmd) Run() error {
	return c.Cmd.Run()
}

// StepLog is a log that contains an array of commands and metadata about where to run the commands
type StepLog struct {
	Name       string
	WorkingDir string
	Cmds       []Cmd
}

// New creates a new log and returns the cmd structure
func (s *StepLog) New(name string, arg ...string) *Cmd {
	s.Cmds = append(s.Cmds, Cmd{})
	cmd := &s.Cmds[len(s.Cmds)-1]
	cmd.Out = CmdWriter{cmd: cmd, Src: CmdTypeOut}
	cmd.Err = CmdWriter{cmd: cmd, Src: CmdTypeErr}
	cmd.Cmd = exec.Command(name, arg...)
	cmd.Cmd.Dir = s.WorkingDir
	cmd.Cmd.Stdout = &cmd.Out
	cmd.Cmd.Stderr = &cmd.Err

	return cmd
}
