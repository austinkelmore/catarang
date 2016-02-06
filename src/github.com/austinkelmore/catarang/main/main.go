package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/austinkelmore/catarang/job"

	"golang.org/x/net/websocket"
)

// Config All of the run time data for the Catarang server
type Config struct {
	Jobs  []job.Job
	conns []*websocket.Conn
}

var config Config
var configFileName = "catarang_config.json"

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

func updateConsoleText() {
	// type inOut struct {
	// 	err int
	// 	out int
	// }
	// var sent []inOut

	// // todo: akelmore - fix up websockets working
	// for {
	// 	if len(config.Jobs) > 0 && len(config.Jobs[0].History) > 0 {
	// 		for index := range config.Jobs[0].History[0].Log {
	// 			if index >= len(sent) {
	// 				if index > len(sent)-1 {
	// 					sent = append(sent, inOut{err: 0, out: 0})
	// 				}

	// 				logger := &config.Jobs[0].History[0].Log[index]
	// 				splitErr := strings.Split(string(logger.Cmds.Bytes()), "\n")
	// 				for i := sent[index].err; i < len(splitErr); i++ {
	// 					sendWebsocketEvent("consoleLog", splitErr[i])
	// 				}
	// 				splitOut := strings.Split(string(logger.Out.Bytes()), "\n")
	// 				for i := sent[index].out; i < len(splitOut); i++ {
	// 					sendWebsocketEvent("consoleLog", splitOut[i])
	// 				}

	// 				sent[index].err = len(splitErr)
	// 				sent[index].out = len(splitOut)
	// 			}
	// 		}
	// 	}
	// 	time.Sleep(time.Second * 2)
	// }
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Running Catarang!")
	readInConfig()

	// todo: akelmore - instead of polling the console logs, have callbacks to send the data to the client
	go updateConsoleText()

	r := createRoutes()
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
