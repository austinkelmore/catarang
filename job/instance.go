package job

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

type Instance struct {
	StartTime    time.Time
	EndTime      time.Time
	Config       Config
	BuildCommand BuildCommand
	Status       Status
}

func NewInstance(config Config) Instance {
	return Instance{StartTime: time.Now(), Status: RUNNING, Config: config}
}

func (i *Instance) UpdateSCMandBuildCommand() error {

	var b bytes.Buffer
	multi := io.MultiWriter(&b, os.Stdout)

	// update the git repo
	// todo: akelmore - pull into the git scm module
	cmd := exec.Command("git", "-C", i.Config.Git.LocalRepo, "pull")
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		log.Println("Error pulling git")
		return err
	} else if bytes.Contains(b.Bytes(), []byte("Already up-to-date.")) {
		return errors.New("Something went wrong with the git pull, it was already up to date. It shouldn't have been.")
	}

	// read in the config file's build command
	file, err := ioutil.ReadFile(i.Config.Git.LocalRepo + i.Config.BuildConfigPath)
	if err != nil {
		return errors.New("Error reading build config file: " + i.Config.BuildConfigPath)
	}

	err = json.Unmarshal(file, &i.BuildCommand)
	if err != nil {
		return errors.New("Error reading JSON from build config file: " + i.Config.BuildConfigPath)
	}

	return nil
}

func (i *Instance) GetElapsedTime() time.Duration {
	if i.Status == RUNNING {
		return time.Since(i.StartTime)
	}

	return i.EndTime.Sub(i.StartTime)
}
