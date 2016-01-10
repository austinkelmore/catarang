package scm

import (
	"bytes"
	"errors"
	"io"
	"os/exec"

	"github.com/austinkelmore/catarang/multilog"
)

// NewGit Creates the git handler
func NewGit(origin string, localPath string) *Git {
	// todo: akelmore - extract out email and username
	return &Git{Auth: Authentication{Email: "catarang@austinkelmore.com", Username: "catarang"},
		LocalRepo: localPath, Origin: origin}
}

// Authentication authentication info for the git handler
type Authentication struct {
	Username string
	Email    string
}

// Git The git handler
type Git struct {
	Auth      Authentication
	LocalRepo string
	Origin    string
}

// FirstTimeSetup Clone the git repository and setup the email and username
func (g Git) FirstTimeSetup(logger *multilog.Log) error {
	cmd := exec.Command("git", "clone", g.Origin, g.LocalRepo)
	cmd.Stdout = &logger.Out
	cmd.Stderr = &logger.Err
	if err := cmd.Run(); err != nil {
		return errors.New("Error doing first time setup for: " + g.Origin)
	}

	return nil
}

// Poll polls the git master to see if the local repository is different from the master's head
func (g *Git) Poll(logger *multilog.Log) (bool, error) {
	var b bytes.Buffer
	multi := io.MultiWriter(&b, &logger.Out)
	cmd := exec.Command("git", "-C", g.LocalRepo, "ls-remote", "origin", "-h", "HEAD")
	cmd.Stdout = multi
	cmd.Stderr = &logger.Err
	if err := cmd.Run(); err != nil {
		return false, errors.New("Error polling head of origin repo: " + err.Error())
	}

	var b2 bytes.Buffer
	multi2 := io.MultiWriter(&b2, &logger.Out)
	cmd = exec.Command("git", "-C", g.LocalRepo, "rev-parse", "HEAD")
	cmd.Stdout = multi2
	cmd.Stderr = &logger.Err
	if err := cmd.Run(); err != nil {
		return false, errors.New("Error finding head of local repo: " + err.Error())
	}

	// empty repositories don't return any text since they have no HEAD
	if len(b.Bytes()) == 0 || len(b2.Bytes()) == 0 {
		return false, nil
	}

	remoteHead := string(bytes.Fields(b.Bytes())[0])
	localHead := string(bytes.Fields(b2.Bytes())[0])
	return remoteHead != localHead, nil
}

// UpdateExisting syncs the git repository
func (g *Git) UpdateExisting(logger *multilog.Log) error {
	// todo: akelmore - do i want to trust that multilog is empty, or should i assume it's not?
	// todo: akelmore - check every place where multilogger is used
	cmd := exec.Command("git", "-C", g.LocalRepo, "pull")
	cmd.Stdout = &logger.Out
	cmd.Stderr = &logger.Err
	if err := cmd.Run(); err != nil {
		return errors.New("Error pulling git.")
	} else if bytes.Contains(logger.Out.Bytes(), []byte("Already up-to-date.")) {
		return errors.New("Something went wrong with git pull, it was already up to date. It shouldn't have been.")
	}

	return nil
}

// LocalRepoPath returns the local path to the repository
func (g *Git) LocalRepoPath() string {
	return g.LocalRepo
}
