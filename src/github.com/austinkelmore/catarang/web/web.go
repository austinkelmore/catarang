package web

import (
	"log"
	"net/http"
	"text/template"

	"github.com/austinkelmore/catarang/catarang"

	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
)

func CreateRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", renderWebpage)
	r.HandleFunc("/jobs", jobsHandler)
	r.HandleFunc("/jobs/add", addJobHandler)
	r.Handle("/jobs/ws", websocket.Handler(handleJobsWSConn))

	r.HandleFunc("/job/{name}", jobHandler)
	r.HandleFunc("/job/{name}/start", startJobHandler)
	r.HandleFunc("/job/{name}/delete", deleteJobHandler)
	r.HandleFunc("/job/{name}/ws", jobWSHandler)
	// r.Handle("/job/{name}/ws", websocket.Handler(handleJobWSConn))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))
	return r
}

func addJobHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	repo := r.FormValue("repo")
	added := catarang.AddJob(name, repo)
	if !added {
		http.Error(w, "Name already exists for a job.", http.StatusConflict)
		return
	}

	d := struct {
		Name string `json:"name"`
		Repo string `json:"repo"`
	}{name, repo}
	sendWebsocketEvent("addJob", d)
}

func deleteJobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobName := vars["name"]

	// todo: akelmore - error check that we can delete this job before trying to delete it
	ok := catarang.DeleteJob(jobName)
	if ok {
		d := struct {
			Name string `json:"name"`
		}{jobName}
		sendWebsocketEvent("deleteJob", d)
	}
}

func startJobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobName := vars["name"]
	go catarang.StartJob(jobName)
}

func jobHandler(w http.ResponseWriter, r *http.Request) {
	// todo: akelmore - unify the html pages and cache them off so they're not being parsed every time the webpage is hit
	root, err := template.ParseFiles("web/job.html")
	if err != nil {
		log.Println("Can't parse root.html file.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jobName := mux.Vars(r)["name"]
	for _, j := range catarang.GetJobs() {
		if j.GetName() == jobName {
			root.Execute(w, &j)
			return
		}
	}

	http.Error(w, "Unknown job: "+jobName, http.StatusInternalServerError)
}

func jobsHandler(w http.ResponseWriter, r *http.Request) {
	// todo: akelmore - unify the html pages and cache them off so they're not being parsed every time the webpage is hit
	root, err := template.ParseFiles("web/jobs.html")
	if err != nil {
		log.Println("Can't parse root.html file.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jobs := catarang.GetJobs()
	root.Execute(w, &jobs)
}

func sendWebsocketEvent(eventType string, data interface{}) {
	d := struct {
		EventType string      `json:"type"`
		Data      interface{} `json:"data"`
	}{eventType, data}
	catarang.SendToJobsConns(d)
}

func handleJobsWSConn(ws *websocket.Conn) {
	catarang.AddJobsConn(ws)

	// never exit from this function otherwise the websocket connection closes
	select {}
}

func jobWSHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobName := vars["name"]

	websocket.Handler(func(ws *websocket.Conn) {
		catarang.AddJobConn(jobName, ws)

		// never exit from this function otherwise the websocket connection closes
		select {}
	}).ServeHTTP(w, r)
}

func renderWebpage(w http.ResponseWriter, r *http.Request) {
	// todo: akelmore - unify the html pages and cache them off so they're not being parsed every time the webpage is hit
	root, err := template.ParseFiles("web/root.html")
	if err != nil {
		log.Println("Can't parse root.html file.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	root.Execute(w, nil)
}