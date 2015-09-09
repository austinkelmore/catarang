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

type Instance struct {
	StartTime time.Time
	EndTime   time.Time
	Config    Config
	// todo: akelmore - make the build command more robust than a string
	BuildCommand BuildCommand
	Status       Status
}

func NewInstance(config Config) Instance {
	return Instance{StartTime: time.Now(), Status: RUNNING, Config: config}
}

func (i *Instance) updateBuildCommand() error {
	// read in the config file's build command
	file, err := ioutil.ReadFile(i.Config.Git.LocalRepo + i.Config.BuildConfigPath)
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

func (i *Instance) RunExecCommand() error {
	if err := i.updateBuildCommand(); err != nil {
		log.Println("Error updating build command from config file.")
		return err
	}

	fields := strings.Fields(i.BuildCommand.ExecCommand)
	if len(fields) > 0 {
		var b bytes.Buffer
		multi := io.MultiWriter(&b, os.Stdout)
		cmd := exec.Command(fields[0], fields[1:]...)
		cmd.Stdout = multi
		cmd.Stderr = multi
		cmd.Dir = i.Config.Git.LocalRepo
		if err := cmd.Run(); err != nil {
			log.Println("Error running exec command.")
			return err
		}
	}

	log.Println("Success running exec command!")

	return nil
}
