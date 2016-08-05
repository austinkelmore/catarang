package job

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"

	"github.com/austinkelmore/catarang/plugin"
	"github.com/austinkelmore/catarang/pluginlist"
	"github.com/austinkelmore/catarang/ulog"
	"github.com/jeffail/gabs"
)

type StepData struct {
	Action plugin.Runner
	Name   string
}

// todo: akelmore - return the correct errors instead of returning nil
func (s *StepData) UnmarshalJSON(b []byte) error {
	// todo: akelmore - handle parse error
	parsed, _ := gabs.ParseJSON(b)
	plug := parsed.Search("plugin")
	if plug == nil {
		log.Println("Couldn't find \"plugin\" in StepData.")
		return nil
	}

	plugName, ok := plug.Data().(string)
	if !ok {
		log.Println("\"plugin\" was not a string in the config file.")
	}

	actionType, ok := pluginlist.Plugins()[plugName]
	if !ok {
		log.Printf("Couldn't find plugin of type \"%s\".\n", plugName)
		return nil
	}

	inter := reflect.New(actionType.Elem())
	s.Action = inter.Interface().(plugin.Runner)

	// shove the data inside the config into the plugin
	data := parsed.Search("data")
	if data == nil {
		log.Printf("No data blob in config associated with plugin \"%s\".\n", plugName)
		return nil
	}

	bytes := data.Bytes()
	err := json.Unmarshal(bytes, s.Action)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		return nil
	}

	return nil
}

// Config is the where a job keeps all of the necessary
// information for running an instance of itself
type Config struct {
	Name            string
	LocalPath       string
	BuildConfigPath string
	Origin          string

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
	job.JobConfig.Origin = origin

	job.JobLog.Name = "job_log"
	job.JobConfig.LocalPath = filepath.Join("jobs/", name)

	job.UpdateConfig()

	// job.JobConfig.Steps = append(job.JobConfig.Steps, StepData{Name: "git", Action: scm.NewGit(origin)})
	// job.JobConfig.Steps = append(job.JobConfig.Steps, StepData{Name: "run command", Action: &plugin.RunCommand{}})
	// job.JobConfig.Steps = append(job.JobConfig.Steps, StepData{Name: "artifact", Action: &plugin.Artifact{}})

	return job
}

func (j *Job) UpdateConfig() {
	// first update the repository that this job is based on
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
