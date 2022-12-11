package api

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
	"strconv"
	"testing"
	"time"

	"github.com/felipe88alves/sortKeyHttpServer/types"
	"github.com/felipe88alves/sortKeyHttpServer/utils"
)

var apiTestBasePath, apiTestRelativePath string

func init() {
	// BasePath and RelPath must be hardcoded here for the tests to be valid.
	_, currFileLocation, _, _ := runtime.Caller(0)
	apiTestBasePath = filepath.Dir(filepath.Dir(currFileLocation))

	apiTestRelativePath = filepath.Join(
		"_test_resources",
		"api_test",
	)

	// Overwritting global params to reduce test time
	retryAttempts = 1
	backoffPeriods = []time.Duration{0}
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
			name:               "SortOption: Views",
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
			svc, err := NewUrlStatDataService(testUrlDataSourceFile, relPath)
			if err != nil {
				t.Fatalf("Test Failed: %v Failed to create UrlStatDataService. Error: %v",
					tc.name, err.Error())
			}
			apiServer := NewApiServer(svc)

			handlerResp := apiServer.handleSortKey(rec, req)

			if handlerResp.StatusCode != tc.expectedStatusCode {
				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
					tc.name, tc.expectedStatusCode, handlerResp.StatusCode)
			}

			expectedResponseUrlStats := new(types.ResponseUrlStats)
			if err := json.Unmarshal([]byte(tc.expectedResponse), &expectedResponseUrlStats); err != nil {
				t.Fatalf("test Failed: %v Internal Test Failure: %v",
					tc.name, err.Error())
			}

			assert := reflect.DeepEqual(expectedResponseUrlStats, handlerResp.resp)
			if !assert {
				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
					tc.name, expectedResponseUrlStats.SortedUrlStats, handlerResp.resp)
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
			inputSortOption:    "/" + relevancescoreOption,
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
			name:               "Invalid sortKeyPath: Invalid Sort Option",
			inputSortKeyPath:   sortkeyPath + "/" + unsupportedKeyPath,
			inputSortOption:    "/" + relevancescoreOption,
			inputTestFileDir:   successDir,
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet,
				fmt.Sprintf("/%s%s", tc.inputSortKeyPath, tc.inputSortOption),
				nil)
			rec := httptest.NewRecorder()

			relPath := filepath.Join(apiTestRelativePath, testFolderDataSource, tc.inputTestFileDir)
			svc, err := NewUrlStatDataService(testUrlDataSourceFile, relPath)
			if err != nil {
				t.Fatalf("Test Failed: %v Failed to create UrlStatDataService. Error: %v",
					tc.name, err.Error())
			}
			apiServer := NewApiServer(svc)

			handlerResp := apiServer.handleSortKey(rec, req)

			if handlerResp.StatusCode != tc.expectedStatusCode {
				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
					tc.name, tc.expectedStatusCode, handlerResp.StatusCode)
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
		name                      string
		httpMethod                string
		expectedStatusCode        int
		expectedUrlsReturnedCount int
		expectedCount             int
	}{
		{
			name:                      "httpMethod: GET",
			httpMethod:                http.MethodGet,
			expectedStatusCode:        http.StatusOK,
			expectedUrlsReturnedCount: 5,
			expectedCount:             5,
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

			svc, err := NewUrlStatDataService(testUrlDataSourceFile, testFileDataSource)
			if err != nil {
				t.Fatalf("Test Failed: %v Failed to create UrlStatDataService. Error: %v",
					tc.name, err.Error())
			}
			apiServer := NewApiServer(svc)

			handlerResp := apiServer.handleSortKey(rec, req)

			if handlerResp.StatusCode != tc.expectedStatusCode {
				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
					tc.name, tc.expectedStatusCode, handlerResp.StatusCode)
			}
			if tc.expectedStatusCode >= 200 && tc.expectedStatusCode < 300 {
				if len(*handlerResp.resp.SortedUrlStats) != tc.expectedUrlsReturnedCount {
					t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
						tc.name, tc.expectedUrlsReturnedCount, len(*handlerResp.resp.SortedUrlStats))
				}
				if handlerResp.resp.Count != tc.expectedCount {
					t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
						tc.name, tc.expectedCount, handlerResp.resp.Count)
				}
			} else {
				if handlerResp.Error() != http.StatusText(tc.expectedStatusCode) {
					t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
						tc.name, tc.expectedUrlsReturnedCount, *handlerResp.resp.SortedUrlStats)
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

			svc, err := NewUrlStatDataService(testUrlDataSourceFile, testFileDataSource)
			if err != nil {
				t.Fatalf("test Failed: %v Internal Test Failure: %v",
					tc.name, err.Error())
			}
			apiServer := NewApiServer(svc)

			handlerResp := apiServer.handleSortKey(rec, req)

			if handlerResp.StatusCode != tc.expectedStatusCode {
				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
					tc.name, tc.expectedStatusCode, handlerResp.StatusCode)
			}

			if len(*handlerResp.resp.SortedUrlStats) != tc.expectedUrlsReturnedCount {
				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
					tc.name, tc.expectedUrlsReturnedCount, len(*handlerResp.resp.SortedUrlStats))
			}
			if handlerResp.resp.Count != tc.expectedCount {
				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
					tc.name, tc.expectedCount, handlerResp.resp.Count)
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
	var expectedJsonResponse = &types.ResponseUrlStats{}

	rec := httptest.NewRecorder()

	writeJson(rec, expectedHttpStatusOk, expectedJsonResponse)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("Internal Testing error: %v", err)
		}
	}()

	if resp.StatusCode != expectedHttpStatusOk {
		t.Fatalf("Test Failed. Expected Result: %v Actual Result: %v",
			expectedHttpStatusOk, resp.StatusCode)
	}

	resultContentType := resp.Header.Get("Content-Type")
	if resultContentType != expectedContentType {
		t.Fatalf("Test Failed. Expected Result: %v Actual Result: %v",
			expectedContentType, resultContentType)
	}

	result := new(types.ResponseUrlStats)
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

