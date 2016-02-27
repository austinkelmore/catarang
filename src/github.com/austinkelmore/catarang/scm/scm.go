package scm

import "github.com/austinkelmore/catarang/ulog"

// SCMer interface for the source control modules
type SCMer interface {
	FirstTimeSetup(cmds *ulog.CmdList) error
	Poll(cmds *ulog.CmdList) (bool, error)
	UpdateExisting(cmds *ulog.CmdList) error
	LocalRepoPath() string
}
