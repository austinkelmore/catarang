package scm

import (
	"bytes"
	"errors"
	"io"
	"log"
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
		log.Println("Error doing first time setup for: " + g.OnlineRepo)
		return err
	}

	b.Reset()
	cmd = exec.Command("git", "-C", g.LocalRepo, "config", "user.email", g.Auth.Email)
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		log.Println("Error trying to set git email for: " + g.Auth.Email)
		return err
	}

	b.Reset()
	cmd = exec.Command("git", "-C", g.LocalRepo, "config", "user.name", g.Auth.Username)
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		log.Println("Error trying to set git username for: " + g.Auth.Username)
		return err
	}

	return nil
}

func (g *Git) Poll() bool {
	// todo: akelmore - pull this multiwriter into Job so it can be output on the web
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

func (g *Git) UpdateExisting() error {
	return nil

	var b bytes.Buffer
	multi := io.MultiWriter(&b, os.Stdout)

	// update the git repo
	// todo: akelmore - pull into the git scm module
	cmd := exec.Command("git", "-C", g.LocalRepo, "pull")
	cmd.Stdout = multi
	cmd.Stderr = multi
	if err := cmd.Run(); err != nil {
		log.Println("Error pulling git")
		return err
	} else if bytes.Contains(b.Bytes(), []byte("Already up-to-date.")) {
		return errors.New("Something went wrong with the git pull, it was already up to date. It shouldn't have been.")
	}

	return nil
}
