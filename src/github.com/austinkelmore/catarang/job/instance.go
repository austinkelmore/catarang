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

type JobStep struct {
	Log    ulog.StepLog
	Action plugin.Runner
}

// Instance a single run of a job
type Instance struct {
	StartTime time.Time
	EndTime   time.Time
	// todo: akelmore - move the Instance Num out of the Instance so it can't be changed (meta data on the job is a better place)
	Num       int
	JobConfig Config

	Steps []JobStep

	Status Status
	Error  error
}

// Start Entry point for the instance
func (i *Instance) Start() {
	i.StartTime = time.Now()
	defer func() { i.EndTime = time.Now() }()
	i.Status = RUNNING

	err := os.MkdirAll(i.JobConfig.LocalPath, 0777)
	if err != nil {
		i.Error = errors.Wrapf(err, "Can't create directory for job at path \"%s\"", i.JobConfig.LocalPath)
		i.Status = FAILED
		return
	}

	path, err := filepath.Abs(i.JobConfig.LocalPath)
	if err != nil {
		i.Error = errors.Wrapf(err, "Can't get absolute path from \"%s\"", i.JobConfig.LocalPath)
		i.Status = FAILED
		return
	}

	for index, _ := range i.JobConfig.Steps {
		step := JobStep{Action: i.JobConfig.Steps[index].Action}
		step.Log.Name = i.JobConfig.Steps[index].Name
		step.Log.WorkingDir = path
		i.Steps = append(i.Steps, step)

		if err = i.Steps[index].Action.Run(&i.Steps[index].Log); err != nil {
			i.Error = errors.Wrapf(err, "Couldn't run step index %v with action name %s", index, i.Steps[index].Action.GetName())
			i.Status = FAILED
			return
		}
	}

	i.Status = SUCCESSFUL
}
