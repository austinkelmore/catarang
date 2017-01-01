package catarang

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/austinkelmore/catarang/job"
	"github.com/austinkelmore/catarang/jobdata"
	"github.com/austinkelmore/catarang/plugin/scm"
	"github.com/pkg/errors"
	"golang.org/x/net/websocket"
)

var cats Catarang

func init() {
	cats.conns = make(map[string][]*websocket.Conn)
}

// Catarang is the root config for the whole app
type Catarang struct {
	Jobs     []job.Job
	conns    map[string][]*websocket.Conn
	jobConns []*websocket.Conn
}

// AddJob will add a new job to the server's config
func AddJob(name string, repo string) error {
	// names must be unique
	for _, j := range cats.Jobs {
		if j.GetName() == name {
			return errors.Errorf("job \"%s\" already exists, job names must be unique", name)
		}
	}

	git := scm.Git{Origin: repo}
	gitjson, err := json.Marshal(git)
	if err != nil {
		return errors.Wrapf(err, "couldn't marshal scm.git to json")
	}
	stepTemplate := jobdata.StepTemplate{PluginName: git.GetName(), PluginData: gitjson}
	template := jobdata.JobTemplate{LocalPath: filepath.Join("jobs/", name)}
	template.Steps = append(template.Steps, stepTemplate)
	job, err := job.New(template)
	if err != nil {
		return errors.Wrapf(err, "couldn't create job %s", name)
	}
	job.JobData.Name = name
	job.JobData.ID = len(cats.Jobs)

	cats.Jobs = append(cats.Jobs, job)
	saveConfig()
	log.Println("Added job: ", name)
	return nil
}

// DeleteJob will delete a job from the server's config
func DeleteJob(jobName string) bool {
	for i := range cats.Jobs {
		if cats.Jobs[i].GetName() == jobName {
			cats.Jobs[i].Clean()
			cats.Jobs = append(cats.Jobs[:i], cats.Jobs[i+1:]...)
			saveConfig()
			log.Println("Deleted job: ", jobName)
			return true
		}
	}
	return false
}

// CleanJob will clean all of the local data from the specified job
func CleanJob(job string) {
	for i := range cats.Jobs {
		if cats.Jobs[i].GetName() == job {
			cats.Jobs[i].Clean()
		}
	}
}

// StartJob will start an instance of the job
func StartJob(jobName string) {
	for i := range cats.Jobs {
		if cats.Jobs[i].GetName() == jobName {
			cats.Jobs[i].Run()
			saveConfig()
			break
		}
	}
}

// GetJobs returns a splice of the jobs
func GetJobs() []job.Job {
	return cats.Jobs
}

// AddJobsConn appends the websocket to the list of jobs page websockets
func AddJobsConn(ws *websocket.Conn) {
	cats.jobConns = append(cats.jobConns, ws)
}

// AddJobConn appends the websocket to the list of specified job page websockets
func AddJobConn(jobName string, ws *websocket.Conn) {
	cats.conns[jobName] = append(cats.conns[jobName], ws)
}

// SendToJobsConns sends data to all of the connected jobs page websockets
func SendToJobsConns(data interface{}) {
	for _, conn := range cats.jobConns {
		if err := websocket.JSON.Send(conn, data); err != nil {
			log.Printf("Error sending websocket: %s\n", err.Error())
		}
	}
}

// SendToJobConns sends data to all of the connected specified job page websockets
func SendToJobConns(jobName string, data interface{}) {
	for _, conn := range cats.conns[jobName] {
		if err := websocket.JSON.Send(conn, data); err != nil {
			log.Printf("Error sending websocket: %s\n", err.Error())
		}
	}
}

var configFileName = "catarang_config.json"

// ReadInConfig will read in the config file and put it into the Catarang config for the server
// todo: akelmore - fix threading with the reading/writing of the config
func ReadInConfig() {
	data, err := ioutil.ReadFile(configFileName)
	if err != nil && os.IsNotExist(err) {
		// create a new config and save it out
		log.Println("No Catarang config detected, creating new one.")
		saveConfig()
		return
	}

	if err = json.Unmarshal(data, &cats); err != nil {
		log.Printf("Error reading in %v: %v\n", configFileName, err.Error())
	}
}

func saveConfig() {
	data, err := json.MarshalIndent(&cats, "", "\t")
	if err != nil {
		log.Println("Error marshaling save data:", err.Error())
		return
	}

	err = ioutil.WriteFile(configFileName, []byte(data), 0644)
	if err != nil {
		log.Printf("Error writing config file %v: %v\n", configFileName, err.Error())
	}
}

// Clean resets the entire app to a fresh install state
func Clean() {
	os.Remove(configFileName)
}
