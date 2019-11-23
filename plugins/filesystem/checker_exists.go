package filesystem

import (
	"fmt"
	"integration_framework/helper"
	"integration_framework/plugins"
	"os"
	"path/filepath"
)

type FileExistsChecker struct {
	filename string
	exists   bool
}

func NewFileExistsChecker(filename string, exists bool) *FileExistsChecker {
	return &FileExistsChecker{
		filename: filename,
		exists:   exists,
	}
}

func (pc FileExistsChecker) Check(mountsRoot string, saveResult plugins.FnResultSaver, variables map[string]interface{}) error {
	filename, err := helper.ApplyInterpolation(pc.filename, variables)
	if err != nil {
		return fmt.Errorf("unable to interpolate %q: %v", pc.filename, err)
	}
	pathToFile := filepath.Join(mountsRoot, filename)
	_, err = os.Stat(pathToFile)
	if os.IsNotExist(err) {
		if pc.exists {
			return fmt.Errorf("file %q not exists", pathToFile)
		}
		return nil
	} else if err == nil {
		if !pc.exists {
			return fmt.Errorf("file %q exists", pathToFile)
		}
		return nil
	}
	return fmt.Errorf("unable to check file existence: %v", err)
}
