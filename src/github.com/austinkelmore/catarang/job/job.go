package job

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/austinkelmore/catarang/plugin/scm"
	"github.com/austinkelmore/catarang/ulog"
)

// Config is the where a job keeps all of the necessary
// information for running an instance of itself
type Config struct {
	Name            string
	LocalPath       string
	BuildConfigPath string
	Origin          string

	Steps []Step
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
	job.JobConfig.Origin = origin

	job.JobLog.Name = "job_log"
	job.JobConfig.LocalPath = filepath.Join("jobs/", name)

	job.UpdateConfig()

	// go through and set the origin on every SCM so its aware of where we got this job from
	for i := range job.JobConfig.Steps {
		if scm, ok := job.JobConfig.Steps[i].Action.(scm.SCMer); ok {
			scm.SetOrigin(origin)
		}
	}

	return job
}

func (j *Job) UpdateConfig() {
	// first update the repository that this job is based on
	// todo: akelmore - make updating the config based on what type of SCM this is instead of hard using git
	// todo: akelmore - make updating the config use a logging that can be passed back to the job
	cmd := exec.Command("git", "clone", j.JobConfig.Origin, j.JobConfig.LocalPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	file := filepath.Join(j.JobConfig.LocalPath, ".catarang.json")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("Error reading in \"%s\"'s config file: %s\n", j.GetName(), file)
		return
	}

	if err = json.Unmarshal(data, &j.JobConfig); err != nil {
		log.Println("Error reading in", file)
		log.Println(err.Error())
	}
}

func (j Job) GetName() string {
	return j.JobConfig.Name
}

func (j *Job) Run() {
	j.History = append(j.History, NewInstance(j.JobConfig, len(j.History)))
	inst := &j.History[len(j.History)-1]

	inst.Start()
}
