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

// todo: akelmore - move git stuff out of the job
type GitPluginOptions struct {
	Username string
	Email    string
	Path     string
}

func (g *GitPluginOptions) path(name string) string {
	return g.Path + name + "/"
}

type Config struct {
	Repo         string
	BuildConfig  string
	BuildCommand string
	Git          GitPluginOptions
}

type Status int

const (
	RUNNING Status = iota
	FAILED
	SUCCESSFUL
)

// todo: akelmore - pull instance out into its own thing

type Job struct {
	Name    string
	Enabled bool
	Config  Config
	History []Instance
}

func CreateJob(name string, config Config) Job {
	job := Job{Name: name, Config: config, Enabled: true}
	job.Config.Git.Email = "catarang@austinkelmore.com"
	job.Config.Git.Username = "catarang"
	job.Config.Git.Path = "jobs/"
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
	inst := Instance{StartTime: time.Now(), Config: j.Config, Status: RUNNING}
	j.History = append(j.History, inst)

	if len(j.History) == 0 {
		j.firstTimeSetup()
	} else {
		j.update()
	}

	if j.getLastInst().Status != FAILED {
		j.runCommand()
	}

	// todo: akelmore - figure out what to do with this
	// saveConfig()
}

func (j *Job) firstTimeSetup() {
	log.Println("Running first time setup for:", j.Name)

	var b bytes.Buffer
	multi := io.MultiWriter(&b, os.Stdout)

	// order to do things:
	// 1. Clone git repo
	// 2. Read in config to see if we need anything else
	// 3. Save Config
	// 4. Run

	inst := j.getLastInst()

	cmd := exec.Command("git", "clone", inst.Config.Repo, inst.Config.Git.path(j.Name))
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		log.Println("Error doing first time setup for:", j.Name)
		inst.Status = FAILED
		return
	}

	b.Reset()
	cmd = exec.Command("git", "-C", inst.Config.Git.path(j.Name), "config", "user.email", inst.Config.Git.Email)
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		log.Println("Error trying to set git email for:", j.Name)
		inst.Status = FAILED
		// todo: akelmore - clean up
		return
	}

	b.Reset()
	cmd = exec.Command("git", "-C", inst.Config.Git.path(j.Name), "config", "user.name", inst.Config.Git.Username)
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		log.Println("Error trying to set git username for:", j.Name)
		inst.Status = FAILED
		// todo: akelmore - clean up
		return
	}

	file, err := ioutil.ReadFile(inst.Config.Git.path(j.Name) + j.Config.BuildConfig)
	if err != nil {
		log.Printf("Error reading build config file: %v for job: %v\n", j.Config.BuildConfig, j.Name)
		inst.Status = FAILED
		// todo: akelmore - clean up
		return
	}

	err = json.Unmarshal(file, &j.Config)
	if err != nil {
		log.Printf("Error reading JSON from build config file: %v for job: %v\n", j.Config.BuildConfig, j.Name)
		inst.Status = FAILED
		// todo: akelmore - clean up
		return
	}
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

	cmd := exec.Command("git", "-C", inst.Config.Git.path(j.Name), "ls-remote", "origin", "-h", "HEAD")
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		return false
	}

	remoteHead := string(bytes.Fields(b.Bytes())[0])

	b.Reset()
	cmd = exec.Command("git", "-C", inst.Config.Git.path(j.Name), "rev-parse", "HEAD")
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

	var b bytes.Buffer
	multi := io.MultiWriter(&b, os.Stdout)

	cmd := exec.Command("git", "-C", inst.Config.Git.path(j.Name), "pull")
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		log.Println("Error pulling git for:", j.Name)
		inst.Status = FAILED
	} else if bytes.Contains(b.Bytes(), []byte("Already up-to-date.")) {
		log.Println("Something went wrong with the git pull, it was already up to date. It shouldn't have been.")
		inst.Status = FAILED
	}
}

func (j *Job) runCommand() {
	log.Println("Running command for:", j.Name)

	inst := j.getLastInst()

	fields := strings.Fields(j.Config.BuildCommand)
	if len(fields) > 0 {
		var b bytes.Buffer
		multi := io.MultiWriter(&b, os.Stdout)
		cmd := exec.Command(fields[0], fields[1:]...)
		cmd.Stdout = multi
		cmd.Stderr = multi
		cmd.Dir = inst.Config.Git.path(j.Name)
		if err := cmd.Run(); err != nil {
			log.Println("ERROR RUNNING BUILD:", j.Name)
			inst.Status = FAILED
			return
		}
	}
}
