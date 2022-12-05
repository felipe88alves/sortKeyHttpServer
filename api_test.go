package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

var apiTestBasePath, apiTestRelativePath string

func init() {
	// Path to test files should be hardcoded here for the tests to be valid.
	_, currFileLocation, _, _ := runtime.Caller(0)
	apiTestRelativePath = filepath.Join(
		"_test_resources",
		"api_test",
	)
	apiTestBasePath = filepath.Dir(currFileLocation)
}

func TestHandleRawStats(t *testing.T) {
	const (
		unsupported = "unsupported"
	)
	var (
		responseUnsorted = `{"data":[{"url":"www.example.com/abc1","views":1000,"relevanceScore":0.5},{"url":"www.example.com/abc5","views":5000,"relevanceScore":0.1},{"url":"www.example.com/abc3","views":3000,"relevanceScore":0.2},{"url":"www.example.com/abc2","views":2000,"relevanceScore":0.4},{"url":"www.example.com/abc4","views":4000,"relevanceScore":0.3}],"count":5}`

		testUrlDataSourceFile = urlDataSourceFile
		testFolderDataSource  = "testHandleRawStats"

		successDir  = "success"
		emptyDir    = "empty-dir"
		gitKeepFile = ".gitkeep"

		urlPathRoot = "/"
	)
	testCases := []struct {
		name               string
		inputUrlPath       string
		inputTestFileDir   string
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name:               "Valid UrlPath. GET UrlStats: Fail ",
			inputUrlPath:       urlPathRoot,
			inputTestFileDir:   emptyDir,
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "{}",
		},
		{
			name:               "Valid UrlPath. GET UrlStats: Success ",
			inputUrlPath:       urlPathRoot,
			inputTestFileDir:   successDir,
			expectedStatusCode: http.StatusOK,
			expectedResponse:   responseUnsorted,
		},
		{
			name:               "Invalid UrlPath",
			inputUrlPath:       urlPathRoot + unsupported,
			inputTestFileDir:   successDir,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{}",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet,
				tc.inputUrlPath,
				nil)
			rec := httptest.NewRecorder()

			relPath := filepath.Join(apiTestRelativePath, testFolderDataSource, tc.inputTestFileDir)

			if tc.inputTestFileDir == emptyDir {
				// Remove .gitkeep file and create it again once the test is finished
				fullFilePath := filepath.Join(serviceTestBasePath, relPath, gitKeepFile)
				os.Remove(fullFilePath)
				defer os.Create(fullFilePath)
			}

			svc, err := newUrlStatDataService(testUrlDataSourceFile, relPath)
			if err != nil {
				t.Fatalf("Test Failed: %v Failed to create UrlStatDataService. Error: %v",
					tc.name, err.Error())
			}
			svc = newLoggingService(svc)
			apiServer := newApiServer(svc)

			if resultStatusCode, ok := apiServer.handleRawStats(rec, req).(apiError); ok {
				if resultStatusCode.Status != tc.expectedStatusCode {
					t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
						tc.name, tc.expectedStatusCode, resultStatusCode)
				}
			} else {
				if http.StatusOK != tc.expectedStatusCode {
					t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
						tc.name, tc.expectedStatusCode, http.StatusOK)
				}
			}

			resp := rec.Result()
			defer resp.Body.Close()

			resultResponseUrlStats := new(responseUrlStats)
			json.NewDecoder(resp.Body).Decode(resultResponseUrlStats)

			expectedResponseUrlStats := new(responseUrlStats)
			if err := json.Unmarshal([]byte(tc.expectedResponse), &expectedResponseUrlStats); err != nil {
				t.Fatalf("test Failed: %v Internal Test Failure: %v",
					tc.name, err.Error())
			}

			assert := reflect.DeepEqual(expectedResponseUrlStats, resultResponseUrlStats)
			if !assert {
				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
					tc.name, expectedResponseUrlStats, resultResponseUrlStats)
			}
		})
	}
}

