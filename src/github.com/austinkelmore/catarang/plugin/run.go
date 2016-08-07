package plugin

import "github.com/austinkelmore/catarang/ulog"

type Runner interface {
	Run(logger *ulog.StepLog) error
	GetName() string
}
