package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/austinkelmore/catarang/job"

	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
)

// Config All of the run time data for the Catarang server
type Config struct {
	Jobs  []job.Job
	conns []*websocket.Conn
}

var config Config
var configFileName = "catarang_config.json"

func addJob(name string, repo string) {
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
}

func addJobHandler(w http.ResponseWriter, r *http.Request) {
	go addJob(r.FormValue("name"), r.FormValue("repo"))
	renderWebpage(w, r)
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

func sendWebsocketEvent(eventType string, data interface{}) {
	d := struct {
		EventType string      `json:"type"`
		Data      interface{} `json:"data"`
	}{
		eventType,
		data,
	}
	// todo: akelmore - handle marshal error case
	m, _ := json.Marshal(d)
	log.Printf("websocket message: %s\n", string(m))

	for _, conn := range config.conns {
		if err := websocket.Message.Send(conn, string(m)); err != nil {
			log.Printf("Error sending websocket: %s\n", err.Error())
		}
	}
}

func handleConsoleText(ws *websocket.Conn) {
	config.conns = append(config.conns, ws)
	type inOut struct {
		err int
		out int
	}
	var sent []inOut

	for {
		if len(config.Jobs) > 0 && len(config.Jobs[0].History) > 0 {
			for index := range config.Jobs[0].History[0].Log {
				if index >= len(sent) {
					if index > len(sent)-1 {
						sent = append(sent, inOut{err: 0, out: 0})
					}

					logger := &config.Jobs[0].History[0].Log[index]
					splitErr := strings.Split(string(logger.Err.Bytes()), "\n")
					for i := sent[index].err; i < len(splitErr); i++ {
						d := struct {
							EventType string      `json:"type"`
							Data      interface{} `json:"data"`
						}{
							"consoleLog",
							splitErr[i],
						}
						websocket.JSON.Send(ws, d)
						// if err := websocket.Message.Send(ws, splitErr[i]); err != nil {
						// 	log.Printf("Error sending websocket: %s\n", err.Error())
						// }
					}
					splitOut := strings.Split(string(logger.Out.Bytes()), "\n")
					for i := sent[index].out; i < len(splitOut); i++ {
						d := struct {
							EventType string      `json:"type"`
							Data      interface{} `json:"data"`
						}{
							"consoleLog",
							splitOut[i],
						}
						websocket.JSON.Send(ws, d)
						// if err := websocket.Message.Send(ws, splitOut[i]); err != nil {
						// 	log.Printf("Error sending websocket: %s\n", err.Error())
						// }
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

	r := mux.NewRouter()

	r.HandleFunc("/", renderWebpage)
	r.HandleFunc("/jobs/add", addJobHandler)

	r.HandleFunc("/job/{name}/start", startJobHandler)
	r.HandleFunc("/job/{name}/delete", deleteJobHandler)

	r.Handle("/ws", websocket.Handler(handleConsoleText))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":8080", nil))
}