package scm

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
)

func CreateGit(localPath string, onlineRepo string) Git {
	// todo: akelmore - extract out email and username
	// todo: akelmore - make the local repo root configurable
	return Git{Auth: Authentication{Email: "catarang@austinkelmore.com", Username: "catarang"},
		LocalRepo: localPath, OnlineRepo: onlineRepo}
}

type Authentication struct {
	Username string
	Email    string
}

type Git struct {
	Auth       Authentication
	LocalRepo  string
	OnlineRepo string
}

func (g *Git) FirstTimeSetup() error {

	// todo: akelmore - pull out the multiwriter into main catarang part and pass in
	var b bytes.Buffer
	multi := io.MultiWriter(&b, os.Stdout)

	// order to do things:
	// 1. Clone git repo
	// 2. Read in config to see if we need anything else
	// 3. Save Config
	// 4. Run

	cmd := exec.Command("git", "clone", g.OnlineRepo, g.LocalRepo)
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		return errors.New("Error doing first time setup for: " + g.OnlineRepo)
	}

	b.Reset()
	cmd = exec.Command("git", "-C", g.LocalRepo, "config", "user.email", g.Auth.Email)
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		return errors.New("Error trying to set git email for: " + g.Auth.Email)
	}

	b.Reset()
	cmd = exec.Command("git", "-C", g.LocalRepo, "config", "user.name", g.Auth.Username)
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		return errors.New("Error trying to set git username for: " + g.Auth.Username)
	}

	return nil
}
