package scm

import (
	"bytes"
	"errors"

	"github.com/austinkelmore/catarang/splitlog"
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
func (g Git) FirstTimeSetup(cmds []splitlog.CmdLog) error {
	cmd := splitlog.New(cmds, "git", "clone", g.Origin, g.LocalRepo)
	if err := cmd.Run(); err != nil {
		return errors.New("Error doing first time setup for: " + g.Origin)
	}

	return nil
}

// Poll polls the git master to see if the local repository is different from the master's head
func (g *Git) Poll(cmds []splitlog.CmdLog) (bool, error) {
	lsremote := splitlog.New(cmds, "git", "-C", g.LocalRepo, "ls-remote", "origin", "-h", "HEAD")
	if err := lsremote.Run(); err != nil {
		return false, errors.New("Error polling head of origin repo: " + err.Error())
	}

	revparse := splitlog.New(cmds, "git", "-C", g.LocalRepo, "rev-parse", "HEAD")
	if err := revparse.Run(); err != nil {
		return false, errors.New("Error finding head of local repo: " + err.Error())
	}

	// empty repositories don't return any text since they have no HEAD
	if len(lsremote.Bytes()) == 0 || len(revparse.Bytes()) == 0 {
		return false, nil
	}

	remoteHead := string(bytes.Fields(lsremote.Bytes())[0])
	localHead := string(bytes.Fields(revparse.Bytes())[0])
	return remoteHead != localHead, nil
}

// UpdateExisting syncs the git repository
func (g *Git) UpdateExisting(cmds []splitlog.CmdLog) error {

	cmd := splitlog.New(cmds, "git", "-C", g.LocalRepo, "pull")
	if err := cmd.Run(); err != nil {
		return errors.New("Error pulling git.")
	} else if bytes.Contains(cmd.Bytes(), []byte("Already up-to-date.")) {
		return errors.New("Something went wrong with git pull, it was already up to date. It shouldn't have been.")
	}

	return nil
}

// LocalRepoPath returns the local path to the repository
func (g *Git) LocalRepoPath() string {
	return g.LocalRepo
}
