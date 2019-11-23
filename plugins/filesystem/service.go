package filesystem

import (
	"context"
)

type Mount struct {
	Name string
	Path string
}

type Service struct {
	name       string
	mountsRoot string
	mounts     []Mount
}

func NewService(name string, mounts []Mount) *Service {
	return &Service{
		name:   name,
		mounts: mounts,
	}
}

func (s *Service) Start() error {
	return nil
}

func (s Service) WaitForPortAvailable(ctx context.Context) error {
	return nil
}

// // Parse filename like `mountName/path/to/file` to `path/to/mount/path/to/file`
// func (s Service) resolveFileName(filename string) (resolvedFileName string, err error) {
// 	nameArr := strings.SplitN(filename, "/", 2)
// 	if len(nameArr) < 2 {
// 		return "", fmt.Errorf("name of file should be in form `mountName/fileName`, not %q", filename)
// 	}
// 	mountName := nameArr[0]
// 	fileName := nameArr[1]
// 	var mount *Mount
// 	for _, m := range s.mounts {
// 		if m.Name == mountName {
// 			mount = &m
// 			break
// 		}
// 	}
// 	if mount == nil {
// 		return "", fmt.Errorf("mount with name %q not defined", mountName)
// 	}
// 	return filepath.Join(mount.Path, fileName), nil
// }
