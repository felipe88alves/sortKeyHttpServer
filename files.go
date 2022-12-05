package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func getFilesInRelativePath(relativePath, basePath string) ([]fs.DirEntry, error) {
	fullPath := filepath.Join(basePath, relativePath)
	dirEntries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	return filterFiles(dirEntries), nil
}

func mustGetBasePath(wdir string) string {
	for _, err := os.ReadFile(filepath.Join(wdir, "go.mod")); err != nil && len(wdir) > 1; {
		wdir = filepath.Dir(wdir)
		_, err = os.ReadFile(filepath.Join(wdir, "go.mod"))
	}

	log.Printf("Base directory: %v", wdir)
	return wdir
}

func filterFiles(dirEntries []fs.DirEntry) []fs.DirEntry {
	var files []fs.DirEntry
	for _, dirEntry := range dirEntries {
		// TODO: Improve identification of "regular" files. Suggestion: use dirEntry.Type() or dirEntry.Info()
		if strings.HasSuffix(dirEntry.Name(), validUrlSuffixFileTypeCfg) || strings.HasSuffix(dirEntry.Name(), validUrlSuffixFileTypeJson) {
			files = append(files, dirEntry)
		}
	}
	return files
}

func mustGetFile(basePath, relativeFilePath string) ([]byte, error) {
	absolutePath := filepath.Join(basePath, relativeFilePath)
	file, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}
