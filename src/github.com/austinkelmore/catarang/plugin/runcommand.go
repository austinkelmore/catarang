package plugin

import (
	"strings"

	"github.com/austinkelmore/catarang/jobdata"
	"github.com/austinkelmore/catarang/ulog"
	"github.com/pkg/errors"
)

type RunCommand struct {
	Commands []string
}

func (r *RunCommand) Run(job jobdata.Data, logger *ulog.StepLog) error {
	for _, cmd := range r.Commands {
		fields := strings.Fields(cmd)
		if len(fields) > 0 {
			exec := logger.New(fields[0], fields[1:]...)
			if err := exec.Run(); err != nil {
				return errors.Wrapf(err, "Error running the command \"%s\"", cmd)
			}
		}
	}
	return nil
}

func (r RunCommand) GetName() string {
	return "runcommand"
}
