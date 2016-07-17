// todo: akelmore - move the artifact out of job?
package job

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

var destStorage = "results/"

type Artifact struct {
	ToSave string
}

// todo: akelmore - fix artifact saving to be more robust
func (a *Artifact) Save(jobName string, instNum int) error {
	destPath := destStorage + jobName + "/" + fmt.Sprintf("%d", instNum+1) + "/" + a.ToSave
	err := os.MkdirAll(filepath.Dir(destPath), 0777)
	if err != nil {
		log.Println("can't create dir structure")
		return err
	}

	// todo: akelmore - this is the completely wrong thing to do for artifact saving, but it's a good prototype
	// fixup later
	srcPath := a.ToSave
	src, err := os.Open(srcPath)
	if err != nil {
		log.Println("can't open srcpath", srcPath)
		return err
	}
	defer src.Close()

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
