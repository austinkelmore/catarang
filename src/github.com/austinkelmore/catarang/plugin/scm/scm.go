package scm

import "github.com/austinkelmore/catarang/cmd"
import "github.com/austinkelmore/catarang/plugin"

// SCMer interface for the source control modules
type SCMer interface {
	FirstTimeSetup(logger *cmd.Log) error
	Poll(logger *cmd.Log) (bool, error)
	UpdateExisting(logger *cmd.Log) error
	SetOrigin(origin string) error
	plugin.JobStep
}
