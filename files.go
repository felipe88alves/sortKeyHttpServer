package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func getFilesInRelativePath(relativePath string, basePath string) ([]fs.DirEntry, error) {
	fullPath := filepath.Join(basePath, relativePath)
	dirEntries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	return filterFiles(dirEntries)
}

func mustGetBasePath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for _, err := os.ReadFile(filepath.Join(dir, "go.mod")); err != nil && len(dir) > 1; {
		dir = filepath.Dir(dir)
		_, err = os.ReadFile(filepath.Join(dir, "go.mod"))
	}

	log.Printf("Base directory: %v", dir)
	return dir, nil
}

func filterFiles(dirEntries []fs.DirEntry) ([]fs.DirEntry, error) {
	var files []fs.DirEntry
	for _, dirEntry := range dirEntries {
		// TODO: Improve identification of "regular" files. Suggestion: use dirEntry.Type() or dirEntry.Info()
		if strings.HasSuffix(dirEntry.Name(), "cfg") || strings.HasSuffix(dirEntry.Name(), "json") {
			files = append(files, dirEntry)
		}
	}
	return files, nil
}

func mustGetFile(relativePath string, basePath string) ([]byte, error) {
	absolutePath := filepath.Join(basePath, relativePath)
	file, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}
