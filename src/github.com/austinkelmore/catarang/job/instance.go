package job

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/austinkelmore/catarang/splitlog"
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
	Config       Config
	BuildCommand BuildCommand
	Status       Status
	Log          []splitlog.JobLog
}

// NewInstance Creates a new instance of a job (and copies off the current config)
// and starts the instance running
func NewInstance(config Config) Instance {
	inst := Instance{StartTime: time.Now(), Status: RUNNING, Config: config}
	return inst
}

func (i *Instance) appendLog(name string) *splitlog.JobLog {
	i.Log = append(i.Log, splitlog.JobLog{Name: name})
	return &i.Log[len(i.Log)-1]
}

func (i *Instance) fail(reason string) {
	i.Status = FAILED
}

// todo: akelmore - do i still need to update the build command before starting every time? should i only do it sometimes or not at all?
func (i *Instance) updateBuildCommand() error {
	// read in the config file's build command
	path := i.Config.BuildConfigPath
	if path == "" {
		path = ".catarang.json"
	}

	file, err := ioutil.ReadFile(i.Config.SourceControl.LocalRepoPath() + path)
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
	// todo: akelmore - pull out the notion of a first time setup and let modules have their own internal states on a per-job basis
	if !*completedSetup {
		logger := i.appendLog("git - initial setup")
		if err := i.Config.SourceControl.FirstTimeSetup(logger.Cmds); err != nil {
			i.fail(err.Error())
			return
		}
		*completedSetup = true
	} else {
		logger := i.appendLog("git - sync")
		if err := i.Config.SourceControl.UpdateExisting(logger.Cmds); err != nil {
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
		cmd := splitlog.New(logger.Cmds, fields[0], fields[1:]...)
		cmd.Cmd.Dir = i.Config.SourceControl.LocalRepoPath()
		if err := cmd.Run(); err != nil {
			i.fail("Error running exec command.")
			return
		}
	}

	log.Println("Success running exec command!")
	i.Status = SUCCESSFUL
}
