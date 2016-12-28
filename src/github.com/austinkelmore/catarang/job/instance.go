package job

import (
	"os"
	"path/filepath"
	"time"

	"github.com/austinkelmore/catarang/plugin"
	"github.com/austinkelmore/catarang/ulog"
	"github.com/pkg/errors"
)

// Status The job instance's status
type Status int

const (
	// INITIALIZED The job instance is initialized, but hasn't started running yet
	INITIALIZED Status = iota
	// RUNNING The job instance is currently running
	RUNNING
	// FAILED The job instance failed
	FAILED
	// SUCCESSFUL The job instance was successful
	SUCCESSFUL
)

// todo: akelmore - get generate stringer working with Status instead of hard coding it
func (s Status) String() string {
	switch s {
	case INITIALIZED:
		return "Initialized"
	case RUNNING:
		return "Running"
	case FAILED:
		return "Failed"
	case SUCCESSFUL:
		return "Successful"
	default:
		return "Unknown, not in String() function"
	}
}

// InstJobStep is a distinct use of a plugin to do a single step or action within a job
// todo: akelmore - figure out why InstJobStep is different from Step
type InstJobStep struct {
	Log    ulog.StepLog
	Action plugin.JobStep
}

// Instance is a single run of a job
type Instance struct {
	StartTime time.Time
	EndTime   time.Time

	InstConfig Config
	Steps      []InstJobStep

	Status Status
	Error  error
}

// Start is an entry point for the instance
func (i *Instance) Start() {
	i.StartTime = time.Now()
	defer func() { i.EndTime = time.Now() }()
	i.Status = RUNNING

	err := os.MkdirAll(i.InstConfig.Data.LocalPath, 0777)
	if err != nil {
		i.Error = errors.Wrapf(err, "can't create directory for job at path \"%s\"", i.InstConfig.Data.LocalPath)
		i.Status = FAILED
		return
	}

	path, err := filepath.Abs(i.InstConfig.Data.LocalPath)
	if err != nil {
		i.Error = errors.Wrapf(err, "can't get absolute path from \"%s\"", i.InstConfig.Data.LocalPath)
		i.Status = FAILED
		return
	}

	for index := range i.InstConfig.Steps {
		step := InstJobStep{Action: i.InstConfig.Steps[index].Plugin}
		step.Log.Name = i.InstConfig.Steps[index].Plugin.GetName()
		step.Log.WorkingDir = path
		i.Steps = append(i.Steps, step)

		if err = i.Steps[index].Action.Run(i.InstConfig.Data, &i.Steps[index].Log); err != nil {
			i.Error = errors.Wrapf(err, "Couldn't run step index %v with action name %s", index, i.Steps[index].Action.GetName())
			i.Status = FAILED
			return
		}
	}

	i.Status = SUCCESSFUL
}
