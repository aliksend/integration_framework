package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
)

func NewClearPrepare() *ClearPrepare {
	return &ClearPrepare{}
}

type ClearPrepare struct {
}

func (pp ClearPrepare) Prepare(mountsRoot string, mounts []Mount) error {
	for _, mount := range mounts {
		path := filepath.Join(mountsRoot, mount.Name)
		d, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("unable to open directory %q: %v", path, err)
		}
		defer d.Close()
		names, err := d.Readdirnames(-1)
		if err != nil {
			return fmt.Errorf("unable to readdirs in %q: %v", path, err)
		}
		for _, name := range names {
			filename := filepath.Join(path, name)
			err = os.RemoveAll(filename)
			if err != nil {
				return fmt.Errorf("unable to remove contents of %q: %v", filename, err)
			}
		}
	}
	return nil
}
