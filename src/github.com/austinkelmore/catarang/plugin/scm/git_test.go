package scm_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/austinkelmore/catarang/plugin/scm"
	"github.com/austinkelmore/catarang/ulog"
)

// TestMain is the entry point for this file's tests
func TestMain(m *testing.M) {
	ret := m.Run()
	cleanUpTests()
	os.Exit(ret)
}

var localPath = "../tests/"

func cleanUpTests() {
	os.RemoveAll(localPath)
}

func initRepo(t *testing.T, origin string) error {
	// clear out the origin if it exists and start from scratch
	err := os.RemoveAll(origin)
	if err != nil {
		t.Logf("Error removing files: %s\n", err.Error())
		return err
	}
	out, err := exec.Command("git", "init", origin).CombinedOutput()
	if err != nil {
		t.Log(out)
		return err
	}

	return nil
}

func createTestRepo(t *testing.T, origin string) error {
	err := initRepo(t, origin)
	if err != nil {
		return err
	}

	filename := "file.txt"
	// write, add, and commit file initially
	err = ioutil.WriteFile(origin+filename, []byte("Test file for git tests."), 0644)
	if err != nil {
		t.Log("Couldn't write test file in test repo.")
		return err
	}
	out, err := exec.Command("git", "-C", origin, "add", filename).CombinedOutput()
	if err != nil {
		t.Logf("Couldn't add test file to git.\n%s\n", out)
		return err
	}
	out, err = exec.Command("git", "-C", origin, "commit", "-m", "Initial commit for test git repository.").CombinedOutput()
	if err != nil {
		t.Logf("Couldn't commit file to git.\n%s\n", out)
		return err
	}

	// write, add, and commit file for a second time
	err = ioutil.WriteFile(origin+filename, []byte("Second edit, test file for git tests."), 0644)
	if err != nil {
		t.Log("Couldn't write test file a second time.")
		return err
	}
	out, err = exec.Command("git", "-C", origin, "add", filename).CombinedOutput()
	if err != nil {
		t.Logf("Couldn't add test file to git a second time.\n%s", out)
		return err
	}
	out, err = exec.Command("git", "-C", origin, "commit", "-m", "Second commit for test git repo.").CombinedOutput()
	if err != nil {
		t.Logf("Couldn't commit file to git a second time.\n%s\n", out)
		return err
	}

	return nil
}

func setupGitClone(t *testing.T, origin string, clone string) (*scm.Git, error) {
	// start the clone from scratch as well
	err := os.RemoveAll(clone)
	if err != nil {
		t.Errorf("Error removing files. %s\n", err.Error())
		return nil, err
	}

	git := &scm.Git{}
	git.SetOrigin(origin)
	logger := ulog.StepLog{}
	logger.WorkingDir = clone
	err = git.FirstTimeSetup(&logger)
	return git, err
}

func setupBothRepos(t *testing.T, origin string, clone string) (*scm.Git, error) {
	err := createTestRepo(t, origin)
	if err != nil {
		t.Error(err)
		return nil, err
	}

	return setupGitClone(t, origin, clone)
}

func syncBackOneRev(t *testing.T, testrepo string) {
	// sync the git repository back one step so we can test polling when we need to sync
	out, err := exec.Command("git", "-C", testrepo, "log", "--oneline").CombinedOutput()
	if err != nil {
		t.Error(err)
	}
	lines := strings.Split(string(out[:]), "\n")
	fields := strings.Fields(lines[1])
	out, err = exec.Command("git", "-C", testrepo, "reset", "--hard", fields[0]).CombinedOutput()
	if err != nil {
		t.Error(err)
	}
}

func TestGitExists(t *testing.T) {
	_, err := exec.LookPath("git")
	if err != nil {
		t.Error("", err)
	}
}

func TestFirstTimeSetupFail(t *testing.T) {
	git := scm.Git{}
	git.SetOrigin("bogus_repo_path/")

	logger := ulog.StepLog{}
	logger.WorkingDir = localPath + "FirstTimeSetupFail/"
	err := git.FirstTimeSetup(&logger)
	if err == nil {
		t.Error("Expected failure for bogus repo path. No error returned.")
	}
}

// todo: akelmore - refactor
func TestSetupPollAndSync(t *testing.T) {
	origin := localPath + "PollOrigin/"
	testrepo := localPath + "Poll/"
	git, err := setupBothRepos(t, origin, testrepo)
	if err != nil {
		t.Error(err)
		return
	}
	logger := ulog.StepLog{}
	logger.WorkingDir = testrepo
	shouldRun, err := git.Poll(&logger)
	if err != nil {
		t.Errorf("Error polling. %s\n", err.Error())
	}
	if shouldRun == true {
		t.Error("Expected to not have to run. Should be fully synced.")
	}

	syncBackOneRev(t, testrepo)

	shouldRun, err = git.Poll(&logger)
	if err != nil {
		t.Errorf("Error polling. %s\n", err.Error())
	}
	if shouldRun == false {
		t.Error("Expected to have to run. Should NOT be fully synced.")
	}

	if err = git.UpdateExisting(&logger); err != nil {
		t.Errorf("Should have been able to update git repo.\n%s\n", err.Error())
	}
	logger.WorkingDir = "bogus_repo_path"

	// todo: akelmore - fix polling tests
	// Poll does two commands on the local repo, i know how to test
	// the first one, but figure out how to test the failure of the second
	// one (how does rev-parse fail)

	if err = git.UpdateExisting(&logger); err == nil {
		t.Error("Should not be able to update bogus local repo.")
	}

	if _, err = git.Poll(&logger); err == nil {
		t.Error("Should not be able to poll bogus local repo.")
	}
	git.Origin = "bogus_repo_path"
	if _, err = git.Poll(&logger); err == nil {
		t.Error("Should not be able to poll bogus origin repo.")
	}
}

func TestPollEmpty(t *testing.T) {
	origin := localPath + "PollEmptyOrigin"
	clone := localPath + "PollEmpty"
	initRepo(t, origin)
	git, err := setupGitClone(t, origin, clone)
	if err != nil {
		t.Error(err.Error())
		return
	}

	// todo: akelmore - do we want to be able to poll an empty repository?
	logger := ulog.StepLog{}
	logger.WorkingDir = clone
	shouldRun, err := git.Poll(&logger)
	if err == nil {
		t.Error("Should not be able to poll an empty repository.")
	}
	if shouldRun {
		t.Error("Should not want to run on an empty git repository after polling.")
	}
}
