package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// TODO: Explore use of fstest.MapFS to mock file system. Sample code below:
// fs := fstest.MapFS{
// 	"hello.txt": {
// 		Data: []byte("hello, world"),
// 	},
// }
// fs.fstest.ReadDir()

var fileTestBasePath, fileTestRelativePath string

func init() {
	// Path to test files should be hardcoded here for the tests to be valid.
	_, currFileLocation, _, _ := runtime.Caller(0)
	fileTestBasePath = filepath.Dir(currFileLocation)

	fileTestRelativePath = filepath.Join(
		"_test_resources",
		"file_test",
	)
}

func TestGetFilesInRelativePath(t *testing.T) {
	const (
		nonExistingPath = "nonExistingPath"

		testFolderDataSource = "testGetFilesInRelativePath"
		successFileNameJson  = "test.json"
		successFileNameCfg   = "test.cfg"
	)

	testCases := []struct {
		name              string
		inputRelativePath string
		expectedFileCount int
		expectedFileNames []string
		expectedErr       bool
	}{
		{
			name:              "Existing directory: 3 files - 2 successful",
			inputRelativePath: testFolderDataSource,
			expectedFileCount: 2,
			expectedFileNames: []string{
				successFileNameJson,
				successFileNameCfg,
			},
			expectedErr: false,
		},
		{
			name:              "Non-existing directory",
			inputRelativePath: nonExistingPath,
			expectedErr:       true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			relTestPath := filepath.Join(fileTestRelativePath, tc.inputRelativePath)
			result, resultErr := getFilesInRelativePath(relTestPath, fileTestBasePath)

			if len(result) != tc.expectedFileCount {
				t.Fatalf("Test Failed: %v. Expected Result Length: %v Actual Result Length: %v",
					tc.name, tc.expectedFileCount, len(result))
			}

			var resultFileNames []string
			for _, dirEntry := range result {
				resultFileNames = append(resultFileNames, dirEntry.Name())

			}
			if len(resultFileNames) != 0 && reflect.DeepEqual(resultFileNames, tc.expectedFileNames) {
				t.Fatalf("Test Failed: %v. Expected Result Names: %v. Actual Result Name: %v",
					tc.name, tc.expectedFileNames, resultFileNames)
			}

			assertErr := resultErr != nil
			if assertErr != tc.expectedErr {
				t.Fatalf("Test Failed: %v. Expected Error to occur: %v. Returned Error: %v",
					tc.name, tc.expectedErr, resultErr.Error())
			}
		})
	}
}

func TestMustGetBasePath(t *testing.T) {

	const (
		lowerWorkingDir = "_test_resources/"
	)

	testCases := []struct {
		name             string
		inputWorkingDir  string
		expectedBasePath string
	}{
		{
			name:             "BasePath not found. Return root dir '/'",
			inputWorkingDir:  filepath.Dir(fileTestBasePath),
			expectedBasePath: "/",
		},
		{
			name:             "BasePath == Working Directory",
			inputWorkingDir:  filepath.Join(fileTestBasePath, lowerWorkingDir),
			expectedBasePath: fileTestBasePath,
		},
		{
			name:             "BasePath == Working Directory",
			inputWorkingDir:  fileTestBasePath,
			expectedBasePath: fileTestBasePath,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := mustGetBasePath(tc.inputWorkingDir)

			if result != tc.expectedBasePath {
				t.Fatalf("Test Failed: %v. Expected Result: %v Actual Result: %v",
					tc.name, tc.expectedBasePath, result)
			}

		})
	}
}

func TestFilterFiles(t *testing.T) {
	const (
		testFolderDataSource = "testFilterFiles"

		emptyDir         = "empty-dir"
		filteredFilesDir = "filtered-files"

		successFileNameJson = "test.json"
		successFileNameCfg  = "test.cfg"
	)
	testCases := []struct {
		name               string
		inputTestDir       string
		inputTestFileCount int
		expectedFileCount  int
		expectedFileNames  []string
		expectedDirEntry   []fs.DirEntry
	}{
		{
			name:               "Valid files: 6 files - 2 successful - .cfg and .json",
			inputTestDir:       filteredFilesDir,
			inputTestFileCount: 6,
			expectedFileCount:  2,
			expectedFileNames: []string{
				successFileNameJson,
				successFileNameCfg,
			},
			expectedDirEntry: []fs.DirEntry{},
		},
		{
			name:             "empty fs.DirEntry input",
			inputTestDir:     emptyDir,
			expectedDirEntry: []fs.DirEntry{},
		},
		// {
		// 	// Can't test as is. Using test.FS could enable this test
		// 	name: "nil fs.DirEntry input",
		// },
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fullPath := filepath.Join(fileTestBasePath, fileTestRelativePath, testFolderDataSource, tc.inputTestDir)
			dirEntries, err := os.ReadDir(fullPath)
			if err != nil {
				t.Fatalf("Internal Testing error: %v", err)
			}

			if len(dirEntries) != tc.inputTestFileCount {
				t.Fatalf("Intrnal Test Failure: %v. Expected number of loaded Files: %v Actual number of loaded Files: %v",
					tc.name, tc.inputTestFileCount, len(dirEntries))
			}

			result := filterFiles(dirEntries)
			if len(result) != tc.expectedFileCount {
				t.Fatalf("Test Failed: %v. Expected Result Length: %v Actual Result Length: %v",
					tc.name, tc.expectedFileCount, len(result))
			}

			var resultFileNames []string
			for _, dirEntry := range result {
				resultFileNames = append(resultFileNames, dirEntry.Name())

			}
			if len(resultFileNames) != 0 && reflect.DeepEqual(resultFileNames, tc.expectedFileNames) {
				t.Fatalf("Test Failed: %v. Expected Result Names: %v. Actual Result Name: %v",
					tc.name, tc.expectedFileNames, resultFileNames)
			}
		})
	}
}

func TestMustGetFile(t *testing.T) {
	const (
		testFolderDataSource     = "testMustGetFile"
		fooBarFilenameAndContent = "foobar"
	)
	testCases := []struct {
		name             string
		inputRelFilePath string
		inputBasePath    string
		expected         string
		expectedErr      bool
	}{
		{
			name:             "Existing file",
			inputRelFilePath: filepath.Join(fileTestRelativePath, testFolderDataSource, fooBarFilenameAndContent),
			expected:         fooBarFilenameAndContent,
			expectedErr:      false,
		},
		{
			name:        "Non-existing file",
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, resultErr := mustGetFile(fileTestBasePath, tc.inputRelFilePath)

			if strings.Compare(string(result), tc.expected) != 0 {
				t.Fatalf("Test Failed: %v. Expected Result: %v Actual Result: %v",
					tc.name, tc.expected, string(result))
			}

			assertErr := resultErr != nil
			if assertErr != tc.expectedErr {
				t.Fatalf("Test Failed: %v. Expected Error to occur: %v. Returned Error: %v",
					tc.name, tc.expectedErr, resultErr.Error())
			}
		})
	}
}
