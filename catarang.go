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

type Config struct {
	Jobs []Job
	Git  GitPluginOptions
}

var config Config
var config_file_name = "catarang_config.json"

func addJob(w http.ResponseWriter, r *http.Request) {
	jobConfig := job.Config{Repo: r.FormValue("repo"), BuildConfig: r.FormValue("build_config")}
	job := CreateJob(r.FormValue("name"), jobConfig)
	// job.Config.Repo = r.FormValue("repo")
	// job.Config.BuildConfig = r.FormValue("build_config")
	config.Jobs = append(config.Jobs, job)
	saveConfig()

	renderWebpage(w, r)
}

func deleteJob(w http.ResponseWriter, r *http.Request) {
	renderWebpage(w, r)
}

func pollJobs() {
	for {
		for index := range config.Jobs {
			if config.Jobs[index].needsRunning() {
				config.Jobs[index].run()
			}
		}
		time.Sleep(time.Second * 10)
	}
}

func renderWebpage(w http.ResponseWriter, r *http.Request) {
	root, err := template.ParseFiles("root.html")
	if err != nil {
		log.Println("Can't parse root.html file.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	root.Execute(w, config)
}

// todo: akelmore - fix threading with the reading/writing of the config
func readInConfig() {
	data, err := ioutil.ReadFile(config_file_name)
	if err == nil {
		if err = json.Unmarshal(data, &config); err != nil {
			log.Println("Error reading in", config_file_name)
			log.Println(err.Error())
		}
		return
	}

	// create a new config and save it out
	log.Println("No catarang config detected, creating new one.")
	saveConfig()
}

func saveConfig() {
	data, err := json.MarshalIndent(&config, "", "\t")
	if err != nil {
		log.Println("Error marshaling save data:", err.Error())
		return
	}

	err = ioutil.WriteFile(config_file_name, []byte(data), 0644)
	if err != nil {
		log.Println("Error writing config file", config_file_name)
		log.Println(err.Error())
	}
}

func main() {
	log.Println("Running Catarang!")
	readInConfig()

	go pollJobs()

	http.HandleFunc("/", renderWebpage)
	http.HandleFunc("/addjob", addJob)
	http.HandleFunc("/deletejob", deleteJob)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
