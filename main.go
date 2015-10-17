package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/austinkelmore/catarang/job"
)

// Config All of the run time data for the Catarang server
type Config struct {
	Jobs []job.Job
}

var config Config
var configFileName = "catarang_config.json"

func addJob(w http.ResponseWriter, r *http.Request) {
	job := job.NewJob(r.FormValue("name"), r.FormValue("repo"), r.FormValue("build_config"))
	config.Jobs = append(config.Jobs, job)
	saveConfig()

	renderWebpage(w, r)
}

func deleteJob(w http.ResponseWriter, r *http.Request) {
	renderWebpage(w, r)
}

func pollJobs() {
	for {
		// todo: akelmore - figure out if this is safe to poll like this if
		// we're inserting/deleting from it or if there needs to be a lock of some sort
		for index := range config.Jobs {
			if config.Jobs[index].NeedsRunning() {
				config.Jobs[index].Run()
				saveConfig()
			}
		}
		time.Sleep(time.Second * 10)
	}
}

func renderWebpage(w http.ResponseWriter, r *http.Request) {
	root, err := template.ParseFiles("web/root.html")
	if err != nil {
		log.Println("Can't parse root.html file.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	root.Execute(w, config)
}

// todo: akelmore - fix threading with the reading/writing of the config
func readInConfig() {
	data, err := ioutil.ReadFile(configFileName)
	if err == nil {
		if err = json.Unmarshal(data, &config); err != nil {
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
	data, err := json.MarshalIndent(&config, "", "\t")
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

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Running Catarang!")
	readInConfig()

	go pollJobs()

	http.HandleFunc("/", renderWebpage)
	http.HandleFunc("/addjob", addJob)
	http.HandleFunc("/deletejob", deleteJob)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