func TestHandleRawStats(t *testing.T) {
	const (
		responseUnsorted string = `{"data":[{"url":"www.example.com/abc1","views":1000,"relevanceScore":0.5},{"url":"www.example.com/abc5","views":5000,"relevanceScore":0.1},{"url":"www.example.com/abc3","views":3000,"relevanceScore":0.2},{"url":"www.example.com/abc2","views":2000,"relevanceScore":0.4},{"url":"www.example.com/abc4","views":4000,"relevanceScore":0.3}],"count":5}`

		testFolderDataSource = "testHandleRawStats"
		autogeneratedUrlDir  = "autogenerated-url"
		autogeneratedUrlFile = "temp-url-autocreated"

		externalServerInvalidJsonDir  = "invalid-json-format"
		externalServerInvalidJsonFile = "/invalid-json-format.json"

		externalServerValidJsonDir  = "valid-json-format"
		externalServerValidJsonFile = "/valid-json-format.json"

		unsupportedDir = "unsupported"

		urlPathRoot     = "/"
		unsupportedPath = "/unsupported"
	)
	var testUrlDataSourceType = urlDataSourceHttp

	testCases := []struct {
		name                                string
		inputUrlPath                        string
		inputTestExternalServerDir          string
		inputTestExternalServerJsonFilename string
		inputExternalServerReachable        bool
		expectedStatusCode                  int
		expectedResponse                    string
	}{
		{
			name:                                "Valid UrlPath. Valid JSON format.",
			inputUrlPath:                        urlPathRoot,
			inputTestExternalServerDir:          externalServerValidJsonDir,
			inputTestExternalServerJsonFilename: externalServerValidJsonFile,
			inputExternalServerReachable:        true,
			expectedStatusCode:                  http.StatusOK,
			expectedResponse:                    responseUnsorted,
		},
		{
			name:                                "Valid UrlPath. Invalid JSON format.",
			inputUrlPath:                        urlPathRoot,
			inputTestExternalServerDir:          externalServerInvalidJsonDir,
			inputTestExternalServerJsonFilename: externalServerInvalidJsonFile,
			inputExternalServerReachable:        true,
			expectedStatusCode:                  http.StatusInternalServerError,
			expectedResponse:                    "",
		},
		{
			name:                                "Unsupported UrlPath. Valid JSON format.",
			inputUrlPath:                        unsupportedPath,
			inputTestExternalServerDir:          unsupportedDir,
			inputTestExternalServerJsonFilename: externalServerValidJsonFile,
			inputExternalServerReachable:        true,
			expectedStatusCode:                  http.StatusBadRequest,
			expectedResponse:                    "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			relPathExternalServer := filepath.Join(apiTestRelativePath, testFolderDataSource, tc.inputTestExternalServerDir)
			relPathSortKeyServer := filepath.Join(apiTestRelativePath, testFolderDataSource, autogeneratedUrlDir+"-"+tc.inputTestExternalServerDir)

			fullPathExternalServer := filepath.Join(utils.BasePath, relPathExternalServer)
			fullPathSortKeyServer := filepath.Join(utils.BasePath, relPathSortKeyServer)

			if err := os.RemoveAll(fullPathSortKeyServer); err != nil {
				t.Fatalf("Internal Testing error: %v", err)
			}
			if err := os.MkdirAll(fullPathSortKeyServer, 0755); err != nil {
				t.Fatalf("Internal Testing error: %v", err)
			}
			defer func() {
				if err := os.RemoveAll(fullPathSortKeyServer); err != nil {
					t.Fatalf("Internal Testing error: %v", err)
				}
			}()

			// Create SortKey Mock Server to retrieve Sorted URL Statistics Data
			sortKeyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

			// Create an External Mock Server with URL Statistics in JSON format
			filePath := filepath.Join(fullPathExternalServer, tc.inputTestExternalServerJsonFilename)
			externalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(tc.inputTestExternalServerJsonFilename))
				w.Header().Set("Content-Type", "application/octet-stream")
				http.ServeFile(w, r, filePath)
			}))

			filenameSortKeyServer := autogeneratedUrlFile + fileTypeBySource[testUrlDataSourceType]
			fullFilePathSortKeyServer := filepath.Join(fullPathSortKeyServer, filenameSortKeyServer)
			file, err := os.Create(fullFilePathSortKeyServer)
			if err != nil {
				t.Fatalf("Internal Testing error: %v", err)
			}

			defer func() {
				if err := file.Close(); err != nil {
					t.Fatalf("Internal Testing error: %v", err)
				}
			}()
			defer func() {
				if err := os.RemoveAll(fullFilePathSortKeyServer); err != nil {
					t.Fatalf("Internal Testing error: %v", err)
				}
			}()

			// Write the HTTP Endpoint/address of the External Mock Server into the .cfg file
			fileContent := externalServer.URL + tc.inputTestExternalServerJsonFilename
			if _, err := file.WriteString(fileContent); err != nil {
				t.Fatalf("Internal Testing error: %v", err)
			}

			// Create a Request to the SortKey Server
			reqURL := sortKeyServer.URL + tc.inputUrlPath
			req := httptest.NewRequest(http.MethodGet,
				reqURL,
				nil)
			rec := httptest.NewRecorder()

			// Begin test
			svc, err := NewUrlStatDataService(testUrlDataSourceType, relPathSortKeyServer)
			if err != nil {
				t.Fatalf("Test Failed: %v Failed to create UrlStatDataService. Error: %v",
					tc.name, err.Error())
			}
			apiServer := NewApiServer(svc)

			handlerResp := apiServer.handleRawStats(rec, req)

			if tc.inputTestExternalServerDir == unsupportedDir {
				if handlerResp != nil {
					t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
						tc.name, "nil", handlerResp.resp)
				}
			} else {
				if handlerResp.StatusCode != tc.expectedStatusCode {
					t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
						tc.name, tc.expectedStatusCode, handlerResp.StatusCode)
				}

				if tc.expectedResponse == "" {
					if handlerResp.resp != nil {
						t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
							tc.name, "nil", handlerResp.resp)
					}
					// } else if  {
				} else {
					expectedResponseUrlStats := new(types.ResponseUrlStats)
					if err := json.Unmarshal([]byte(tc.expectedResponse), &expectedResponseUrlStats); err != nil {
						t.Fatalf("test Failed: %v Internal Test Failure: %v",
							tc.name, err.Error())
					}

					assert := reflect.DeepEqual(expectedResponseUrlStats, handlerResp.resp)
					if !assert {
						t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
							tc.name, expectedResponseUrlStats, handlerResp.resp)
					}
				}
			}
		})
	}
}
