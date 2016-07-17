package step

import "github.com/austinkelmore/catarang/ulog"

type Runner interface {
	Run(logger *ulog.StepLog) bool
	GetName() string
}
