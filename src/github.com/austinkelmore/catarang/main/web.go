package main

import (
	"log"
	"net/http"
	"text/template"

	"github.com/austinkelmore/catarang/job"

	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
)

func createRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", renderWebpage)
	r.HandleFunc("/jobs", jobsHandler)
	r.HandleFunc("/jobs/add", addJobHandler)

	r.HandleFunc("/job/{name}", jobHandler)
	r.HandleFunc("/job/{name}/start", startJobHandler)
	r.HandleFunc("/job/{name}/delete", deleteJobHandler)

	r.Handle("/ws", websocket.Handler(handleWebsocketConn))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))
	return r
}

func addJob(name string, repo string) bool {
	// names must be unique
	for _, j := range config.Jobs {
		if j.Name == name {
			return false
		}
	}

	job := job.NewJob(name, repo)
	config.Jobs = append(config.Jobs, job)
	saveConfig()
	d := struct {
		Name string `json:"name"`
		Repo string `json:"repo"`
	}{
		name,
		repo,
	}
	sendWebsocketEvent("addJob", d)
	log.Println("Added job: ", name)

	return true
}

func addJobHandler(w http.ResponseWriter, r *http.Request) {
	added := addJob(r.FormValue("name"), r.FormValue("repo"))
	if !added {
		http.Error(w, "Name already exists for a job.", http.StatusConflict)
	}
}

func deleteJob(jobName string) {
	for i := range config.Jobs {
		if config.Jobs[i].Name == jobName {
			config.Jobs = append(config.Jobs[:i], config.Jobs[i+1:]...)
			saveConfig()
			d := struct {
				Name string `json:"name"`
			}{
				jobName,
			}
			sendWebsocketEvent("deleteJob", d)
			log.Println("Deleted job: ", jobName)
			break
		}
	}
}

func deleteJobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobName := vars["name"]
	// todo: akelmore - error check that we can delete this job before trying to delete it
	go deleteJob(jobName)
}

func startJob(jobName string) {
	for i := range config.Jobs {
		if config.Jobs[i].Name == jobName {
			config.Jobs[i].Run()
			saveConfig()
			break
		}
	}
}

func startJobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobName := vars["name"]
	go startJob(jobName)
}

func jobHandler(w http.ResponseWriter, r *http.Request) {
	// todo: akelmore - unify the html pages and cache them off so they're not being parsed every time the webpage is hit
	root, err := template.ParseFiles("web/job.html")
	if err != nil {
		log.Println("Can't parse root.html file.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var job *job.Job
	jobName := mux.Vars(r)["name"]
	for i, j := range config.Jobs {
		if j.Name == jobName {
			job = &config.Jobs[i]
		}
	}

	if job != nil {
		root.Execute(w, *job)
	} else {
		http.Error(w, "Unknown job: "+jobName, http.StatusInternalServerError)
	}
}

func jobsHandler(w http.ResponseWriter, r *http.Request) {
	// todo: akelmore - unify the html pages and cache them off so they're not being parsed every time the webpage is hit
	root, err := template.ParseFiles("web/jobs.html")
	if err != nil {
		log.Println("Can't parse root.html file.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	root.Execute(w, config)
}

func sendWebsocketEvent(eventType string, data interface{}) {
	d := struct {
		EventType string      `json:"type"`
		Data      interface{} `json:"data"`
	}{
		eventType,
		data,
	}

	for _, conn := range config.conns {
		if err := websocket.JSON.Send(conn, d); err != nil {
			log.Printf("Error sending websocket: %s\n", err.Error())
		}
	}
}

func handleWebsocketConn(ws *websocket.Conn) {
	config.conns = append(config.conns, ws)

	// never exit from this function otherwise the websocket connection closes
	select {}
}

func renderWebpage(w http.ResponseWriter, r *http.Request) {
	// todo: akelmore - unify the html pages and cache them off so they're not being parsed every time the webpage is hit
	root, err := template.ParseFiles("web/root.html")
	if err != nil {
		log.Println("Can't parse root.html file.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	root.Execute(w, config)
}
