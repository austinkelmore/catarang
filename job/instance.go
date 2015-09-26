package job

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Status The job instance's status
type Status int

const (
	// RUNNING The job instance is currently running
	RUNNING Status = iota
	// FAILED The job instance failed
	FAILED
	// SUCCESSFUL The job instance was successful
	SUCCESSFUL
)

// Instance a single run of a job
type Instance struct {
	StartTime time.Time
	EndTime   time.Time
	Config    Config
	// todo: akelmore - make the build command more robust than a string
	BuildCommand BuildCommand
	Status       Status
	Out          io.Writer
	Err          io.Writer
}

// NewInstance Creates a new instance of a job (and copies off the current config)
// and starts the instance running
func NewInstance(config Config) Instance {
	inst := Instance{StartTime: time.Now(), Status: RUNNING, Config: config}
	var bOut bytes.Buffer
	inst.Out = io.MultiWriter(&bOut, os.Stdout)
	var bErr bytes.Buffer
	inst.Err = io.MultiWriter(&bErr, os.Stderr)
	return inst
}

func (i *Instance) updateBuildCommand() error {
	// read in the config file's build command
	file, err := ioutil.ReadFile(i.Config.SourceControl.LocalRepoPath() + i.Config.BuildConfigPath)
	if err != nil {
		log.Println("Error reading build config file: " + i.Config.BuildConfigPath)
		return err
	}

	err = json.Unmarshal(file, &i.BuildCommand)
	if err != nil {
		log.Println("Error reading JSON from build config file: " + i.Config.BuildConfigPath)
		return err
	}

	return nil
}

// Start Entry point for the instance
func (i *Instance) Start() error {
	if err := i.updateBuildCommand(); err != nil {
		log.Println("Error updating build command from config file.")
		return err
	}

	fields := strings.Fields(i.BuildCommand.ExecCommand)
	if len(fields) > 0 {
		cmd := exec.Command(fields[0], fields[1:]...)
		cmd.Stdout = i.Out
		cmd.Stderr = i.Err
		cmd.Dir = i.Config.SourceControl.LocalRepoPath()
		if err := cmd.Run(); err != nil {
			log.Println("Error running exec command.")
			return err
		}
	}

	log.Println("Success running exec command!")

	return nil
}
