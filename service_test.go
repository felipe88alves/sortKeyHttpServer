package main

import (
	"reflect"
	"testing"
)

const unsupportedValue = "UNSUPPORTED"

func TestNewUrlStatDataService(t *testing.T) {

	testCases := []struct {
		name                string
		inputDataSourceType string
		inputDataSourcePath string
		expected            Service
	}{
		{
			name:                "inputDataSourceType empty - inputDataSourcePath set to FILE Path - Use default HTTP value for inputDataSourceType",
			inputDataSourcePath: defaultFileDataSource,
			expected: &urlStatDataService{
				dataSourceType: urlDataSourceHttp,
				dataSourcePath: defaultFileDataSource,
			},
		},
		{
			name:                "inputDataSourceType empty - inputDataSourcePath set - Use set values",
			inputDataSourcePath: defaultHttpDataSource,
			expected: &urlStatDataService{
				dataSourceType: urlDataSourceHttp,
				dataSourcePath: defaultHttpDataSource,
			},
		},
		{
			name:                "inputDataSourceType set to HTTP - inputDataSourcePath empty - Use default values",
			inputDataSourceType: urlDataSourceHttp,
			expected: &urlStatDataService{
				dataSourceType: urlDataSourceHttp,
				dataSourcePath: defaultHttpDataSource,
			},
		},
		{
			name:                "inputDataSourceType set to File - inputDataSourcePath empty - Use default values",
			inputDataSourceType: urlDataSourceFile,
			expected: &urlStatDataService{
				dataSourceType: urlDataSourceFile,
				dataSourcePath: defaultFileDataSource,
			},
		},
		{
			name: "Empty inputs - Use default values",
			expected: &urlStatDataService{
				dataSourceType: urlDataSourceHttp,
				dataSourcePath: defaultHttpDataSource,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := newUrlStatDataService(tc.inputDataSourceType, tc.inputDataSourcePath)
			assert := reflect.DeepEqual(result, tc.expected)

			if !assert {
				t.Fatalf("Test Failed: %v. Expected Result: %v Actual Result: %v",
					tc.name, tc.expected, result)
			}
		})
	}
}

func TestGetUrlStatsDataFromFile(t *testing.T) {

}

// func TestgetUrlStatsData(t *testing.T) {

// 	testCases := []struct {
// 		name         string
// 		inputData    *urlStatDataService
// 		expectedData *UrlStatData
// 		expectedErr  error
// 	}{
// 		{
// 			name: "Empty inputs - Use default values",
// 			expected: &urlStatDataService{
// 				dataSourceType: urlDataSourceHttp,
// 				dataSourcePath: defaultHttpDataSource,
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		t.Run(tc.name, func(t *testing.T) {
// 			result := getUrlStatsData(tc.inputDataSourceType, tc.inputDataSourcePath)
// 			assert := reflect.DeepEqual(result, tc.expected)

// 			if !assert {
// 				t.Fatalf("Test Failed: %v. Expected Result: %v Actual Result: %v",
// 					tc.name, tc.expected, result)
// 			}
// 		})
// 	}
// }

// func TestGetUrlStatsData(t *testing.T) {
// 	testCases := []struct {
// 		name             string
// 		input            string
// 		expectedUrlStats *UrlStatData
// 		expectedErr      error
// 	}{
// 		{
// 			name:  fmt.Sprintf("Data Source %q", defaultFileDataSource),
// 			input: defaultFileDataSource,
// 		},
// 		// {
// 		// 	name:  fmt.Sprintf("Data Source %q", urldatasource_http),
// 		// 	input: urldatasource_http,
// 		// },
// 	}
// 	for _, tc := range testCases {
// 		tc := tc
// 		t.Run(tc.name, func(t *testing.T) {

// 			var urlStats = new(urlStatDataService)
// 			_, errResult := urlStats.getUrlStatsData(context.TODO())

// 			if errResult != tc.expectedErr {
// 				t.Fatalf("Test Failed: %v. Expected Result: %v Actual Result: %v",
// 					tc.name, tc.expectedErr, errResult)
// 			}
// 		})
// 	}
// }
