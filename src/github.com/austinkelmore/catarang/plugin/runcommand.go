package plugin

import (
	"strings"

	"github.com/austinkelmore/catarang/ulog"
)

type RunCommand struct {
	Commands []string
}

func (r *RunCommand) Run(logger *ulog.StepLog) bool {
	for _, cmd := range r.Commands {
		fields := strings.Fields(cmd)
		if len(fields) > 0 {
			exec := logger.New(fields[0], fields[1:]...)
			if err := exec.Run(); err != nil {
				return false
			}
		}
	}
	return true
}

func (r RunCommand) GetName() string {
	return "runcommand"
}
