package plugin

import (
	"github.com/austinkelmore/catarang/jobdata"
	"github.com/austinkelmore/catarang/cmd"
)

// JobStep is the interface that defines the least amount needed to
// get a plugin working as part of a job
type JobStep interface {
	Run(job jobdata.MetaData, logger *cmd.Log) error
	GetName() string
}
