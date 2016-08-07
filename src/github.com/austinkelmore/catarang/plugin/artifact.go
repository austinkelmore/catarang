package plugin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/austinkelmore/catarang/ulog"
	"github.com/pkg/errors"
)

var destStorage = "results/"

type Artifact struct {
	ToSave string
}

// todo: akelmore - fix artifact saving to be more robust
// todo: akelmore - fix artifact logging to go to the steplog
func (a *Artifact) Save(jobName string, instNum int) error {
	destPath := filepath.Join(destStorage, jobName, fmt.Sprintf("%d", instNum+1), a.ToSave)
	err := os.MkdirAll(filepath.Dir(destPath), 0777)
	if err != nil {
		return errors.Wrapf(err, "Can't create directory structure \"%s\"", destPath)
	}

	// todo: akelmore - make artifact saving more resilient to save and copy errors
	srcPath := a.ToSave
	src, err := os.Open(srcPath)
	if err != nil {
		return errors.Wrapf(err, "Can't open file \"%s\"", srcPath)
	}
	defer src.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return errors.Wrapf(err, "Can't create artifact file \"%s\"", destPath)
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	if err != nil {
		return errors.Wrapf(err, "Can't copy file from \"%s\" to \"%s\"", srcPath, destPath)
	}
	return nil
}

func (a *Artifact) Run(logger *ulog.StepLog) error {
	// todo: akelmore - make artifact saving work again
	return nil
}

func (a Artifact) GetName() string {
	return "artifact"
}
