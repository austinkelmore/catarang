package scm

import "github.com/austinkelmore/catarang/ulog"

// SCMer interface for the source control modules
type SCMer interface {
	FirstTimeSetup(cmds *ulog.Commands) error
	Poll(cmds *ulog.Commands) (bool, error)
	UpdateExisting(cmds *ulog.Commands) error
	LocalRepoPath() string
}
