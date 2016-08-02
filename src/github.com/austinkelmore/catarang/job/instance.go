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
	Num       int
	JobConfig Config

	Steps []JobStep

	Status Status
}

// NewInstance Creates a new instance of a job (and copies off the current config)
// and starts the instance running
func NewInstance(config Config, instanceNum int) Instance {
	inst := Instance{JobConfig: config, Num: instanceNum}

	for _, s := range inst.JobConfig.Steps {
		jobStep := JobStep{Action: s.Action}
		jobStep.Log.Name = s.Name
		inst.Steps = append(inst.Steps, jobStep)
	}
	return inst
}

// Start Entry point for the instance
func (i *Instance) Start() {
	i.StartTime = time.Now()
	defer func() { i.EndTime = time.Now() }()
	i.Status = RUNNING

	err := os.MkdirAll(i.JobConfig.LocalPath, 0777)
	if err != nil {
		log.Println("FAILED! Can't create directory for job: " + i.JobConfig.LocalPath)
		i.Status = FAILED
		return
	}

	for index, _ := range i.Steps {
		path, err := filepath.Abs(i.JobConfig.LocalPath)
		if err != nil {
			log.Println("FAILED! Can't get absolute path: " + err.Error())
			i.Status = FAILED
			return
		}
		i.Steps[index].Log.WorkingDir = path
		if i.Steps[index].Action.Run(&i.Steps[index].Log) == false {
			log.Printf("FAILED! %+v\n", i.Steps[index].Log)
			i.Status = FAILED
			return
		}
	}

	i.Status = SUCCESSFUL
}
