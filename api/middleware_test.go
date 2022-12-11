package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/felipe88alves/sortKeyHttpServer/types"
)

var middlewareTestBasePath, middlewareTestRelativePath string

func init() {
	// BasePath and RelPath must be hardcoded here for the tests to be valid.
	_, currFileLocation, _, _ := runtime.Caller(0)
	middlewareTestBasePath = filepath.Dir(filepath.Dir(currFileLocation))

	middlewareTestRelativePath = filepath.Join(
		"_test_resources",
		"api_test",
	)

	// Overwritting global params to reduce test time
	retryAttempts = 1
	backoffPeriods = []time.Duration{0}
}

func (s *apiServer) stubHandlerResponseSuccess(w http.ResponseWriter, r *http.Request) *handlerResponse {
	succStatusCode := http.StatusOK
	return &handlerResponse{StatusCode: succStatusCode, resp: &types.ResponseUrlStats{}}
}
func (s *apiServer) stubHandlerResponseError(w http.ResponseWriter, r *http.Request) *handlerResponse {
	err := errors.New("Stub Handler Response Error")
	errStatusCode := http.StatusBadRequest
	return &handlerResponse{Err: err, StatusCode: errStatusCode}
}
func TestMiddlewareHandler(t *testing.T) {
	const (
		maxAttempts = 5
		inputPort   = ":6060"

		testApiErrorPath = "/apiError/"
		testSuccessPath  = "/success/"
	)
	var (
		resp     *http.Response
		httpConn bool
		// err      error
	)
	testCases := []struct {
		name                  string
		inputApiErrorReturned bool
		inputTestUrlPath      string
		expectedStatusCode    int
		expectedErr           string
	}{
		{
			name:                  "Error: true",
			inputApiErrorReturned: true,
			inputTestUrlPath:      testApiErrorPath,
			expectedStatusCode:    http.StatusBadRequest,
			expectedErr:           "Stub Handler apiError",
		},
		{
			name:                  "Error: false",
			inputApiErrorReturned: false,
			inputTestUrlPath:      testSuccessPath,
			expectedStatusCode:    http.StatusOK,
		},
	}

	svc := new(urlStatDataService)
	apiSvc := NewApiServer(svc)

	http.HandleFunc(testApiErrorPath, middlewareHandler(apiSvc.stubHandlerResponseError))
	http.HandleFunc(testSuccessPath, middlewareHandler(apiSvc.stubHandlerResponseSuccess))

	go func() {
		log.Fatal(http.ListenAndServe(inputPort, nil))
	}()

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			for attempt := 1; attempt <= maxAttempts; attempt++ {
				req, err := http.NewRequest("GET",
					fmt.Sprintf("http://localhost%s%s", inputPort, tc.inputTestUrlPath),
					nil)
				if err != nil {
					t.Fatalf("Internal Testing error: %v", err)
				}
				req.Close = true
				client := &http.Client{}
				resp, err = client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					httpConn = true
					break
				}

				time.Sleep(time.Second)
				t.Logf("Retrying connection to Mock Server (%d/%d)", attempt, maxAttempts)
			}
			if !httpConn {
				t.Fatalf("Failed to connect to Mock Server")
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Fatalf("Internal Testing error: %v", err)
				}
			}()

			apiResp := new(types.ResponseUrlStats)
			if err := json.NewDecoder(resp.Body).Decode(apiResp); err != nil {
				t.Fatalf("Internal Testing error: %v", err)
			}

			if resp.StatusCode != tc.expectedStatusCode {
				t.Fatalf("Test Failed. Expected Result: %v Actual Result: %v",
					tc.expectedStatusCode, resp.StatusCode)
			}
		})
	}
}
