package job

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/austinkelmore/catarang/jobdata"
	"github.com/austinkelmore/catarang/plugin/scm"
	"github.com/austinkelmore/catarang/ulog"
	"github.com/pkg/errors"
)

// Config is the the necessary information to run a job.
type Config struct {
	Data  jobdata.Data
	Steps []Step
}

// InstData is the combination of the instance data itself and the metadata associated with it.
// This includes information that would be useful to know about the run within the context of a history of a job.
type InstData struct {
	Inst Instance
	Num  int
}

// Job is the way in which you can run commands on the server or nodes
// it's the main reason this whole build system is created - to run jobs.
type Job struct {
	JobLog  ulog.StepLog // log for the job outside of instances of it being run (used for polling)
	History []InstData

	JobConfig Config
}

// NewJob creates a new job and initializes it with necessary values.
func NewJob(name string, origin string) (*Job, error) {
	job := Job{}
	job.JobConfig.Data = jobdata.Data{Name: name, Origin: origin, LocalPath: filepath.Join("jobs/", name)}
	job.JobLog.Name = "job_log"

	if err := job.UpdateConfig(); err != nil {
		return nil, errors.Wrap(err, "Couldn't update the config.")
	}

	// go through and set the origin on every SCM so its aware of where we got this job from
	for i := range job.JobConfig.Steps {
		if scm, ok := job.JobConfig.Steps[i].Action.(scm.SCMer); ok {
			if err := scm.SetOrigin(origin); err != nil {
				return nil, errors.Wrapf(err, "Couldn't Set Origin on source control manager %s", job.JobConfig.Steps[i].Action.GetName())
			}
		}
	}

	return &job, nil
}

// UpdateConfig updates the configuration information for the job. This needs to be done every time in case
// it has changed from the previous run.
func (j *Job) UpdateConfig() error {
	// first update the repository that this job is based on
	// todo: akelmore - make updating the config based on what type of SCM this is instead of hard using git
	// todo: akelmore - make updating the config use a logging that can be passed back to the job
	cmd := exec.Command("git", "clone", j.JobConfig.Data.Origin, j.JobConfig.Data.LocalPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "Couldn't git clone origin \"%s\" to local path \"%s\"", j.JobConfig.Data.Origin, j.JobConfig.Data.LocalPath)
	}

	file := filepath.Join(j.JobConfig.Data.LocalPath, ".catarang.json")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.Wrapf(err, "Error reading in config file \"%s\"", file)
	}

	if err = json.Unmarshal(data, &j.JobConfig); err != nil {
		return errors.Wrapf(err, "Error unmarshaling json from \"%s\"", file)
	}

	return nil
}

// GetName returns the name of the job
func (j Job) GetName() string {
	return j.JobConfig.Data.Name
}

// Run is the entry point to start the job
func (j *Job) Run() {
	j.JobConfig.Data.TimesRun++
	j.History = append(j.History, InstData{Num: j.JobConfig.Data.TimesRun})
	inst := &j.History[len(j.History)-1].Inst
	inst.JobConfig = j.JobConfig

	inst.Start()
}

// Clean will delete all local job data
func (j *Job) Clean() {
	log.Printf("Cleaning Job \"%s\" from local path \"%s\"\n", j.GetName(), j.JobConfig.Data.LocalPath)
	os.RemoveAll(j.JobConfig.Data.LocalPath)
}
