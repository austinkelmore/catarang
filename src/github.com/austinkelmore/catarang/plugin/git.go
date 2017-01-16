package plugin

import (
	"bytes"

	"github.com/austinkelmore/catarang/cmd"
	"github.com/pkg/errors"
)

// Authentication authentication info for the git handler
type Authentication struct {
	Username string
	Email    string
}

// Git The git handler
type Git struct {
	Auth   Authentication
	Origin string
}

// FirstTimeSetup Clone the git repository and setup the email and username
func (g Git) FirstTimeSetup(logger *cmd.Log) error {
	cmd := logger.New("git", "clone", g.Origin, ".")
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "error trying to git clone \"%s\"", g.Origin)
	}

	return nil
}

// Poll polls the git master to see if the local repository is different from the master's head
func (g *Git) Poll(logger *cmd.Log) (bool, error) {
	lsremote := logger.New("git", "ls-remote", "origin", "-h", "HEAD")
	if err := lsremote.Run(); err != nil {
		return false, errors.Wrapf(err, "error polling head of origin repo \"%s\" at dir \"%s\"", g.Origin, logger.WorkingDir)
	}

	revparse := logger.New("git", "rev-parse", "HEAD")
	if err := revparse.Run(); err != nil {
		return false, errors.Wrapf(err, "error finding head of local repo at dir \"%s\"", logger.WorkingDir)
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
func (g *Git) UpdateExisting(logger *cmd.Log) error {
	cmd := logger.New("git", "pull")
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "error pulling git at dir \"%s\"", logger.WorkingDir)
	}

	return nil
}

// Run is the entry point into the git plugin
func (g *Git) Run(logger *cmd.Log) error {
	if err := g.UpdateExisting(logger); err != nil {
		return errors.Wrap(err, "error running git's UpdateExisting")
	}

	return nil
}

// GetName returns the name of the plugin
func (g Git) GetName() string {
	return "git"
}

// SetOrigin sets the origin of the git repository
func (g *Git) SetOrigin(origin string) error {
	g.Origin = origin
	return nil
}
