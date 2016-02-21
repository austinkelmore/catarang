package catarang

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/austinkelmore/catarang/job"
	"github.com/austinkelmore/catarang/util"
	"golang.org/x/net/websocket"
)

var cats Catarang

func init() {
	cats.conns = make(map[string][]*websocket.Conn)
}

type Catarang struct {
	Jobs      []job.Job
	conns     map[string][]*websocket.Conn
	jobsConns []*websocket.Conn
}

func AddJob(name string, repo string) bool {
	// names must be unique
	for _, j := range cats.Jobs {
		if j.Name == name {
			return false
		}
	}

	job := job.NewJob(name, repo)
	cats.Jobs = append(cats.Jobs, job)
	saveConfig()
	log.Println("Added job: ", name)

	return true
}

func DeleteJob(jobName string) bool {
	for i := range cats.Jobs {
		if cats.Jobs[i].Name == jobName {
			util.ForceRemoveAll(cats.Jobs[i].CurConfig.LocalPath)
			cats.Jobs = append(cats.Jobs[:i], cats.Jobs[i+1:]...)
			saveConfig()
			log.Println("Deleted job: ", jobName)
			return true
		}
	}
	return false
}

func StartJob(jobName string) {
	for i := range cats.Jobs {
		if cats.Jobs[i].Name == jobName {
			cats.Jobs[i].Run()
			saveConfig()
			break
		}
	}
}

func GetJobs() []job.Job {
	return cats.Jobs
}

func AddJobsConn(ws *websocket.Conn) {
	cats.jobsConns = append(cats.jobsConns, ws)
}

func AddJobConn(jobName string, ws *websocket.Conn) {
	cats.conns[jobName] = append(cats.conns[jobName], ws)
}

func SendToJobsConns(data interface{}) {
	for _, conn := range cats.jobsConns {
		if err := websocket.JSON.Send(conn, data); err != nil {
			log.Printf("Error sending websocket: %s\n", err.Error())
		}
	}
}

func SendToJobConns(jobName string, data interface{}) {
	for _, conn := range cats.conns[jobName] {
		if err := websocket.JSON.Send(conn, data); err != nil {
			log.Printf("Error sending websocket: %s\n", err.Error())
		}
	}
}

var configFileName = "catarang_config.json"

// todo: akelmore - fix threading with the reading/writing of the config
func ReadInConfig() {
	data, err := ioutil.ReadFile(configFileName)
	if err == nil {
		if err = json.Unmarshal(data, &cats); err != nil {
			log.Println("Error reading in", configFileName)
			log.Println(err.Error())
		}
		return
	}

	// create a new config and save it out
	log.Println("No Catarang config detected, creating new one.")
	saveConfig()
}

func saveConfig() {
	data, err := json.MarshalIndent(&cats, "", "\t")
	if err != nil {
		log.Println("Error marshaling save data:", err.Error())
		return
	}

	err = ioutil.WriteFile(configFileName, []byte(data), 0644)
	if err != nil {
		log.Println("Error writing config file", configFileName)
		log.Println(err.Error())
	}
}
