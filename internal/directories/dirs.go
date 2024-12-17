package directories

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

/*
	GetAllDirs returns a list of all directories, including subdirectories, for the provided root directories.

It skips directories that cannot be accessed and logs a warning.
*/
func GetAllDirs(rootDirs []string) (map[string]struct{}, error) {
	// var allDirs []string
	allDirs := make(map[string]struct{})

	for _, root := range rootDirs {
		subDirs, err := collectAllSubDirs(root)
		if err != nil {
			log.Errorf("Error processing root directory %s: %v", root, err)
			continue // skip this directory with errors
		}
		// allDirs = append(allDirs, subDirs...)
		allDirs = mergeRootDirs(allDirs, subDirs)
	}

	return allDirs, nil
}

/*
	collectAllSubDirs collects all subdirectories for the provided root directory.

It returns an error if it fails to access the root directory.
*/
func collectAllSubDirs(root string) (map[string]struct{}, error) {
	// var dirs []string
	dirs := make(map[string]struct{})

	err := safeWalkDir(root, func(path string, d fs.DirEntry) error {
		if d.IsDir() {
			// dirs = append(dirs, path)
			dirs[path] = struct{}{}
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

/*
	filterValidDirs filters out invalid directories from the list of paths.

It returns a list of valid directories.
*/
func FilterValidDirs(paths []string) []string {
	validDirs := []string{}
	for _, dir := range paths {
		absPath, err := filepath.Abs(dir)
		if err != nil {
			log.Errorf("Invalid path %s: %v", dir, err)
			continue
		}
		if stat, err := os.Stat(absPath); err == nil && stat.IsDir() {
			validDirs = append(validDirs, absPath)
		} else {
			log.Warnf("Skipping invalid or non-existent directory: %s", dir)
		}
	}
	return validDirs
}

/*
	prepareDirsForMonitoring prepares the list of directories for monitoring.

It filters out invalid directories and returns a list of all recursive directories.
*/
func PrepareDirsForMonitoring(paths []string) map[string]struct{} {
	validDirs := FilterValidDirs(paths)

	allRecursiveDirs, err := GetAllDirs(validDirs)
	if err != nil {
		log.Fatalf("Failed to get directories: %v", err)
	}
	return allRecursiveDirs
}

/*
	mergeRootDirs merges two maps of root directories.

It returns a new map with all unique directories from both maps.
*/
func mergeRootDirs(d1, d2 map[string]struct{}) map[string]struct{} {
	result := make(map[string]struct{})

	for k := range d1 {
		result[k] = struct{}{}
	}

	for k := range d2 {
		result[k] = struct{}{}
	}

	return result
}
