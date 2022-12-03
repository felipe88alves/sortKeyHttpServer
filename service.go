package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	urlDataSourceHttp = "http"
	urlDataSourceFile = "file"
)

var (
	defaultHttpDataSource = "config"
	defaultFileDataSource = filepath.Join("resources", "raw-json-files")
)

type Service interface {
	getUrlStatsData(context.Context) (*UrlStatData, error)
}

type urlStatDataService struct {
	dataSourceType string
	dataSourcePath string
}

func newUrlStatDataService(dataSourceType string, dataSourcePath string) Service {
	dataSourceType = getDataSourceType(dataSourceType)
	dataSourcePath = getDataSource(dataSourceType, dataSourcePath)
	return &urlStatDataService{
		dataSourceType: dataSourceType,
		dataSourcePath: dataSourcePath,
	}
}

func (uS *urlStatDataService) getUrlStatsData(ctx context.Context) (*UrlStatData, error) {
	var err error
	basePath, err := mustGetBasePath()
	if err != nil {
		return nil, err
	}
	files, err := getFilesInRelativePath(uS.dataSourcePath, basePath)
	if err != nil {
		return nil, err
	} else if files == nil {
		return nil, fmt.Errorf("no URL Endpoint was loaded from configured Data Sources")
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
		log.Printf("Error: Invalid Data Source Selected: %v. Using default value: %v", dataSourceType, urlDataSourceHttp)
		return urlDataSourceHttp
	}
	// If not defined. Defaults to fetching data from URLs (HTTP Endpoints)

}

func getDataSource(dataSourceType string, dataSourcePath string) string {

	if dataSourcePath != "" {
		return dataSourcePath
	}
	switch dataSourceType {
	case urlDataSourceHttp:
		dataSourcePath = defaultHttpDataSource
	case urlDataSourceFile:
		dataSourcePath = defaultFileDataSource
	default:
		panic("Unsupported Data Source Type")
	}
	return dataSourcePath
}

func (uS *urlStatDataService) getUrlStatsDataHttpEndpointsFromFile(ctx context.Context, files []fs.DirEntry, basePath string) (*UrlStatData, error) {
	urlStats := new(UrlStatData)
	errCount := 0

	for _, file := range files {
		fullPath := filepath.Join(uS.dataSourcePath, file.Name())
		fileName, err := mustGetFile(fullPath, basePath)
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
			case *UrlStatData:
				if r != nil {
					urlStats.Data = append(urlStats.Data, r.Data...)
				}
			default:
				log.Print("HTTP Data Source Endpoint returned an Unsupported type")
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

		retryAttempts  = 5
		backoffPeriods = []time.Duration{
			1 * time.Second,
			// 5 * time.Second,
			// 10 * time.Second,
		}
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
	urlStats := new(UrlStatData)
	json.NewDecoder(r.Body).Decode(urlStats)
	ch <- urlStats
}

func (uS *urlStatDataService) getUrlStatsDataFromFile(files []fs.DirEntry, basePath string) (*UrlStatData, error) {
	var (
		urlStats = new(UrlStatData)
	)
	for _, file := range files {
		fullPath := filepath.Join(uS.dataSourcePath, file.Name())
		fileName, err := mustGetFile(fullPath, basePath)
		if err != nil {
			return nil, err
		}
		urlStatsInstance := UrlStatData{}
		if err := json.Unmarshal(fileName, &urlStatsInstance); err != nil {
			log.Printf("Failed to unmarshal json data from file-based source. File: %v Error: %v", fullPath, err)
			// TODO: Investigate: Should we allow the program to continue if one files fails to be loaded?
			continue
		}
		urlStats.Data = append(urlStats.Data, urlStatsInstance.Data...)
	}
	return urlStats, nil
}

func validateUrls(urls []string) ([]string, error) {
	var validatedUrls []string
	for _, url := range urls {
		if strings.HasPrefix(url, "http") && strings.HasSuffix(url, "json") {
			validatedUrls = append(validatedUrls, url)
		}
	}
	if len(validatedUrls) == 0 {
		return nil, fmt.Errorf("no valid urls were found as data source")
	}
	return validatedUrls, nil
}
