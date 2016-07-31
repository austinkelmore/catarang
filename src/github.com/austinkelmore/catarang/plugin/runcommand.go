package plugin

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/austinkelmore/catarang/jobcommand"
	"github.com/austinkelmore/catarang/ulog"
)

type RunCommand struct {
	Commands  []string
	ConfigLoc string
}

func (r *RunCommand) Run(logger *ulog.StepLog) bool {
	if err := r.updateRunCommand(logger); err != nil {
		// todo: akelmore - add some sort of logging that there's an error syncing the run command
		log.Println(err.Error())
		return false
	}
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

func (r *RunCommand) GetName() string {
	return "run command"
}

func (r *RunCommand) updateRunCommand(logger *ulog.StepLog) error {
	// read in the config file's build command
	path := r.ConfigLoc
	if path == "" {
		path = ".catarang.json"
	}
	path = filepath.Join(logger.WorkingDir, path)

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.New("Error reading build config file: " + path)
	}

	cmd := jobcommand.Cmd{}
	err = json.Unmarshal(file, &cmd)
	if err != nil {
		return errors.New("Error reading JSON from build config file: " + path)
	}

	r.Commands = cmd.ExecCommands

	return nil
}