func TestHandleSortKey_sortOption(t *testing.T) {
	const (
		unsupported = "unsupported"
	)
	var (
		responseSortedRelevancescore = `{"data":[{"url":"www.example.com/abc5","views":5000,"relevanceScore":0.1},{"url":"www.example.com/abc3","views":3000,"relevanceScore":0.2},{"url":"www.example.com/abc4","views":4000,"relevanceScore":0.3},{"url":"www.example.com/abc2","views":2000,"relevanceScore":0.4},{"url":"www.example.com/abc1","views":1000,"relevanceScore":0.5}],"count":5}`
		responseSortedViews          = `{"data":[{"url":"www.example.com/abc1","views":1000,"relevanceScore":0.5},{"url":"www.example.com/abc2","views":2000,"relevanceScore":0.4},{"url":"www.example.com/abc3","views":3000,"relevanceScore":0.2},{"url":"www.example.com/abc4","views":4000,"relevanceScore":0.3},{"url":"www.example.com/abc5","views":5000,"relevanceScore":0.1}],"count":5}`

		testUrlDataSourceFile = urlDataSourceFile
		testFolderDataSource  = "testHandleSortKey"

		successDir = "success"
	)
	testCases := []struct {
		name               string
		inputSortOption    string
		inputTestFileDir   string
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name:               "SortOption: relevanceScore",
			inputSortOption:    relevancescoreOption,
			inputTestFileDir:   successDir,
			expectedStatusCode: http.StatusOK,
			expectedResponse:   responseSortedRelevancescore,
		},
		{
			name:               "SortOption: relevanceScore",
			inputSortOption:    viewsOption,
			inputTestFileDir:   successDir,
			expectedStatusCode: http.StatusOK,
			expectedResponse:   responseSortedViews,
		},
		{
			name:               "SortOption: unsupported - defaults to relevanceScore",
			inputSortOption:    unsupported,
			inputTestFileDir:   successDir,
			expectedStatusCode: http.StatusOK,
			expectedResponse:   responseSortedRelevancescore,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet,
				fmt.Sprintf("/%s/%s", sortkeyPath, tc.inputSortOption),
				nil)
			rec := httptest.NewRecorder()

			relPath := filepath.Join(apiTestRelativePath, testFolderDataSource, tc.inputTestFileDir)
			svc, err := newUrlStatDataService(testUrlDataSourceFile, relPath)
			if err != nil {
				t.Fatalf("Test Failed: %v Failed to create UrlStatDataService. Error: %v",
					tc.name, err.Error())
			}
			svc = newLoggingService(svc)
			apiServer := newApiServer(svc)

			if resultStatusCode, ok := apiServer.handleSortKey(rec, req).(apiError); ok {
				if resultStatusCode.Status != tc.expectedStatusCode {
					t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
						tc.name, tc.expectedStatusCode, resultStatusCode)
				}
			} else {
				if http.StatusOK != tc.expectedStatusCode {
					t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
						tc.name, tc.expectedStatusCode, http.StatusOK)
				}
			}

			resp := rec.Result()
			defer resp.Body.Close()

			resultResponseUrlStats := new(responseUrlStats)
			json.NewDecoder(resp.Body).Decode(resultResponseUrlStats)

			expectedResponseUrlStats := new(responseUrlStats)
			if err := json.Unmarshal([]byte(tc.expectedResponse), &expectedResponseUrlStats); err != nil {
				t.Fatalf("test Failed: %v Internal Test Failure: %v",
					tc.name, err.Error())
			}

			assert := reflect.DeepEqual(expectedResponseUrlStats, resultResponseUrlStats)
			if !assert {
				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
					tc.name, expectedResponseUrlStats.SortedUrlStats, resultResponseUrlStats.SortedUrlStats)
			}
		})
	}
}

