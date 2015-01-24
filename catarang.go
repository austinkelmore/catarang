package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"time"
)

type Agent struct {
	name string
}

type Job struct {
	Name    string
	Enabled bool
	running bool
	Git_url string
}

type CatarangConfig struct {
	Jobs []Job
}

var config CatarangConfig
var config_file_name = "catarang_config.json"

func addJob(w http.ResponseWriter, r *http.Request) {
	job := Job{Enabled: true}
	job.Name = r.FormValue("name")
	job.Git_url = r.FormValue("git_url")
	config.Jobs = append(config.Jobs, job)
	saveConfig()

	renderWebpage(w, r)
}

func deleteJob(w http.ResponseWriter, r *http.Request) {
	renderWebpage(w, r)
}

func pollJobs() {
	sleepAmount := time.Second * 30

	for {
		log.Println("Checking all jobs.")
		for _, job := range config.Jobs {
			log.Printf("Checking for job: %v\n", job.Name)
		}
		time.Sleep(sleepAmount)
	}
}

func renderWebpage(w http.ResponseWriter, r *http.Request) {
	root, err := template.ParseFiles("root.html")
	if err != nil {
		fmt.Println("Can't parse root.html file.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	root.Execute(w, config)
}

// todo: akelmore - fix threading with the reading/writing of the config
func readInConfig() {
	data, err := ioutil.ReadFile(config_file_name)
	if err != nil {
		return
	}

	if err = json.Unmarshal(data, &config); err != nil {
		fmt.Println("Error reading in", config_file_name)
		fmt.Println(err.Error())
	}
}

func saveConfig() {
	data, err := json.Marshal(&config)
	if err != nil {
		fmt.Println("Error marshaling save data:", err.Error())
		return
	}

	err = ioutil.WriteFile(config_file_name, []byte(data), 0644)
	if err != nil {
		fmt.Println("Error writing config file", config_file_name)
		fmt.Println(err.Error())
	}
}

func main() {
	readInConfig()

	go pollJobs()

	http.HandleFunc("/", renderWebpage)
	http.HandleFunc("/addjob", addJob)
	http.HandleFunc("/deletejob", deleteJob)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
