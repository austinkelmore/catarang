package job

import (
	"log"
	"time"

	"github.com/austinkelmore/catarang/scm"
	"github.com/austinkelmore/catarang/ulog"
)

// Config is the where a job keeps all of the necessary
// information for running an instance of itself
type Config struct {
	LocalPath       string
	BuildConfigPath string
	// todo: akelmore - don't use git instead of an interface for scms. Fix saving/loading interface
	SourceControl *scm.Git
}

// Job is the way in which you can run commands on the server or nodes
// it's the main reason this whole build system is created - to run jobs
type Job struct {
	Name           string
	Enabled        bool
	CurConfig      Config
	CompletedSetup bool
	History        []Instance
}

// NewJob creates a new job and initializes it with necessary values
func NewJob(name string, origin string) Job {
	job := Job{Name: name, Enabled: true, CompletedSetup: false}

	// todo: akelmore - configure local path
	job.CurConfig.LocalPath = "jobs/" + name + "/"
	job.CurConfig.SourceControl = scm.NewGit(origin, job.CurConfig.LocalPath)
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

	// create a new instance and start it up
	j.History = append(j.History, NewInstance(j.CurConfig))
	inst := &j.History[len(j.History)-1]

	inst.Start(&j.CompletedSetup)
	inst.EndTime = time.Now()

	log.Println("Job finished:", j.Name)
}

func (j *Job) needsUpdate() bool {
	inst := j.getLastInst()
	// todo: akelmore - specify polling interval in config value
	if inst != nil && time.Since(inst.StartTime) < 30*time.Second {
		return false
	}
	log.Println("Running needsUpdate for:", j.Name)

	// todo: akelmore - make these use a real log
	logger := ulog.NewJob("poll")
	shouldRun, err := j.CurConfig.SourceControl.Poll(&logger.Cmds)
	if err != nil {
		log.Println(err.Error())
	}
	return shouldRun
}
