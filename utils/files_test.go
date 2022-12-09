package utils

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
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

var fileTestRelativePath string

func init() {
	// BasePath and RelPath must be hardcoded here for the tests to be valid.
	fileTestRelativePath = filepath.Join(
		"_test_resources",
		"file_test",
	)
}

func TestGetFilesInRelativePathByType(t *testing.T) {
	const (
		success         = "success"
		emptyDir        = "empty-dir"
		noValidFileType = "invalid-files"
		nonExistingPath = "nonExistingPath"

		testFolderDataSource = "testGetFilesInRelativePath"
		successFileNameJson  = "test.json"
		successFileNameCfg   = "test.cfg"

		jsonFileType = ".json"
		cfgFileType  = ".cfg"
	)

	testCases := []struct {
		name              string
		inputTestDir      string
		inputTypeFilter   string
		expectedFileNames []string
		expectedErr       bool
	}{
		{
			name:            fmt.Sprintf("Filter: %s Total/Filtered Files: 3/1", jsonFileType),
			inputTestDir:    success,
			inputTypeFilter: jsonFileType,
			expectedFileNames: []string{
				successFileNameJson,
			},
			expectedErr: false,
		},
		{
			name:            fmt.Sprintf("Filter: %s Total/Filtered Files: 3/1", cfgFileType),
			inputTestDir:    success,
			inputTypeFilter: cfgFileType,
			expectedFileNames: []string{
				successFileNameJson,
			},
			expectedErr: false,
		},
		{
			name:            fmt.Sprintf("Filter: %s Total/Returned Files: 4/0", jsonFileType),
			inputTestDir:    noValidFileType,
			inputTypeFilter: jsonFileType,
			expectedErr:     true,
		},
		{
			name:            fmt.Sprintf("Filter: %s Total/Returned Files: 4/0", cfgFileType),
			inputTestDir:    noValidFileType,
			inputTypeFilter: cfgFileType,
			expectedErr:     true,
		},
		{
			name:         "Empty directory",
			inputTestDir: emptyDir,
			expectedErr:  true,
		},
		{
			name:         "Non-existing directory",
			inputTestDir: nonExistingPath,
			expectedErr:  true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			relTestPath := filepath.Join(fileTestRelativePath, testFolderDataSource, tc.inputTestDir)
			fullPath := filepath.Join(BasePath, relTestPath)

			if tc.inputTestDir == emptyDir {
				if err := os.RemoveAll(fullPath); err != nil {
					t.Fatalf("Internal Testing error: %v", err)
				}
				if err := os.Mkdir(fullPath, 0755); err != nil {
					t.Fatalf("Internal Testing error: %v", err)
				}
				defer func() {
					if err := os.Remove(fullPath); err != nil {
						t.Fatalf("Internal Testing error: %v", err)
					}
				}()
			}

			result, resultErr := GetFilesInRelativePathByType(relTestPath, tc.inputTypeFilter)

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

	dirAboveBasePath, _ := mustGetBasePath()
	dirAboveBasePath = filepath.Dir(dirAboveBasePath)

	testCases := []struct {
		name             string
		inputWorkingDir  string
		expectedBasePath string
		expectedErr      bool
	}{
		{
			name:             "BasePath != Working Directory - Higher",
			inputWorkingDir:  dirAboveBasePath,
			expectedBasePath: "/",
		},
		{
			name:             "BasePath != Working Directory - Lower",
			inputWorkingDir:  filepath.Join(BasePath, fileTestRelativePath),
			expectedBasePath: BasePath,
		},
		{
			name:             "BasePath == Working Directory",
			inputWorkingDir:  BasePath,
			expectedBasePath: BasePath,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := os.Chdir(tc.inputWorkingDir); err != nil {
				t.Fatalf("Internal Testing error: %v", err)
			}

			result, resultErr := mustGetBasePath()

			if result != tc.expectedBasePath {
				t.Fatalf("Test Failed: %v. Expected Result: %v Actual Result: %v",
					tc.name, tc.expectedBasePath, result)
			}

			assertErr := resultErr != nil
			if assertErr != tc.expectedErr {
				t.Fatalf("Test Failed: %v. Expected Error to occur: %v. Returned Error: %v",
					tc.name, tc.expectedErr, resultErr.Error())
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

		jsonFileType = ".json"
		cfgFileType  = ".cfg"
	)
	testCases := []struct {
		name               string
		inputTypeFilter    string
		inputTestDir       string
		inputTestFileCount int
		expectedFileNames  []string
		expectedDirEntry   []fs.DirEntry
	}{
		{
			name:               "Valid files: 6 files - 1 successful - .json",
			inputTypeFilter:    jsonFileType,
			inputTestDir:       filteredFilesDir,
			inputTestFileCount: 6,
			expectedFileNames: []string{
				successFileNameJson,
			},
			expectedDirEntry: []fs.DirEntry{},
		},
		{
			name:               "Valid files: 6 files - 1 successful - .cfg",
			inputTypeFilter:    cfgFileType,
			inputTestDir:       filteredFilesDir,
			inputTestFileCount: 6,
			expectedFileNames: []string{
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
			fullPath := filepath.Join(BasePath, fileTestRelativePath, testFolderDataSource, tc.inputTestDir)
			if tc.inputTestDir == emptyDir {
				if err := os.RemoveAll(fullPath); err != nil {
					t.Fatalf("Internal Testing error: %v", err)
				}
				if err := os.Mkdir(fullPath, 0755); err != nil {
					t.Fatalf("Internal Testing error: %v", err)
				}
				defer func() {
					if err := os.Remove(fullPath); err != nil {
						t.Fatalf("Internal Testing error: %v", err)
					}
				}()
			}
			dirEntries, err := os.ReadDir(fullPath)
			if err != nil {
				t.Fatalf("Internal Testing error: %v", err)
			}

			if len(dirEntries) != tc.inputTestFileCount {
				t.Fatalf("Internal Test Failure: %v. Expected number of loaded Files: %v Actual number of loaded Files: %v",
					tc.name, tc.inputTestFileCount, len(dirEntries))
			}

			result := filterFiles(dirEntries, tc.inputTypeFilter)

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
			result, resultErr := MustGetFile(tc.inputRelFilePath)

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
