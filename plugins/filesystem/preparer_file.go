package filesystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func NewFilePrepare(filename string, content string) *FilePrepare {
	return &FilePrepare{
		filename: filename,
		content:  content,
	}
}

type FilePrepare struct {
	filename string
	content  string
}

func (pp FilePrepare) Prepare(mountsRoot string, mounts []Mount) error {
	pathToFile := filepath.Join(mountsRoot, pp.filename)
	err := os.MkdirAll(filepath.Dir(pathToFile), os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create directories for file %q: %v", pathToFile, err)
	}
	fmt.Printf(".. create file %q with content %q\n", pp.filename, pp.content)
	// TODO единообразно логгировать работу всех preparer-ов и checker-ов чтобы по логу тестов можно было понять что упало и почему (баг внутри приложения или integration_framework)
	err = ioutil.WriteFile(pathToFile, []byte(pp.content), os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to write file %q: %v", pathToFile, err)
	}
	return nil
}
