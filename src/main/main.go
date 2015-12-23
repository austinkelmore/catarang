package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"job"

	"github.com/golang/net/websocket"
)

// Config All of the run time data for the Catarang server
type Config struct {
	Jobs  []job.Job
	Conns []*websocket.Conn
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

func startJob(w http.ResponseWriter, r *http.Request) {

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

func handleConsoleText(ws *websocket.Conn) {
	config.Conns = append(config.Conns, ws)
	type inOut struct {
		err int
		out int
	}
	var sent []inOut

	for {
		if len(config.Jobs) > 0 && len(config.Jobs[0].History) > 0 {
			for index := range config.Jobs[0].History[0].Log {
				if index >= len(sent) {
					log.Printf("Index = %v\n", index)
					if index > len(sent)-1 {
						sent = append(sent, inOut{err: 0, out: 0})
					}

					logger := &config.Jobs[0].History[0].Log[index]
					splitErr := strings.Split(string(logger.Err.Bytes()), "\n")
					for i := sent[index].err; i < len(splitErr); i++ {
						if err := websocket.Message.Send(ws, splitErr[i]); err != nil {
							log.Printf("Error sending websocket: %s\n", err.Error())
						}

					}
					splitOut := strings.Split(string(logger.Out.Bytes()), "\n")
					for i := sent[index].out; i < len(splitOut); i++ {
						if err := websocket.Message.Send(ws, splitOut[i]); err != nil {
							log.Printf("Error sending websocket: %s\n", err.Error())
						}
					}

					sent[index].err = len(splitErr)
					sent[index].out = len(splitOut)
				}
			}
		}
		time.Sleep(time.Second * 2)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Running Catarang!")
	readInConfig()

	go pollJobs()

	http.HandleFunc("/", renderWebpage)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))
	http.HandleFunc("/addjob", addJob)
	http.HandleFunc("/deletejob", deleteJob)
	http.HandleFunc("/startjob", startJob)

	http.Handle("/ws", websocket.Handler(handleConsoleText))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
