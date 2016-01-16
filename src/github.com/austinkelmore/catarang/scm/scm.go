package scm

import "github.com/austinkelmore/catarang/splitlog"

// SCMer interface for the source control modules
type SCMer interface {
	FirstTimeSetup(cmds []splitlog.CmdLog) error
	Poll(cmds []splitlog.CmdLog) (bool, error)
	UpdateExisting(cmds []splitlog.CmdLog) error
	LocalRepoPath() string
}
