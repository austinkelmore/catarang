package job

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/austinkelmore/catarang/plugin"
	"github.com/austinkelmore/catarang/ulog"
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
}

// Start Entry point for the instance
func (i *Instance) Start() {
	i.StartTime = time.Now()
	defer func() { i.EndTime = time.Now() }()
	i.Status = RUNNING

	// todo: akelmore - make the instance's work be captured in the job's logging
	err := os.MkdirAll(i.JobConfig.LocalPath, 0777)
	if err != nil {
		log.Println("FAILED! Can't create directory for job: " + i.JobConfig.LocalPath)
		i.Status = FAILED
		return
	}

	path, err := filepath.Abs(i.JobConfig.LocalPath)
	if err != nil {
		log.Println("FAILED! Can't get absolute path: " + err.Error())
		i.Status = FAILED
		return
	}

	for index, _ := range i.JobConfig.Steps {
		step := JobStep{Action: i.JobConfig.Steps[index].Action}
		step.Log.Name = i.JobConfig.Steps[index].Name
		step.Log.WorkingDir = path
		i.Steps = append(i.Steps, step)

		if i.Steps[index].Action.Run(&i.Steps[index].Log) == false {
			log.Printf("FAILED! %+v\n", i.Steps[index].Log)
			i.Status = FAILED
			return
		}
	}

	i.Status = SUCCESSFUL
}
