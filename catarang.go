package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"text/template"
	"time"
)

type GitPluginOptions struct {
	Username string
	Email    string
	Path     string
}

func (g *GitPluginOptions) path(name string) string {
	return g.Path + name + "/"
}

type JobStatus int

const (
	NEVER_RUN JobStatus = iota
	POLLING
	RUNNING
	FAILED
	RECOVERED
	SUCCESSFUL
)

type Job struct {
	Name       string
	Enabled    bool
	running    bool
	Git_url    string
	LastRun    time.Time
	LastStatus JobStatus
	CurStatus  JobStatus
}

// make git have a username and email for catarang

func (j *Job) needsRunning() bool {
	return j.CurStatus == NEVER_RUN || j.needsUpdate()
}

func (j *Job) run() {
	log.Println("Running job:", j.Name)
	// todo: akelmore - make status a stack, not just two
	j.LastStatus = j.CurStatus
	j.CurStatus = RUNNING
	j.LastRun = time.Now()

	if j.LastStatus == NEVER_RUN {
		j.firstTimeSetup()
	} else {
		j.update()
	}

	if j.CurStatus != FAILED {
		switch j.LastStatus {
		case FAILED:
			j.CurStatus = RECOVERED
		default:
			j.CurStatus = SUCCESSFUL
		}
	}

	saveConfig()
}

func (j *Job) firstTimeSetup() {
	log.Println("Running first time setup for:", j.Name)

	var b bytes.Buffer
	multi := io.MultiWriter(&b, os.Stdout)

	cmd := exec.Command("git", "clone", "--depth", "1", j.Git_url, config.Git.path(j.Name))
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		log.Println("Error doing first time setup for:", j.Name)
		j.CurStatus = FAILED
		return
	}

	b.Reset()
	cmd = exec.Command("git", "-C", config.Git.path(j.Name), "config", "user.email", config.Git.Email)
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		log.Println("Error trying to set git email for:", j.Name)
		j.CurStatus = FAILED
		// todo: akelmore - clean up
		return
	}

	b.Reset()
	cmd = exec.Command("git", "-C", config.Git.path(j.Name), "config", "user.name", config.Git.Username)
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		log.Println("Error trying to set git username for:", j.Name)
		j.CurStatus = FAILED
		// todo: akelmore - clean up
	}
}

func (j *Job) needsUpdate() bool {
	if time.Since(j.LastRun) < 30*time.Second {
		return false
	}
	log.Println("Running needsUpdate for:", j.Name)

	// todo: akelmore - pull this multiwriter into Job so it can be output on the web
	var b bytes.Buffer
	multi := io.MultiWriter(&b, os.Stdout)

	cmd := exec.Command("git", "-C", config.Git.path(j.Name), "ls-remote", "origin", "-h", "HEAD")
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		return false
	}

	remoteHead := string(bytes.Fields(b.Bytes())[0])

	b.Reset()
	cmd = exec.Command("git", "-C", config.Git.path(j.Name), "rev-parse", "HEAD")
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		return false
	}

	localHead := string(bytes.Fields(b.Bytes())[0])

	return remoteHead != localHead
}

func (j *Job) update() {
	log.Println("Running update for:", j.Name)

	var b bytes.Buffer
	multi := io.MultiWriter(&b, os.Stdout)

	cmd := exec.Command("git", "-C", config.Git.path(j.Name), "pull", "--depth=1")
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		log.Println("Error pulling git for:", j.Name)
		j.CurStatus = FAILED
	} else if bytes.Contains(b.Bytes(), []byte("Already up-to-date.")) {
		log.Println("Something went wrong with the git pull, it was already up to date. It shouldn't have been.")
		j.CurStatus = FAILED
	}
}

type CatarangConfig struct {
	Jobs []Job
	Git  GitPluginOptions
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
	config.Git.Email = "catarang@austinkelmore.com"
	config.Git.Username = "catarang"
	config.Git.Path = "jobs/"
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