func TestHandleSortKey_sortKeyPath(t *testing.T) {
	const unsupportedKeyPath = "unsupported"

	var (
		testUrlDataSourceFile = urlDataSourceFile
		testFolderDataSource  = "testHandleSortKey"

		successDir = "success"
	)

	testCases := []struct {
		name               string
		inputSortKeyPath   string
		inputSortOption    string
		inputTestFileDir   string
		expectedStatusCode int
	}{
		{
			name:               "sortKeyPath: sortkey",
			inputSortKeyPath:   sortkeyPath,
			inputSortOption:    relevancescoreOption,
			inputTestFileDir:   successDir,
			expectedStatusCode: http.StatusOK,
		},

		{
			name:               "Invalid sortKeyPath: No Sort Option declared",
			inputSortKeyPath:   sortkeyPath,
			inputTestFileDir:   successDir,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "Invalid sortKeyPath: duplicated sortkey",
			inputSortKeyPath:   sortkeyPath + "/" + unsupportedKeyPath + "/" + sortkeyPath,
			inputSortOption:    relevancescoreOption,
			inputTestFileDir:   successDir,
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet,
				fmt.Sprintf("/%s/%s", tc.inputSortKeyPath, tc.inputSortOption),
				nil)
			rec := httptest.NewRecorder()

			relPath := filepath.Join(apiTestRelativePath, testFolderDataSource, tc.inputTestFileDir)
			svc, err := newUrlStatDataService(testUrlDataSourceFile, relPath)
			if err != nil {
				t.Fatalf("Test Failed: %v Failed to create UrlStatDataService. Error: %v",
					tc.name, err.Error())
			}
			svc = newLoggingService(svc)
			apiServer := newApiServer(svc)

			if resultStatusCode, ok := apiServer.handleSortKey(rec, req).(apiError); ok {
				if resultStatusCode.Status != tc.expectedStatusCode {
					t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
						tc.name, tc.expectedStatusCode, resultStatusCode)
				}
			} else {
				if http.StatusOK != tc.expectedStatusCode {
					t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
						tc.name, tc.expectedStatusCode, http.StatusOK)
				}
			}
		})
	}
}

func TestHandleSortKey_httpMethod(t *testing.T) {
	var (
		testUrlDataSourceFile = urlDataSourceFile
		testFileDataSource    = filepath.Join("_test_resources", "api_test", "testHandler")
	)
	const unsupportedHttMethod = "unsupported"
	testCases := []struct {
		name               string
		httpMethod         string
		expectedStatusCode int
	}{
		{
			name:               "httpMethod: GET",
			httpMethod:         http.MethodGet,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "httpMethod: PUT",
			httpMethod:         http.MethodPut,
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "httpMethod: unsupported",
			httpMethod:         unsupportedHttMethod,
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tc.httpMethod,
				fmt.Sprintf("/%s/%s", sortkeyPath, relevancescoreOption),
				nil)
			rec := httptest.NewRecorder()

			svc, err := newUrlStatDataService(testUrlDataSourceFile, testFileDataSource)
			if err != nil {
				t.Fatalf("Test Failed: %v Failed to create UrlStatDataService. Error: %v",
					tc.name, err.Error())
			}
			svc = newLoggingService(svc)
			apiServer := newApiServer(svc)

			if resultStatusCode, ok := apiServer.handleSortKey(rec, req).(apiError); ok {
				if resultStatusCode.Status != tc.expectedStatusCode {
					t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
						tc.name, tc.expectedStatusCode, resultStatusCode)
				}
			} else {
				if http.StatusOK != tc.expectedStatusCode {
					t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
						tc.name, tc.expectedStatusCode, http.StatusOK)
				}
			}
		})
	}
}

