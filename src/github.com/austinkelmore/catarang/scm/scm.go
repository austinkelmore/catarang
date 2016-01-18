package scm

import "github.com/austinkelmore/catarang/ulog"

// SCMer interface for the source control modules
type SCMer interface {
	FirstTimeSetup(cmds []ulog.Cmd) error
	Poll(cmds []ulog.Cmd) (bool, error)
	UpdateExisting(cmds []ulog.Cmd) error
	LocalRepoPath() string
}
