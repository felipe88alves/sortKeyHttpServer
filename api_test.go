package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func testHandleSortKey_sortOption(t *testing.T) {
	const (
		UNSUPPORTED_SORT_OPTION        = "unsupported"
		RESPONSE_SORTED_RELEVANCESCORE = `{"data":[{"url":"www.example.com/abc5","views":5000,"relevanceScore":0.1},{"url":"www.example.com/abc3","views":3000,"relevanceScore":0.2},{"url":"www.example.com/abc4","views":4000,"relevanceScore":0.3},{"url":"www.example.com/abc2","views":2000,"relevanceScore":0.4},{"url":"www.example.com/abc1","views":1000,"relevanceScore":0.5}],"count":5}`
		RESPONSE_SORTED_VIEWS          = `{"data":[{"url":"www.example.com/abc1","views":1000,"relevanceScore":0.5},{"url":"www.example.com/abc2","views":2000,"relevanceScore":0.4},{"url":"www.example.com/abc3","views":3000,"relevanceScore":0.2},{"url":"www.example.com/abc4","views":4000,"relevanceScore":0.3},{"url":"www.example.com/abc5","views":5000,"relevanceScore":0.1}],"count":5}`
	)
	testCases := []struct {
		name               string
		sortOption         string
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name:               "SortOption: relevanceScore",
			sortOption:         relevancescoreOption,
			expectedStatusCode: http.StatusOK,
			expectedResponse:   RESPONSE_SORTED_RELEVANCESCORE,
		},
		{
			name:               "SortOption: relevanceScore",
			sortOption:         viewsOption,
			expectedStatusCode: http.StatusOK,
			expectedResponse:   RESPONSE_SORTED_VIEWS,
		},
		// {
		// 	name:               "SortOption: unsupported - defaults to relevanceScore",
		// 	sortOption:         UNSUPPORTED_SORT_OPTION,
		// 	expectedStatusCode: http.StatusOK,
		// 	expectedResponse:   RESPONSE_SORTED_RELEVANCESCORE,
		// },
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel()
			req := httptest.NewRequest(http.MethodGet,
				fmt.Sprintf("/%s/%s", sortkeyPath, tc.sortOption),
				nil)
			rec := httptest.NewRecorder()

			dataSourceType := urlDataSourceFile
			dataSourcePath := filepath.Join("test_resources", "raw-json-files", "success")

			svc := newUrlStatDataService(dataSourceType, dataSourcePath)
			apiServer := NewApiServer(svc)
			_ = apiServer.handleSortKey(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatusCode {
				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
					tc.name, tc.expectedStatusCode, resp.StatusCode)
			}

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Internal Test Failure: %v", err)
			}

			if string(respBody) != tc.expectedResponse {
				t.Fatalf("Test Failed: %v\nExpected Result: %v\nActual Result: %v",
					tc.name, tc.expectedResponse, string(respBody))
			}
		})
	}
}

// func TestHandleSortKey_sortKeyPath(t *testing.T) {
// 	const UNSUPPORTED_SORT_KEY_PATH = "unsupported"
// 	testCases := []struct {
// 		name               string
// 		sortKeyPath        string
// 		expectedStatusCode int
// 	}{
// 		{
// 			name:               "sortKeyPath: sortkey",
// 			sortKeyPath:        Sortkey_path,
// 			expectedStatusCode: http.StatusOK,
// 		},
// 		// {
// 		// 	name:               "sortKeyPath: unsupported",
// 		// 	sortKeyPath:        UNSUPPORTED_SORT_KEY_PATH,
// 		// 	expectedStatusCode: http.StatusBadRequest,
// 		// },
// 		{
// 			name:               "sortKeyPath: duplicated sortkey",
// 			sortKeyPath:        Sortkey_path + "/" + UNSUPPORTED_SORT_KEY_PATH + "/" + Sortkey_path,
// 			expectedStatusCode: http.StatusBadRequest,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		t.Run(tc.name, func(t *testing.T) {
// 			// t.Parallel()
// 			req := httptest.NewRequest(http.MethodGet,
// 				fmt.Sprintf("/%s/%s", tc.sortKeyPath, relevancescore_option_path),
// 				nil)
// 			rec := httptest.NewRecorder()

// 			HandleSortKey(rec, req)

// 			resp := rec.Result()
// 			defer resp.Body.Close()
// 			result := resp.StatusCode

// 			if result != tc.expectedStatusCode {
// 				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
// 					tc.name, tc.expectedStatusCode, result)
// 			}
// 		})
// 	}
// }

