package remove

import (
	"os"
	"path/filepath"
)

// this function has to exist because Go's os.RemoveAll doesn't remove locked files
// on Windows, which git has for some reason
func ForceRemoveAll(path string) error {
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
		err = ForceRemoveAll(path + string(filepath.Separator) + name)
		if err != nil {
			return err
		}
	}
	os.RemoveAll(path)
	return nil
}
