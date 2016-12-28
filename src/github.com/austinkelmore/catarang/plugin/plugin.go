package plugin

import (
	"github.com/austinkelmore/catarang/jobdata"
	"github.com/austinkelmore/catarang/ulog"
)

// JobStep is the interface that defines the least amount needed to
// get a plugin working as part of a job
type JobStep interface {
	Run(job jobdata.Data, logger *ulog.StepLog) error
	GetName() string
}