func TestHandleSortKey_limitFilter(t *testing.T) {
	var (
		testUrlDataSourceFile = urlDataSourceFile
		testFileDataSource    = filepath.Join("_test_resources", "api_test", "testHandler")
	)

	testCases := []struct {
		name                      string
		limitFilter               string
		expectedStatusCode        int
		expectedUrlsReturnedCount int
		expectedCount             int
	}{
		{
			name:                      "limit filter within range - limit return",
			limitFilter:               "1",
			expectedStatusCode:        http.StatusOK,
			expectedUrlsReturnedCount: 1,
			expectedCount:             1,
		},
		{
			name:                      "limit filter larger than available - return all",
			limitFilter:               fmt.Sprint(math.MaxInt),
			expectedStatusCode:        http.StatusOK,
			expectedUrlsReturnedCount: 5,
			expectedCount:             5,
		},
		{
			name:                      "limit filter < 0 - Invalid filter, return all",
			limitFilter:               "-1",
			expectedStatusCode:        http.StatusOK,
			expectedUrlsReturnedCount: 5,
			expectedCount:             5,
		},
		{
			name:                      "limit filter == 0 - Invalid filter, return all",
			limitFilter:               "0",
			expectedStatusCode:        http.StatusOK,
			expectedUrlsReturnedCount: 5,
			expectedCount:             5,
		},
		{
			name:                      "emtpy limit filter - No filter, return all",
			expectedStatusCode:        http.StatusOK,
			expectedUrlsReturnedCount: 5,
			expectedCount:             5,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet,
				fmt.Sprintf("/%s/%s?%s=%s", sortkeyPath, relevancescoreOption, limitFilterOption, tc.limitFilter),
				nil)
			rec := httptest.NewRecorder()

			svc, err := newUrlStatDataService(testUrlDataSourceFile, testFileDataSource)
			if err != nil {
				t.Fatalf("test Failed: %v Internal Test Failure: %v",
					tc.name, err.Error())
			}
			svc = newLoggingService(svc)
			apiServer := newApiServer(svc)

			if resultStatusCode, ok := apiServer.handleSortKey(rec, req).(apiError); ok {
				if resultStatusCode.Status != tc.expectedStatusCode {
					t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
						tc.name, tc.expectedStatusCode, resultStatusCode)
				}
			} else {
				if http.StatusOK != tc.expectedStatusCode {
					t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
						tc.name, tc.expectedStatusCode, http.StatusOK)
				}
			}

			resp := rec.Result()
			defer resp.Body.Close()

			testResponseUrlStats := new(responseUrlStats)
			json.NewDecoder(resp.Body).Decode(testResponseUrlStats)
			if len(*testResponseUrlStats.SortedUrlStats) != tc.expectedUrlsReturnedCount {
				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
					tc.name, tc.expectedUrlsReturnedCount, len(*testResponseUrlStats.SortedUrlStats))
			}
			if testResponseUrlStats.Count != tc.expectedCount {
				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
					tc.name, tc.expectedCount, testResponseUrlStats.Count)
			}
		})
	}
}

func TestWriteJson(t *testing.T) {
	t.Parallel()
	// TODO: missing negative test case. Error when Encoding Json
	const (
		expectedHttpStatusOk = http.StatusOK
		expectedContentType  = "application/json"
	)
	var expectedJsonResponse = &responseUrlStats{}

	rec := httptest.NewRecorder()

	writeJson(rec, expectedHttpStatusOk, expectedJsonResponse)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != expectedHttpStatusOk {
		t.Fatalf("Test Failed. Expected Result: %v Actual Result: %v",
			expectedHttpStatusOk, resp.StatusCode)
	}

	resultContentType := resp.Header.Get("Content-Type")
	if resultContentType != expectedContentType {
		t.Fatalf("Test Failed. Expected Result: %v Actual Result: %v",
			expectedContentType, resultContentType)
	}

	result := new(responseUrlStats)
	err := json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		t.Fatalf("test Failed: %v Internal Test Failure: %v",
			"TestWriteJson", err.Error())
	}

	assert := reflect.DeepEqual(result, expectedJsonResponse)
	if !assert {
		t.Fatalf("Test Failed. Expected Result: %v Actual Result: %v",
			result, expectedJsonResponse)
	}
}
