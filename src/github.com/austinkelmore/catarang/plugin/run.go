package plugin

import "github.com/austinkelmore/catarang/ulog"
import "github.com/austinkelmore/catarang/jobdata"

// Runner is the interface that defines the least amount needed to
// get a plugin working as part of a job
type Runner interface {
	Run(job jobdata.Data, logger *ulog.StepLog) error
	GetName() string
}
