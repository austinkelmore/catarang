package scm_test

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/austinkelmore/catarang/scm"
)

func TestGitExists(t *testing.T) {
	_, err := exec.LookPath("git")
	if err != nil {
		t.Error("", err)
	}
}

func TestGitFirstTimeSetup(t *testing.T) {

	path := "../tests/test_repos/FirstTimeSetup"

	// clean up previous tests
	os.RemoveAll(path)

	git := scm.NewGit(path, "../tests/git_test")

	b := new(bytes.Buffer)
	w := bufio.NewWriter(b)
	err := git.FirstTimeSetup(w, w)
	if err != nil {
		log.Printf("%s", b.Bytes())
		t.Error(err)
	}
}

func TestGitFirstTimeSetupFail(t *testing.T) {
	git := scm.NewGit("path", "bogus_repo_path")

	b := new(bytes.Buffer)
	w := bufio.NewWriter(b)
	err := git.FirstTimeSetup(w, w)
	if err == nil {
		t.Error("Expected failure for bogus repo path. No error returned.")
	}
}
