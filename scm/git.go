package scm

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
)

// NewGit Creates the git handler
func NewGit(localPath string, onlineRepo string) *Git {
	// todo: akelmore - extract out email and username
	return &Git{Auth: Authentication{Email: "catarang@austinkelmore.com", Username: "catarang"},
		LocalRepo: localPath, OnlineRepo: onlineRepo}
}

// Authentication authentication info for the git handler
type Authentication struct {
	Username string
	Email    string
}

// Git The git handler
type Git struct {
	Auth       Authentication
	LocalRepo  string
	OnlineRepo string
}

// FirstTimeSetup Clone the git repository and setup the email and username
func (g Git) FirstTimeSetup(outWriter io.Writer, errWriter io.Writer) error {
	// order to do things:
	// 1. Clone git repo
	// 2. Read in config to see if we need anything else
	// 3. Save Config
	// 4. Run

	cmd := exec.Command("git", "clone", g.OnlineRepo, g.LocalRepo)
	cmd.Stdout = outWriter
	cmd.Stderr = errWriter
	if err := cmd.Run(); err != nil {
		log.Println("Error doing first time setup for: " + g.OnlineRepo)
		return err
	}

	cmd = exec.Command("git", "-C", g.LocalRepo, "config", "user.email", g.Auth.Email)
	cmd.Stdout = outWriter
	cmd.Stderr = errWriter
	if err := cmd.Run(); err != nil {
		log.Println("Error trying to set git email for: " + g.Auth.Email)
		return err
	}

	cmd = exec.Command("git", "-C", g.LocalRepo, "config", "user.name", g.Auth.Username)
	cmd.Stdout = outWriter
	cmd.Stderr = errWriter
	if err := cmd.Run(); err != nil {
		log.Println("Error trying to set git username for: " + g.Auth.Username)
		return err
	}

	return nil
}

// Poll polls the git master to see if the local repository is different from the master's head
func (g *Git) Poll() bool {
	var b bytes.Buffer
	multi := io.MultiWriter(&b, os.Stdout)

	cmd := exec.Command("git", "-C", g.LocalRepo, "ls-remote", "origin", "-h", "HEAD")
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		log.Println("Error polling head of git repo: " + err.Error())
		return false
	}

	remoteHead := string(bytes.Fields(b.Bytes())[0])

	b.Reset()
	cmd = exec.Command("git", "-C", g.LocalRepo, "rev-parse", "HEAD")
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		log.Println("Error finding head of local repo: " + err.Error())
		return false
	}

	localHead := string(bytes.Fields(b.Bytes())[0])

	return remoteHead != localHead
}

// UpdateExisting syncs the git repository
func (g *Git) UpdateExisting(outWriter *io.Writer, errWriter *io.Writer) error {
	var b bytes.Buffer
	multi := io.MultiWriter(&b, *outWriter)

	// update the git repo
	cmd := exec.Command("git", "-C", g.LocalRepo, "pull")
	cmd.Stdout = multi
	cmd.Stderr = *errWriter
	if err := cmd.Run(); err != nil {
		log.Println("Error pulling git")
		return err
	} else if bytes.Contains(b.Bytes(), []byte("Already up-to-date.")) {
		return errors.New("Something went wrong with the git pull, it was already up to date. It shouldn't have been.")
	}

	return nil
}

// LocalRepoPath returns the local path to the repository
func (g *Git) LocalRepoPath() string {
	return g.LocalRepo
}
