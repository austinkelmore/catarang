package plugin

import "github.com/austinkelmore/catarang/cmd"

// JobStep is the interface that defines the least amount needed to
// get a plugin working as part of a job
type JobStep interface {
	Run(logger *cmd.Log) error
	GetName() string
}
