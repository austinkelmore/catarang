package plugin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/austinkelmore/catarang/jobdata"
	"github.com/austinkelmore/catarang/ulog"
	"github.com/pkg/errors"
)

var destStorage = "results/"

// Artifact stores off the specified paths/objects
type Artifact struct {
	ToSave []string
}

func save(srcDir, toSave, destDir string) error {
	srcPath := filepath.Join(srcDir, toSave)
	destPath := filepath.Join(destDir, toSave)
	if err := os.MkdirAll(filepath.Dir(destPath), 0777); err != nil {
		return errors.Wrapf(err, "can't create directory structure \"%s\"", destPath)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return errors.Wrapf(err, "can't open file \"%s\"", srcPath)
	}

	dest, err := os.Create(destPath)
	if err != nil {
		return errors.Wrapf(err, "can't create artifact file \"%s\"", destPath)
	}

	if _, err = io.Copy(dest, src); err != nil {
		return errors.Wrapf(err, "can't copy file from \"%s\" to \"%s\"", srcPath, destPath)
	}

	if err = src.Close(); err != nil {
		return errors.Wrapf(err, "can't close src file %s", srcPath)
	}
	if err = dest.Close(); err != nil {
		return errors.Wrapf(err, "can't close/save dest file %s", destPath)
	}
	return nil
}

// Run is the entry point into the Artifact plugin
func (a *Artifact) Run(job jobdata.Data, logger *ulog.StepLog) error {
	destPath := filepath.Join(destStorage, job.Name, fmt.Sprintf("%d/", job.TimesRun))
	for _, loc := range a.ToSave {
		if err := save(job.LocalPath, loc, destPath); err != nil {
			return errors.Wrapf(err, "couldn't save %s", loc)
		}
	}

	return nil
}

// GetName returns the name of the plugin
func (a Artifact) GetName() string {
	return "artifact"
}