// func TestHandleSortKey_httpMethod(t *testing.T) {
// 	const UNSUPPORTED_HTTP_METHOD = "unsupported"
// 	testCases := []struct {
// 		name               string
// 		httpMethod         string
// 		expectedStatusCode int
// 	}{
// 		{
// 			name:               "httpMethod: GET",
// 			httpMethod:         http.MethodGet,
// 			expectedStatusCode: http.StatusOK,
// 		},
// 		{
// 			name:               "httpMethod: PUT",
// 			httpMethod:         http.MethodPut,
// 			expectedStatusCode: http.StatusMethodNotAllowed,
// 		},
// 		{
// 			name:               "httpMethod: unsupported",
// 			httpMethod:         UNSUPPORTED_HTTP_METHOD,
// 			expectedStatusCode: http.StatusMethodNotAllowed,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		t.Run(tc.name, func(t *testing.T) {
// 			// t.Parallel()
// 			req := httptest.NewRequest(tc.httpMethod,
// 				fmt.Sprintf("/%s/%s", Sortkey_path, relevancescore_option_path),
// 				nil)
// 			rec := httptest.NewRecorder()

// 			HandleSortKey(rec, req)

// 			resp := rec.Result()
// 			defer resp.Body.Close()
// 			result := resp.StatusCode

// 			if result != tc.expectedStatusCode {
// 				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
// 					tc.name, tc.expectedStatusCode, result)
// 			}
// 		})
// 	}
// }

// func TestHandleSortKey_limitFilter(t *testing.T) {
// 	testCases := []struct {
// 		name                      string
// 		limitFilter               string
// 		expectedStatusCode        int
// 		expectedUrlsReturnedCount int
// 		expectedCount             int
// 	}{
// 		{
// 			name:                      "limit filter within range - limit return",
// 			limitFilter:               "1",
// 			expectedStatusCode:        http.StatusOK,
// 			expectedUrlsReturnedCount: 1,
// 			expectedCount:             1,
// 		},
// 		{
// 			name:                      "limit filter larger than available - return all",
// 			limitFilter:               fmt.Sprint(math.MaxInt),
// 			expectedStatusCode:        http.StatusOK,
// 			expectedUrlsReturnedCount: 5,
// 			expectedCount:             5,
// 		},
// 		{
// 			name:                      "limit filter < 0 - Invalid filter, return all",
// 			limitFilter:               "-1",
// 			expectedStatusCode:        http.StatusOK,
// 			expectedUrlsReturnedCount: 5,
// 			expectedCount:             5,
// 		},
// 		{
// 			name:                      "limit filter == 0 - Invalid filter, return all",
// 			limitFilter:               "0",
// 			expectedStatusCode:        http.StatusOK,
// 			expectedUrlsReturnedCount: 5,
// 			expectedCount:             5,
// 		},
// 		{
// 			name:                      "emtpy limit filter - No filter, return all",
// 			expectedStatusCode:        http.StatusOK,
// 			expectedUrlsReturnedCount: 5,
// 			expectedCount:             5,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		t.Run(tc.name, func(t *testing.T) {
// 			// t.Parallel()
// 			req := httptest.NewRequest(http.MethodGet,
// 				fmt.Sprintf("/%s/%s?%s=%s", Sortkey_path, relevancescore_option_path, limit_filter_path, tc.limitFilter),
// 				nil)
// 			rec := httptest.NewRecorder()

// 			HandleSortKey(rec, req)
// 			resp := rec.Result()
// 			defer resp.Body.Close()

// 			if resp.StatusCode != tc.expectedStatusCode {
// 				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
// 					tc.name, tc.expectedStatusCode, resp.StatusCode)
// 			}

// 			testResponseUrlStats := new(responseUrlStats)
// 			json.NewDecoder(resp.Body).Decode(testResponseUrlStats)
// 			if len(testResponseUrlStats.SortedUrlStats) != tc.expectedUrlsReturnedCount {
// 				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
// 					tc.name, tc.expectedUrlsReturnedCount, len(testResponseUrlStats.SortedUrlStats))
// 			}
// 			if testResponseUrlStats.Count != tc.expectedCount {
// 				t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
// 					tc.name, tc.expectedCount, testResponseUrlStats.Count)
// 			}
// 		})
// 	}
// }

// func TestNegative_File_WrongJsonFormat(t *testing.T) {
// 	// // t.Parallel()
// 	jsonFilesRelativePath = filepath.Join("test_resources", "raw-json-files", "fail-unmarshal")
// 	name := "Negative Test case: Wrong JSON format"

// 	req := httptest.NewRequest(http.MethodGet,
// 		fmt.Sprintf("/%s/%s", Sortkey_path, relevancescore_option_path),
// 		nil)
// 	rec := httptest.NewRecorder()

// 	HandleSortKey(rec, req)

// 	resp := rec.Result()
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusInternalServerError {
// 		t.Fatalf("Test Failed: %v Expected Result: %v Actual Result: %v",
// 			name, http.StatusInternalServerError, resp.StatusCode)
// 	}
// }
