package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	urlDataSourceHttp = "http"
	urlDataSourceFile = "file"

	validUrlPrefixProtocolHttp  = "http://"
	validUrlPrefixProtocolHttps = "https://"

	validUrlSuffixFileTypeCfg  = ".cfg"
	validUrlSuffixFileTypeJson = ".json"
)

var (
	defaultHttpDataSource = "config"
	defaultFileDataSource = filepath.Join("resources", "raw-json-files")

	retryAttempts  = 5
	backoffPeriods = []time.Duration{
		1 * time.Second,
		5 * time.Second,
		10 * time.Second,
	}
)

type service interface {
	getUrlStatsData(context.Context) (*urlStatData, error)
}

type urlStatDataService struct {
	dataSourceType string
	dataSourcePath string
}

func newUrlStatDataService(dataSourceType string, dataSourcePath string) (service, error) {
	var err error

	dataSourceType = getDataSourceType(dataSourceType)
	dataSourcePath, err = getDataSource(dataSourceType, dataSourcePath)
	if err != nil {
		return nil, err
	}
	return &urlStatDataService{
		dataSourceType: dataSourceType,
		dataSourcePath: dataSourcePath,
	}, nil
}

func (uS *urlStatDataService) getUrlStatsData(ctx context.Context) (*urlStatData, error) {
	var err error
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	basePath := mustGetBasePath(wd)
	files, err := getFilesInRelativePath(uS.dataSourcePath, basePath)
	if err != nil {
		return nil, err
	} else if len(files) == 0 {
		return nil, fmt.Errorf("no files were loaded from the configured Data Sources")
	}
	switch uS.dataSourceType {
	case urlDataSourceHttp:
		return uS.getUrlStatsDataHttpEndpointsFromFile(ctx, files, basePath)
	case urlDataSourceFile:
		return uS.getUrlStatsDataFromFile(files, basePath)
		// default:
		// 	return nil, fmt.Errorf("invalid method for getting json data. Data Source: %q. Set %q env variable to either %q or %q", method, URLSTATS_ENV_VAR, URLSTATSDATA_URL, URLSTATSDATA_FILE)
	}
	return nil, err
}

func getDataSourceType(dataSourceType string) string {

	switch dataSourceType {
	case urlDataSourceHttp:
		return dataSourceType
	case urlDataSourceFile:
		log.Printf("WARNING: Do not use this setting in produciton. Overriding Data Source to custom value: %v", urlDataSourceFile)
		return dataSourceType
	default:
		// If not defined. Defaults to fetching data from URLs (HTTP Endpoints)
		log.Printf("Error: Invalid Data Source Selected: %v. Using default value: %v", dataSourceType, urlDataSourceHttp)
		return urlDataSourceHttp
	}
}

func getDataSource(dataSourceType, dataSourcePath string) (string, error) {

	if dataSourcePath != "" {
		return dataSourcePath, nil
	}
	switch dataSourceType {
	case urlDataSourceHttp:
		dataSourcePath = defaultHttpDataSource
	case urlDataSourceFile:
		dataSourcePath = defaultFileDataSource
	default:
		return "", fmt.Errorf("unsupported Data Source Type: %s", dataSourceType)
	}
	return dataSourcePath, nil
}

func (uS *urlStatDataService) getUrlStatsDataHttpEndpointsFromFile(ctx context.Context, files []fs.DirEntry, basePath string) (*urlStatData, error) {
	urlStats := new(urlStatData)
	errCount := 0

	for _, file := range files {
		relativeFilePath := filepath.Join(uS.dataSourcePath, file.Name())
		fileName, err := mustGetFile(basePath, relativeFilePath)
		if err != nil {
			return nil, err
		}
		urls := strings.Split(string(fileName), "\n")
		urls, err = validateUrls(urls)
		if err != nil {
			return nil, err
		}

		ch := make(chan interface{})
		var wg sync.WaitGroup

		for _, urlAddr := range urls {
			wg.Add(1)
			go getUrlStatsDataHttp(urlAddr, ch, &wg)
		}

		go func() {
			wg.Wait()
			close(ch)
		}()

		for chReturn := range ch {
			switch r := chReturn.(type) {
			case error:
				log.Printf("Error: %v", r)
				errCount++
			case *urlStatData:
				if r != nil {
					urlStats.Data = append(urlStats.Data, r.Data...)
				}
			default:
				log.Print("Error: HTTP Data Source Endpoint returned an Unsupported Type")
				errCount++
			}
		}
	}
	if len(urlStats.Data) == 0 {
		return nil, fmt.Errorf("all %v http get attempts failed", errCount)
	}
	return urlStats, nil
}

func getUrlStatsDataHttp(urlAddr string, ch chan<- interface{}, wg *sync.WaitGroup) {
	var (
		r       *http.Response
		success bool
		err     error
	)
	defer wg.Done()

retry_loop:
	for _, backoff := range backoffPeriods {
		for i := 0; i < retryAttempts; i++ {
			r, err = http.Get(urlAddr)
			if err != nil {
				log.Printf("Failed to HTTP GET %v. Retrying in %s", urlAddr, backoff)
				time.Sleep(backoff)
				continue
			}
			success = true
			break retry_loop
		}
	}
	if !success {
		ch <- fmt.Errorf("ERROR: Retry limit exceeded. Failed to HTTP GET %v", urlAddr)
		return
	}
	statusOK := r.StatusCode >= 200 && r.StatusCode < 300
	if !statusOK {
		ch <- fmt.Errorf(http.StatusText(r.StatusCode))
		return
	}

	defer r.Body.Close()
	urlStats := new(urlStatData)
	json.NewDecoder(r.Body).Decode(urlStats)
	ch <- urlStats
}

func (uS *urlStatDataService) getUrlStatsDataFromFile(files []fs.DirEntry, basePath string) (*urlStatData, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no valid files were found in the configured Data Source Path. Data Source Type: %v", urlDataSourceFile)
	}
	urlStats := new(urlStatData)

	for _, file := range files {
		relativeFilePath := filepath.Join(uS.dataSourcePath, file.Name())
		fileName, err := mustGetFile(basePath, relativeFilePath)
		if err != nil {
			return nil, err
		}
		urlStatsInstance := urlStatData{}
		if err := json.Unmarshal(fileName, &urlStatsInstance); err != nil {
			log.Printf("Failed to unmarshal json data from file-based source. File: %v Error: %v", relativeFilePath, err)
			// TODO: Investigate: Should we allow the program to continue if one files fails to be loaded?
			continue
		}
		urlStats.Data = append(urlStats.Data, urlStatsInstance.Data...)
	}
	if len(urlStats.Data) == 0 {
		return nil, fmt.Errorf("no valid JSON data was found within the configured Data Source files")
	}

	return urlStats, nil
}

func validateUrls(urls []string) ([]string, error) {
	var validatedUrls []string
	for _, url := range urls {
		if isValidUrlPrefixProtocol(url) && isValidUrlSuffixFileType(url) {
			validatedUrls = append(validatedUrls, url)
		}
	}
	if len(validatedUrls) == 0 {
		return nil, fmt.Errorf("no valid urls were found as data source")
	}
	return validatedUrls, nil
}

func isValidUrlPrefixProtocol(url string) bool {
	return strings.HasPrefix(url, validUrlPrefixProtocolHttp) || strings.HasPrefix(url, validUrlPrefixProtocolHttps)
}

func isValidUrlSuffixFileType(url string) bool {
	return strings.HasSuffix(url, validUrlSuffixFileTypeJson)
}
