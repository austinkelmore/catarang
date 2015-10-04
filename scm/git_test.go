package scm_test

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/austinkelmore/catarang/scm"
)

// this function has to exist because Go's os.RemoveAll doesn't remove locked files
// on Windows, which git has for some reason
func forceRemoveAll(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
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

func createTestRepo(t *testing.T, path string) error {
	// clear out the path
	err := forceRemoveAll(path)
	if err != nil {
		t.Logf("Error removing files: %s", err.Error())
	}
	out, err := exec.Command("git", "init", path).CombinedOutput()
	if err != nil {
		log.Println(out)
		return err
	}

	text := "Test file for git tests."
	filename := "file.txt"
	err = ioutil.WriteFile(path+filename, []byte(text), 0644)
	if err != nil {
		log.Println("Couldn't write test file in test repo.")
		return err
	}

	out, err = exec.Command("git", "-C", path, "add", filename).CombinedOutput()
	if err != nil {
		log.Printf("Couldn't add test file to git.\n%s", out)
		return err
	}

	out, err = exec.Command("git", "-C", path, "commit", "-m", "Initial commit for test git repository.").CombinedOutput()
	if err != nil {
		log.Printf("Couldn't commit file to git.\n%s", out)
		return err
	}

	return nil
}

func TestGitExists(t *testing.T) {
	_, err := exec.LookPath("git")
	if err != nil {
		t.Error("", err)
	}
}

func TestGitFirstTimeSetup(t *testing.T) {
	path := "../tests/git_test/"
	newpath := "../tests/FirstTimeSetup"
	forceRemoveAll(newpath)

	err := createTestRepo(t, path)
	if err != nil {
		t.Error(err)
		return
	}

	git := scm.NewGit(newpath, path)
	b := new(bytes.Buffer)
	w := bufio.NewWriter(b)
	err = git.FirstTimeSetup(w, w)
	if err != nil {
		log.Printf("%s", b.Bytes())
		t.Error(err)
	}
}

func TestGitFirstTimeSetupFail(t *testing.T) {
	git := scm.NewGit("../tests/FirstTimeSetupFail", "bogus_repo_path")

	b := new(bytes.Buffer)
	w := bufio.NewWriter(b)
	err := git.FirstTimeSetup(w, w)
	if err == nil {
		t.Error("Expected failure for bogus repo path. No error returned.")
	}
}
