// todo: akelmore - move the artifact out of job?
package job

import (
	"io"
	"log"
	"os"
)

var localStorage = "results/"

type Artifact struct {
	ToSave string
}

// todo: akelmore - fix artifact saving to be more robust
func (a *Artifact) Save(localRepoPath string) error {
	err := os.MkdirAll(localStorage, 0666)
	if err != nil {
		log.Println("can't create dir structure")
		return err
	}

	// todo: akelmore - this is the completely wrong thing to do for artifact saving, but it's a good prototype
	// fixup later
	srcPath := localRepoPath + a.ToSave
	src, err := os.Open(srcPath)
	if err != nil {
		log.Println("can't open srcpath", srcPath)
		return err
	}
	defer src.Close()

	destPath := localStorage + a.ToSave
	dest, err := os.Create(destPath)
	if err != nil {
		log.Println("can't create file", destPath)
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	if err != nil {
		log.Println("can't copy file")
		return err
	}
	return nil
}
