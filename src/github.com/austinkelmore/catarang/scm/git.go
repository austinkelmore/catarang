package scm

import (
	"bytes"
	"errors"

	"github.com/austinkelmore/catarang/ulog"
)

// NewGit Creates the git handler
func NewGit(origin string) *Git {
	// todo: akelmore - extract out email and username
	return &Git{Auth: Authentication{Email: "catarang@austinkelmore.com", Username: "catarang"},
		Origin: origin}
}

// Authentication authentication info for the git handler
type Authentication struct {
	Username string
	Email    string
}

// Git The git handler
type Git struct {
	Auth   Authentication
	Origin string

	CompletedSetup bool
}

// FirstTimeSetup Clone the git repository and setup the email and username
func (g Git) FirstTimeSetup(logger *ulog.StepLog) error {
	cmd := logger.New("git", "clone", g.Origin, ".")
	if err := cmd.Run(); err != nil {
		return errors.New("Error doing first time setup for: " + g.Origin)
	}

	return nil
}

// Poll polls the git master to see if the local repository is different from the master's head
func (g *Git) Poll(logger *ulog.StepLog) (bool, error) {
	lsremote := logger.New("git", "ls-remote", "origin", "-h", "HEAD")
	if err := lsremote.Run(); err != nil {
		return false, errors.New("Error polling head of origin repo: " + err.Error())
	}

	revparse := logger.New("git", "rev-parse", "HEAD")
	if err := revparse.Run(); err != nil {
		return false, errors.New("Error finding head of local repo: " + err.Error())
	}

	// empty repositories don't return any text since they have no HEAD
	if len(lsremote.Str) == 0 || len(revparse.Str) == 0 {
		return false, nil
	}

	remoteHead := string(bytes.Fields([]byte(lsremote.Str))[0])
	localHead := string(bytes.Fields([]byte(revparse.Str))[0])
	return remoteHead != localHead, nil
}

// UpdateExisting syncs the git repository
func (g *Git) UpdateExisting(logger *ulog.StepLog) error {
	cmd := logger.New("git", "pull")
	if err := cmd.Run(); err != nil {
		return errors.New("Error pulling git.")
	}

	return nil
}

func (g *Git) Run(logger *ulog.StepLog) bool {
	if g.CompletedSetup == false {
		if err := g.FirstTimeSetup(logger); err != nil {
			return false
		}
		g.CompletedSetup = true
	} else {
		if err := g.UpdateExisting(logger); err != nil {
			return false
		}
	}
	return true
}

func (g Git) GetName() string {
	return "git Plugin"
}
