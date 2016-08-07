package job

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/austinkelmore/catarang/plugin/scm"
	"github.com/austinkelmore/catarang/ulog"
	"github.com/pkg/errors"
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
}

// NewJob creates a new job and initializes it with necessary values
// todo: akelmore - this can fail if we can't get the config, relay that to the user
func NewJob(name string, origin string) (*Job, error) {
	job := Job{}
	job.JobConfig.Name = name
	job.JobConfig.Origin = origin

	job.JobLog.Name = "job_log"
	job.JobConfig.LocalPath = filepath.Join("jobs/", name)

	if err := job.UpdateConfig(); err != nil {
		return nil, errors.Wrap(err, "Couldn't update the config.")
	}

	// go through and set the origin on every SCM so its aware of where we got this job from
	for i := range job.JobConfig.Steps {
		if scm, ok := job.JobConfig.Steps[i].Action.(scm.SCMer); ok {
			if err := scm.SetOrigin(origin); err != nil {
				return nil, errors.Wrapf(err, "Couldn't Set Origin on SCM %s", job.JobConfig.Steps[i].Action.GetName())
			}
		}
	}

	return &job, nil
}

func (j *Job) UpdateConfig() error {
	// first update the repository that this job is based on
	// todo: akelmore - make updating the config based on what type of SCM this is instead of hard using git
	// todo: akelmore - make updating the config use a logging that can be passed back to the job
	cmd := exec.Command("git", "clone", j.JobConfig.Origin, j.JobConfig.LocalPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "Couldn't git clone origin \"%s\" to local path \"%s\"", j.JobConfig.Origin, j.JobConfig.LocalPath)
	}

	file := filepath.Join(j.JobConfig.LocalPath, ".catarang.json")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.Wrapf(err, "Error reading in config file \"%s\"", file)
	}

	if err = json.Unmarshal(data, &j.JobConfig); err != nil {
		return errors.Wrapf(err, "Error unmarshaling json from \"%s\"", file)
	}

	return nil
}

func (j Job) GetName() string {
	return j.JobConfig.Name
}

func (j *Job) Run() {
	i := Instance{}
	i.JobConfig = j.JobConfig
	i.Num = len(j.History)
	j.History = append(j.History, i)
	inst := &j.History[len(j.History)-1]

	inst.Start()
}
