package utils

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var BasePath string

func init() {
	var err error
	if BasePath, err = mustGetBasePath(); err != nil {
		log.Fatalf("Failed to identify working directory: Error: %s", err.Error())
	}
}

func GetFilesInRelativePathByType(relativePath, typeFilter string) ([]fs.DirEntry, error) {
	fullPath := filepath.Join(BasePath, relativePath)
	dirEntries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	if len(dirEntries) == 0 {
		return nil, fmt.Errorf("the %s folder is empty", fullPath)
	}
	files := filterFiles(dirEntries, typeFilter)
	if len(files) == 0 {
		return nil, fmt.Errorf("no files with file type %q were found in %s",
			files, fullPath)
	}
	return files, nil
}

func mustGetBasePath() (string, error) {
	wdir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for _, err := os.ReadFile(filepath.Join(wdir, "go.mod")); err != nil && len(wdir) > 1; {
		wdir = filepath.Dir(wdir)
		_, err = os.ReadFile(filepath.Join(wdir, "go.mod"))
	}
	return wdir, nil
}

func filterFiles(dirEntries []fs.DirEntry, typeFilter string) []fs.DirEntry {
	var files []fs.DirEntry
	for _, dirEntry := range dirEntries {
		if strings.HasSuffix(dirEntry.Name(), typeFilter) {
			files = append(files, dirEntry)
		}
	}
	return files
}

func MustGetFile(relativeFilePath string) ([]byte, error) {
	absolutePath := filepath.Join(BasePath, relativeFilePath)
	file, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}
