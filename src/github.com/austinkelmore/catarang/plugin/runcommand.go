package plugin

import (
	"strings"

	"github.com/austinkelmore/catarang/cmd"
	"github.com/pkg/errors"
)

// RunCommand is a plugin that runs a specified set of commands
type RunCommand struct {
	Commands []string
}

// Run is the entry point into the RunCommand plugin
func (r *RunCommand) Run(logger *cmd.Log) error {
	for _, cmd := range r.Commands {
		fields := strings.Fields(cmd)
		if len(fields) > 0 {
			exec := logger.New(fields[0], fields[1:]...)
			if err := exec.Run(); err != nil {
				return errors.Wrapf(err, "error running the command \"%s\"", cmd)
			}
		}
	}
	return nil
}

// GetName returns the name of the plugin
func (r RunCommand) GetName() string {
	return "runcommand"
}
