package scm_test

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/austinkelmore/catarang/scm"
)

// this function has to exist because Go's os.RemoveAll doesn't remove locked files
// on Windows, which git has for some reason
func forceRemoveAll(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return nil
	}
	if !fi.IsDir() {
		err := os.Chmod(path, 0666)
		if err != nil {
			return err
		}
	}
	fd, err := os.Open(path)
	defer fd.Close()
	if err != nil {
		return err
	}
	names, _ := fd.Readdirnames(-1)
	for _, name := range names {
		err = forceRemoveAll(path + string(filepath.Separator) + name)
		if err != nil {
			return err
		}
	}
	os.RemoveAll(path)
	return nil
}

func createTestRepo(t *testing.T, origin string, clone string) error {
	// clear out the origin if it exists and start from scratch
	err := forceRemoveAll(origin)
	if err != nil {
		t.Logf("Error removing files: %s\n", err.Error())
		return err
	}
	out, err := exec.Command("git", "init", origin).CombinedOutput()
	if err != nil {
		t.Log(out)
		return err
	}

	filename := "file.txt"
	// write, add, and commit file initially
	err = ioutil.WriteFile(origin+filename, []byte("Test file for git tests."), 0644)
	if err != nil {
		t.Log("Couldn't write test file in test repo.")
		return err
	}
	out, err = exec.Command("git", "-C", origin, "add", filename).CombinedOutput()
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

	// start the clone from scratch as well
	err = forceRemoveAll(clone)
	if err != nil {
		t.Logf("Error removing files. %s\n", err.Error())
		return err
	}

	return nil
}

func setupTest(t *testing.T, origin string, clone string) (*scm.Git, error) {
	err := createTestRepo(t, origin, clone)
	if err != nil {
		t.Error(err)
		return nil, err
	}

	git := scm.NewGit(origin, clone)
	b := new(bytes.Buffer)
	rw := bufio.NewWriter(b)
	err = git.FirstTimeSetup(rw, rw)
	return git, err
}

func TestGitExists(t *testing.T) {
	_, err := exec.LookPath("git")
	if err != nil {
		t.Error("", err)
	}
}

func TestGitFirstTimeSetup(t *testing.T) {
	_, err := setupTest(t, "../tests/FirstTimeSetupOrigin/", "../tests/FirstTimeSetup")
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGitFirstTimeSetupFail(t *testing.T) {
	git := scm.NewGit("bogus_repo_path", "../tests/FirstTimeSetupFail")

	b := new(bytes.Buffer)
	w := bufio.NewWriter(b)
	err := git.FirstTimeSetup(w, w)
	if err == nil {
		t.Error("Expected failure for bogus repo path. No error returned.")
	}
}

func TestGitPoll(t *testing.T) {
	origin := "../tests/PollOrigin/"
	testrepo := "../tests/Poll/"
	git, err := setupTest(t, origin, testrepo)
	if err != nil {
		t.Error(err)
	}
	shouldRun, err := git.Poll()
	if err != nil {
		t.Errorf("Error polling. %s\n", err.Error())
	}
	if shouldRun == true {
		t.Error("Expected to not have to run. Should be fully synced.")
	}

	// sync the git repository back one step so we can test polling when we need to sync
	out, err := exec.Command("git", "-C", testrepo, "log", "--oneline").CombinedOutput()
	if err != nil {
		t.Logf("Couldn't call git log on the test repo.\n%s\n", out)
		t.Error(err)
	}
	lines := strings.Split(string(out[:]), "\n")
	fields := strings.Fields(lines[1])
	out, err = exec.Command("git", "-C", testrepo, "reset", "--hard", fields[0]).CombinedOutput()
	if err != nil {
		t.Logf("Couldn't reset back to older changelist in git.\n%s\n", out)
		t.Error(err)
	}

	shouldRun, err = git.Poll()
	if err != nil {
		t.Errorf("Error polling. %s\n", err.Error())
	}
	if shouldRun == false {
		t.Error("Expected to have to run. Should NOT be fully synced.")
	}
}
