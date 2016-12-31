package job

import (
	"os"
	"path/filepath"
	"time"

	"github.com/austinkelmore/catarang/cmd"
	"github.com/austinkelmore/catarang/jobdata"
	"github.com/austinkelmore/catarang/plugin"
	"github.com/austinkelmore/catarang/plugin/scm"
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
	Log    cmd.Log
	Action plugin.JobStep
}

// Instance is a single run of a job
type Instance struct {
	StartTime time.Time
	EndTime   time.Time

	MetaData jobdata.MetaData
	Steps    []InstJobStep

	Status Status
	Error  error
}

// todo: akelmore - rename from NewInstance
// todo: akelmore - handle an error being thrown somewhere in here and return it
func NewInstance(t Template) *Instance {
	i := Instance{MetaData: t.MetaData}

	path, err := filepath.Abs(i.MetaData.LocalPath)
	if err != nil {
		i.Error = errors.Wrapf(err, "can't get absolute path from \"%s\"", i.MetaData.LocalPath)
		i.Status = FAILED
		return nil
	}

	for _, step := range t.Steps {
		instStep := InstJobStep{Action: step.Plugin()}
		instStep.Log.Name = step.Plugin().GetName()
		instStep.Log.WorkingDir = path
		i.Steps = append(i.Steps, instStep)
	}
	return &i
}

// Start is an entry point for the instance
func (i *Instance) Start() {
	i.StartTime = time.Now()
	defer func() { i.EndTime = time.Now() }()
	i.Status = RUNNING

	err := os.MkdirAll(i.MetaData.LocalPath, 0777)
	if err != nil {
		i.Error = errors.Wrapf(err, "can't create directory for job at path \"%s\"", i.MetaData.LocalPath)
		i.Status = FAILED
		return
	}

	// for i := range job.JobConfig.Steps {
	// 	if scm, ok := job.JobConfig.Steps[i].Plugin.(scm.SCMer); ok {
	// 		if err := scm.SetOrigin(origin); err != nil {
	// 			return nil, errors.Wrapf(err, "couldn't Set Origin on source control manager %s", job.JobConfig.Steps[i].Plugin.GetName())
	// 		}
	// 	}
	// }

	// todo: akelmore - remove hard coded
	if i.MetaData.TimesRun == 1 {
		for index := range i.Steps {
			if scm, ok := i.Steps[index].Action.(scm.SCMer); ok {
				scm.FirstTimeSetup(&i.Steps[index].Log)
			}
		}
	}

	for index := range i.Steps {
		if err = i.Steps[index].Action.Run(i.MetaData, &i.Steps[index].Log); err != nil {
			i.Error = errors.Wrapf(err, "couldn't run step index %v with action name %s", index, i.Steps[index].Action.GetName())
			i.Status = FAILED
			return
		}
	}

	i.Status = SUCCESSFUL
}
