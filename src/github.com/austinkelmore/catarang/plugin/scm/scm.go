package scm

import "github.com/austinkelmore/catarang/ulog"
import "github.com/austinkelmore/catarang/plugin"

// SCMer interface for the source control modules
type SCMer interface {
	FirstTimeSetup(logger *ulog.StepLog) error
	Poll(logger *ulog.StepLog) (bool, error)
	UpdateExisting(logger *ulog.StepLog) error
	SetOrigin(origin string) error
	plugin.JobStep
}
