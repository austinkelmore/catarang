package job

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/austinkelmore/catarang/scm"
)

type Status int

const (
	RUNNING Status = iota
	FAILED
	SUCCESSFUL
)

type Config struct {
	LocalPath       string
	BuildConfigPath string
	// todo: akelmore - make the build command more robust than a string
	BuildCommand string
	// todo: akelmore don't hard code the git scm in the job config
	Git scm.Git
}

type Job struct {
	Name      string
	Enabled   bool
	CurConfig Config
	History   []Instance
}

func CreateJob(name string, onlineRepo string, configPath string) Job {
	job := Job{Name: name, Enabled: true}

	job.CurConfig.BuildConfigPath = configPath
	// todo: akelmore - configure local path
	job.CurConfig.LocalPath = "jobs/" + name + "/"
	job.CurConfig.Git = scm.CreateGit(job.CurConfig.LocalPath, onlineRepo)
	return job
}

func (j *Job) getLastInst() *Instance {
	if len(j.History) == 0 {
		return nil
	}

	return &j.History[len(j.History)-1]
}

func (j *Job) isRunning() bool {
	inst := j.getLastInst()
	return inst != nil && inst.Status == RUNNING
}

func (j *Job) NeedsRunning() bool {
	return len(j.History) == 0 || j.needsUpdate()
}

func (j *Job) Run() {
	log.Println("Running job:", j.Name)
	inst := NewInstance(j.CurConfig)
	j.History = append(j.History, inst)

	// todo: akelmore - track whether the first time setup ever worked instead of seeing if this is the first instance
	if len(j.History) == 0 {
		log.Println("Running first time setup for:", j.Name)
		if err := inst.Config.Git.FirstTimeSetup(); err != nil {
			log.Println(err.Error())
			inst.Status = FAILED
		}
	} else {
		j.update()
	}

	if j.getLastInst().Status != FAILED {
		j.runCommand()
	}

	// todo: akelmore - figure out what to do with this
	// saveConfig()
}

func (j *Job) needsUpdate() bool {
	inst := j.getLastInst()
	if inst != nil && time.Since(inst.StartTime) < 30*time.Second {
		return false
	}
	log.Println("Running needsUpdate for:", j.Name)

	// todo: akelmore - pull this multiwriter into Job so it can be output on the web
	var b bytes.Buffer
	multi := io.MultiWriter(&b, os.Stdout)

	cmd := exec.Command("git", "-C", inst.Config.Git.LocalRepo, "ls-remote", "origin", "-h", "HEAD")
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		return false
	}

	remoteHead := string(bytes.Fields(b.Bytes())[0])

	b.Reset()
	cmd = exec.Command("git", "-C", inst.Config.Git.LocalRepo, "rev-parse", "HEAD")
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		return false
	}

	localHead := string(bytes.Fields(b.Bytes())[0])

	return remoteHead != localHead
}

func (j *Job) update() {
	log.Println("Running update for:", j.Name)

	inst := j.getLastInst()

	if err := inst.UpdateSCMandBuildCommand(); err != nil {

	}
}

func (j *Job) runCommand() {
	log.Println("Running command for:", j.Name)

	inst := j.getLastInst()

	fields := strings.Fields(inst.Config.BuildCommand)
	if len(fields) > 0 {
		var b bytes.Buffer
		multi := io.MultiWriter(&b, os.Stdout)
		cmd := exec.Command(fields[0], fields[1:]...)
		cmd.Stdout = multi
		cmd.Stderr = multi
		cmd.Dir = inst.Config.Git.LocalRepo
		if err := cmd.Run(); err != nil {
			log.Println("ERROR RUNNING BUILD:", j.Name)
			inst.Status = FAILED
			return
		}
	}
}
