package directories

import (
	"fmt"
	"io/fs"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

/*
	GetAllDirs returns a list of all directories, including subdirectories, for the provided root directories.

It skips directories that cannot be accessed and logs a warning.
*/
func GetAllDirs(rootDirs []string) ([]string, error) {
	var allDirs []string

	for _, root := range rootDirs {
		subDirs, err := collectAllSubDirs(root)
		if err != nil {
			log.Errorf("Error processing root directory %s: %v", root, err)
			continue // skip this directory with errors
		}
		allDirs = append(allDirs, subDirs...)
	}

	return allDirs, nil
}

/*
	collectAllSubDirs collects all subdirectories for the provided root directory.

It returns an error if it fails to access the root directory.
*/
func collectAllSubDirs(root string) ([]string, error) {
	var dirs []string

	err := safeWalkDir(root, func(path string, d fs.DirEntry) error {
		if d.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to collect subdirectories for %s: %w", root, err)
	}

	return dirs, nil
}

func safeWalkDir(root string, fn func(string, fs.DirEntry) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if fsErr, ok := err.(*fs.PathError); ok {
				log.Warnf("Warning: cannot access %s: %v", fsErr.Path, fsErr.Err)
				// skip this directory after logging the error
				return nil
			}
			// return the error if it's not a path error
			return err
		}

		return fn(path, d)
	})
}
