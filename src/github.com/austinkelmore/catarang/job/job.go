package job

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/austinkelmore/catarang/cmd"
	"github.com/austinkelmore/catarang/jobdata"
	"github.com/austinkelmore/catarang/step"
	"github.com/pkg/errors"
)

type Template struct {
	Steps    []step.Template
	MetaData jobdata.MetaData
}

// todo: akelmore - can creating a job fail?
func New(t Template) (Job, error) {
	j := Job{}
	j.Template = t
	j.JobLog.Name = "job_log"

	return j, nil
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
	JobLog  cmd.Log // log for the job outside of instances of it being run (used for polling)
	History []InstData

	Template Template
}

// UpdateTemplate updates the configuration information for the job. This needs to be done every time in case
// it has changed from the previous run.
func (j *Job) UpdateTemplate() error {
	// first update the repository that this job is based on
	// todo: akelmore - make updating the config based on what type of SCM this is instead of hard using git
	// todo: akelmore - make updating the config use a logging that can be passed back to the job
	// cmd := exec.Command("git", "clone", j.MetaData.Origin, j.MetaData.LocalPath)
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// if err := cmd.Run(); err != nil {
	// 	return errors.Wrapf(err, "couldn't git clone origin \"%s\" to local path \"%s\"", j.MetaData.Origin, j.MetaData.LocalPath)
	// }

	file := filepath.Join(j.Template.MetaData.LocalPath, ".catarang.json")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.Wrapf(err, "error reading in config file \"%s\"", file)
	}

	if err = json.Unmarshal(data, &j.Template); err != nil {
		return errors.Wrapf(err, "error unmarshaling json from \"%s\"", file)
	}

	return nil
}

// GetName returns the name of the job
func (j Job) GetName() string {
	return j.Template.MetaData.Name
}

// Run is the entry point to start the job
func (j *Job) Run() {
	j.Template.MetaData.TimesRun++

	j.UpdateTemplate()

	inst := NewInstance(j.Template)
	// todo: akelmore - handle nil inst
	j.History = append(j.History, InstData{Num: j.Template.MetaData.TimesRun, Inst: *inst})

	inst.Start()
}

// Clean will delete all local job data
func (j *Job) Clean() {
	log.Printf("Cleaning Job \"%s\" from local path \"%s\"\n", j.GetName(), j.Template.MetaData.LocalPath)
	os.RemoveAll(j.Template.MetaData.LocalPath)
}
