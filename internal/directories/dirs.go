package directories

import (
	"fmt"
	"io/fs"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func GetAllDirs(dirs []string) ([]string, error) {
	var allDirs []string

	for _, dir := range dirs {
		children, err := getChildrenDirs(dir)
		if err != nil {
			return nil, fmt.Errorf("error processing directory %s: %w", dir, err)
		}
		allDirs = append(allDirs, children...)
	}

	return allDirs, nil
}

func getChildrenDirs(dir string) ([]string, error) {
	var dirs []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if fsErr, ok := err.(*fs.PathError); ok {
				log.Warnf("Warning: cannot access %s: %v", fsErr.Path, fsErr.Err)
				return nil
			}
			return err
		}

		if d.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return dirs, nil
}
