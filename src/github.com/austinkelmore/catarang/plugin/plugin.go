package plugin

import "github.com/austinkelmore/catarang/cmd"

// JobStep is the interface that defines the least amount needed to
// get a plugin working as part of a job
type JobStep interface {
	Run(logger *cmd.Log) error
	GetName() string
}

// SCM is an interface for the source control modules
type SCM interface {
	FirstTimeSetup(logger *cmd.Log) error
	Poll(logger *cmd.Log) (bool, error)
	UpdateExisting(logger *cmd.Log) error
	SetOrigin(origin string) error
	JobStep
}
