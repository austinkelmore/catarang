package plugin

import "github.com/austinkelmore/catarang/ulog"
import "github.com/austinkelmore/catarang/jobdata"

type Runner interface {
	Run(job jobdata.Data, logger *ulog.StepLog) error
	GetName() string
}
