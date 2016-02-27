package job

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/austinkelmore/catarang/ulog"
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
	StartTime    time.Time
	EndTime      time.Time
	Num          int
	JobConfig    Config
	BuildCommand BuildCommand
	Status       Status
	Artifacts    Artifact
	Log          []ulog.Job
}

// NewInstance Creates a new instance of a job (and copies off the current config)
// and starts the instance running
func NewInstance(config Config, instanceNum int) Instance {
	inst := Instance{StartTime: time.Now(), Status: RUNNING, JobConfig: config, Num: instanceNum}
	return inst
}

func (i *Instance) appendLog(name string) *ulog.Job {
	i.Log = append(i.Log, *ulog.NewJob(name))
	return &i.Log[len(i.Log)-1]
}

func (i *Instance) fail(reason string) {
	i.Status = FAILED
}

// todo: akelmore - do i still need to update the build command before starting every time? should i only do it sometimes or not at all?
func (i *Instance) updateBuildCommand() error {
	// read in the config file's build command
	path := i.JobConfig.BuildConfigPath
	if path == "" {
		path = ".catarang.json"
	}

	file, err := ioutil.ReadFile(i.JobConfig.SourceControl.LocalRepoPath() + path)
	if err != nil {
		return errors.New("Error reading build config file: " + path)
	}
	err = json.Unmarshal(file, &i.BuildCommand)
	if err != nil {
		return errors.New("Error reading JSON from build config file: " + path)
	}

	return nil
}

// Start Entry point for the instance
// todo: akelmore - i don't like passing a bool for the completedSetup, figure something better out
func (i *Instance) Start(completedSetup *bool) {
	// todo: akelmore - make jobs have an array of "things" to do rather than hard code scm stuff first
	if !*completedSetup {
		logger := i.appendLog("git - initial setup")
		if err := i.JobConfig.SourceControl.FirstTimeSetup(&logger.Cmds); err != nil {
			i.fail(err.Error())
			return
		}
		*completedSetup = true
	} else {
		logger := i.appendLog("git - sync")
		if err := i.JobConfig.SourceControl.UpdateExisting(&logger.Cmds); err != nil {
			i.fail(err.Error())
			return
		}
	}

	logger := i.appendLog("cmd")
	if err := i.updateBuildCommand(); err != nil {
		i.fail("Error updating build command from config file.")
		return
	}

	fields := strings.Fields(i.BuildCommand.ExecCommand)
	if len(fields) > 0 {
		// todo: akelmore - make sure that fields has more than one field before trying to access it
		cmd := ulog.NewCmd(&logger.Cmds, fields[0], fields[1:]...)
		cmd.Cmd.Dir = i.JobConfig.SourceControl.LocalRepoPath()
		if err := cmd.Run(); err != nil {
			i.fail("Error running exec command.")
			return
		}
	}

	// todo: akelmore - make artifacts part of the array of things
	log.Printf("artifacts: %+v\n", i.BuildCommand.Artifacts)
	for _, artifact := range i.BuildCommand.Artifacts {
		if err := artifact.Save(i.JobConfig.SourceControl.LocalRepoPath(), i.JobConfig.Name, i.Num); err != nil {
			log.Printf("Error saving artifact. %s\n", artifact.ToSave)
			i.fail("Error saving artifact.")
			return
		}
	}

	log.Println("Success running exec command!")
	i.Status = SUCCESSFUL
}
