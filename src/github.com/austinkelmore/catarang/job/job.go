package job

import (
	"log"

	"github.com/austinkelmore/catarang/plugin"
	"github.com/austinkelmore/catarang/plugin/scm"
	"github.com/austinkelmore/catarang/ulog"
)

type StepData struct {
	Action plugin.Runner
	Name   string
}

// Config is the where a job keeps all of the necessary
// information for running an instance of itself
type Config struct {
	Name            string
	LocalPath       string
	BuildConfigPath string

	Steps []StepData
}

// Job is the way in which you can run commands on the server or nodes
// it's the main reason this whole build system is created - to run jobs
type Job struct {
	JobLog  ulog.StepLog // log for the job outside of instances of it being run (used for polling)
	History []Instance

	JobConfig Config

	// todo: akelmore - move CompletedSetup to a global space for the individual plugin
	CompletedSetup bool
}

// NewJob creates a new job and initializes it with necessary values
func NewJob(name string, origin string) Job {
	job := Job{CompletedSetup: false}
	job.JobConfig.Name = name

	job.JobLog.Name = "job_log"

	// todo: akelmore - configure local path
	job.JobConfig.LocalPath = "jobs/" + name + "/"

	job.JobConfig.Steps = append(job.JobConfig.Steps, StepData{Name: "git", Action: scm.NewGit(origin)})
	job.JobConfig.Steps = append(job.JobConfig.Steps, StepData{Name: "run command", Action: &plugin.RunCommand{}})
	job.JobConfig.Steps = append(job.JobConfig.Steps, StepData{Name: "artifact", Action: &plugin.Artifact{}})

	return job
}

func (j Job) GetName() string {
	return j.JobConfig.Name
}

func (j *Job) Run() {
	log.Println("Running job:", j.GetName())

	// create a new instance and start it up
	j.History = append(j.History, NewInstance(j.JobConfig, len(j.History)))
	inst := &j.History[len(j.History)-1]

	inst.Start()

	log.Println("Job finished:", j.GetName())
}
