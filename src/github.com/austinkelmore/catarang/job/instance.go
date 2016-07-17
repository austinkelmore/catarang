package job

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/austinkelmore/catarang/step"
	"github.com/austinkelmore/catarang/ulog"
)

// Status The job instance's status
type Status int

const (
	// RUNNING The job instance is currently running
	RUNNING Status = iota
	// FAILED The job instance failed
	FAILED
	// SUCCESSFUL The job instance was successful
	SUCCESSFUL
)

type JobStep struct {
	Log    ulog.StepLog
	Action step.Runner
}

// Instance a single run of a job
type Instance struct {
	StartTime time.Time
	EndTime   time.Time
	Num       int
	JobConfig Config

	Steps []JobStep

	Status    Status
	Artifacts Artifact
}

// NewInstance Creates a new instance of a job (and copies off the current config)
// and starts the instance running
func NewInstance(config Config, instanceNum int) Instance {
	inst := Instance{StartTime: time.Now(), Status: RUNNING, JobConfig: config, Num: instanceNum}

	for _, s := range inst.JobConfig.Steps {
		jobStep := JobStep{Action: s.Action}
		jobStep.Log.Name = s.Name
		inst.Steps = append(inst.Steps, jobStep)
	}
	return inst
}

func (i *Instance) fail(reason string) {
	i.Status = FAILED
}

// Start Entry point for the instance
// todo: akelmore - i don't like passing a bool for the completedSetup, figure something better out
func (i *Instance) Start() {

	// todo: akelmore - handle folder creation error
	os.MkdirAll(i.JobConfig.LocalPath, 0777)

	for index, _ := range i.Steps {
		// todo: akelmore - handle filepath error
		path, _ := filepath.Abs(i.JobConfig.LocalPath)
		i.Steps[index].Log.WorkingDir = path
		if i.Steps[index].Action.Run(&i.Steps[index].Log) == false {
			log.Printf("FAILED! %+v\n", i.Steps[index].Log)
			i.fail("Runner failed.")
			return
		}
	}

	// todo: akelmore - make artifacting a step
	// for _, artifact := range i.BuildCommand.Artifacts {
	// 	if err := artifact.Save(i.JobConfig.Name, i.Num); err != nil {
	// 		log.Printf("Error saving artifact. %s\n", artifact.ToSave)
	// 		i.fail("Error saving artifact.")
	// 		return
	// 	}
	// }

	i.Status = SUCCESSFUL
}
