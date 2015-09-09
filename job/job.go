package job

import (
	"log"
	"time"

	"github.com/austinkelmore/catarang/scm"
)

type Status int

const (
	RUNNING Status = iota
	FAILED
	SUCCESSFUL
)

type Config struct {
	LocalPath       string
	BuildConfigPath string
	// todo: akelmore don't hard code the git scm in the job config
	Git scm.Git
}

type Job struct {
	Name           string
	Enabled        bool
	CurConfig      Config
	CompletedSetup bool
	History        []Instance
}

func CreateJob(name string, onlineRepo string, configPath string) Job {
	job := Job{Name: name, Enabled: true, CompletedSetup: false}

	job.CurConfig.BuildConfigPath = configPath
	// todo: akelmore - configure local path
	job.CurConfig.LocalPath = "jobs/" + name + "/"
	job.CurConfig.Git = scm.CreateGit(job.CurConfig.LocalPath, onlineRepo)
	return job
}

func (j *Job) getLastInst() *Instance {
	if len(j.History) == 0 {
		return nil
	}

	return &j.History[len(j.History)-1]
}

func (j *Job) NeedsRunning() bool {
	return len(j.History) == 0 || j.needsUpdate()
}

func (j *Job) Run() {
	log.Println("Running job:", j.Name)
	j.History = append(j.History, NewInstance(j.CurConfig))
	inst := &j.History[len(j.History)-1]

	if !j.CompletedSetup {
		log.Println("Running first time setup for:", j.Name)
		if err := inst.Config.Git.FirstTimeSetup(); err != nil {
			log.Println("Error in first time setup: " + err.Error())
			inst.Status = FAILED
		} else {
			j.CompletedSetup = true
		}
	} else {
		if err := inst.Config.Git.UpdateExisting(); err != nil {
			log.Println("Error updating an existing depot: " + err.Error())
			inst.Status = FAILED
		}
	}

	if inst.Status != FAILED {
		if err := inst.RunExecCommand(); err != nil {
			log.Println("Error running exec command: " + err.Error())
			inst.Status = FAILED
		}
	}

	inst.EndTime = time.Now()
}

func (j *Job) needsUpdate() bool {
	inst := j.getLastInst()
	// todo: akelmore - specify polling interval in config value
	if inst != nil && time.Since(inst.StartTime) < 30*time.Second {
		return false
	}
	log.Println("Running needsUpdate for:", j.Name)

	return j.CurConfig.Git.Poll()
}
