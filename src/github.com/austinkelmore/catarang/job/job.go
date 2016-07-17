package job

import (
	"log"
	"time"

	"github.com/austinkelmore/catarang/plugin"
	"github.com/austinkelmore/catarang/scm"
	"github.com/austinkelmore/catarang/step"
	"github.com/austinkelmore/catarang/ulog"
)

type StepData struct {
	Action step.Runner
	Name   string
}

// Config is the where a job keeps all of the necessary
// information for running an instance of itself
type Config struct {
	Name            string
	LocalPath       string
	BuildConfigPath string

	Steps []StepData
	// todo: akelmore - don't use git instead of an interface for scms. Fix saving/loading interface
	SourceControl *scm.Git
}

// Job is the way in which you can run commands on the server or nodes
// it's the main reason this whole build system is created - to run jobs
type Job struct {
	Enabled bool
	JobLog  ulog.StepLog // log for the job outside of instances of it being run (used for polling)
	History []Instance

	JobConfig Config

	// todo: akelmore - move CompletedSetup to a global space for the individual plugin
	CompletedSetup bool
}

// NewJob creates a new job and initializes it with necessary values
func NewJob(name string, origin string) Job {
	job := Job{Enabled: true, CompletedSetup: false}
	job.JobConfig.Name = name

	job.JobLog.Name = "job_log"

	// todo: akelmore - configure local path
	job.JobConfig.LocalPath = "jobs/" + name + "/"
	job.JobConfig.SourceControl = scm.NewGit(origin)

	job.JobConfig.Steps = append(job.JobConfig.Steps, StepData{Name: "git", Action: job.JobConfig.SourceControl})
	cmd := plugin.RunCommand{}
	job.JobConfig.Steps = append(job.JobConfig.Steps, StepData{Name: "run command", Action: &cmd})

	return job
}

func (j Job) GetName() string {
	return j.JobConfig.Name
}

func (j *Job) getLastInst() *Instance {
	if len(j.History) == 0 {
		return nil
	}

	return &j.History[len(j.History)-1]
}

func (j *Job) Run() {
	log.Println("Running job:", j.GetName())

	// create a new instance and start it up
	j.History = append(j.History, NewInstance(j.JobConfig, len(j.History)))
	inst := &j.History[len(j.History)-1]

	inst.Start()
	inst.EndTime = time.Now()

	log.Println("Job finished:", j.GetName())
}
