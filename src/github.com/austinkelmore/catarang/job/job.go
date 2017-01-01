package job

import (
	"log"
	"os"

	"github.com/austinkelmore/catarang/cmd"
	"github.com/austinkelmore/catarang/template"
)

// New creates a Job from a job template
func New(t template.Job) (Job, error) {
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

// JobData is all of the metadata about a job (that is outside of the template)
type JobData struct {
	ID       int
	Name     string
	TimesRun int
}

// Job is the way in which you can run commands on the server or nodes
// it's the main reason this whole build system is created - to run jobs.
type Job struct {
	JobLog  cmd.Log // log for the job outside of instances of it being run (used for polling)
	History []InstData
	JobData JobData

	Template template.Job
}

// GetName returns the name of the job
func (j Job) GetName() string {
	return j.JobData.Name
}

// Run is the entry point to start the job
func (j *Job) Run() {
	j.JobData.TimesRun++
	inst, err := NewInstance(j.Template)
	if err != nil {
		log.Printf("Error creating new instance: %v\n", err.Error())
		return
	}

	j.History = append(j.History, InstData{Num: j.JobData.TimesRun, Inst: *inst})
	j.History[len(j.History)-1].Inst.Start(j.JobData.TimesRun == 1)

	j.Template = j.History[len(j.History)-1].Inst.Template
}

// Clean will delete all local job data
func (j *Job) Clean() {
	log.Printf("Cleaning Job \"%s\" from local path \"%s\"\n", j.GetName(), j.Template.LocalPath)
	os.RemoveAll(j.Template.LocalPath)
}
