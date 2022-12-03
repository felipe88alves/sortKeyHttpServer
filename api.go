package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	sortkeyPath = "sortkey"

	relevancescoreOption = "relevanceScore"
	viewsOption          = "views"
)

type ApiServer struct {
	svc Service
}

func NewApiServer(svc Service) *ApiServer {
	return &ApiServer{
		svc: svc,
	}
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func makeHttpHandler(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			if e, ok := err.(apiError); ok {
				writeJson(w, e.Status, e)
				return
			}
			writeJson(w, http.StatusInternalServerError, apiError{Err: err.Error(), Status: http.StatusInternalServerError})
		}
	}
}

func (s *ApiServer) Start(listenAddr string) error {
	http.HandleFunc("/", makeHttpHandler(s.handleStats))
	http.HandleFunc(fmt.Sprintf("/%s/", sortkeyPath), makeHttpHandler(s.handleSortKey))
	return http.ListenAndServe(listenAddr, nil)
}

func (s *ApiServer) handleStats(w http.ResponseWriter, r *http.Request) error {
	if r.URL.Path != "/" {
		return apiError{Err: http.StatusText(http.StatusBadRequest), Status: http.StatusBadRequest}
	}
	urlStats, err := s.svc.getUrlStatsData((context.Background()))
	if err != nil {
		if errStatusCode, errStrconv := strconv.Atoi(err.Error()); errStrconv != nil {
			return apiError{Err: err.Error(), Status: http.StatusInternalServerError}
		} else {
			return apiError{Err: err.Error(), Status: errStatusCode}
		}
	}

	switch r.Method {
	case http.MethodGet:
		jsonReturnMsg := responseUrlStats{
			SortedUrlStats: urlStats.Data,
			Count:          len(urlStats.Data),
		}
		writeJson(w, http.StatusOK, jsonReturnMsg)

		return nil

	default:
		return apiError{Err: http.StatusText(http.StatusMethodNotAllowed), Status: http.StatusMethodNotAllowed}
	}
}

func (s *ApiServer) handleSortKey(w http.ResponseWriter, r *http.Request) error {
	urlPathSegments := strings.Split(r.URL.Path, fmt.Sprintf("%s/", sortkeyPath))
	if len(urlPathSegments) == 1 || len(urlPathSegments) > 2 {
		return apiError{Err: http.StatusText(http.StatusBadRequest), Status: http.StatusBadRequest}
	}
	urlPathSegments = strings.Split(urlPathSegments[1], "/")
	if len(urlPathSegments) != 1 || urlPathSegments[0] == "" {
		return apiError{Err: http.StatusText(http.StatusBadRequest), Status: http.StatusBadRequest}
	}

	urlStats, err := s.svc.getUrlStatsData((context.Background()))
	if err != nil {
		if errStatusCode, errStrconv := strconv.Atoi(err.Error()); errStrconv != nil {
			return apiError{Err: err.Error(), Status: http.StatusInternalServerError}
		} else {
			return apiError{Err: err.Error(), Status: errStatusCode}
		}
	}

	switch r.Method {
	case http.MethodGet:
		sortOption := urlPathSegments[0]
		urlStatResponse, err := mergeSort(urlStats.Data, sortOption)
		if err != nil {
			log.Printf("error performing merge sort algorithm. Sort Option: %v, Error: %v", sortOption, err)
			return apiError{Err: err.Error(), Status: http.StatusInternalServerError}
		}

		limitValue := getLimitValue(r.URL.Query())
		urlStatResponse = limitReponse(urlStatResponse, limitValue)

		jsonReturnMsg := responseUrlStats{
			SortedUrlStats: urlStatResponse,
			Count:          len(urlStatResponse),
		}
		writeJson(w, http.StatusOK, jsonReturnMsg)
		return nil
	default:
		return apiError{Err: http.StatusText(http.StatusMethodNotAllowed), Status: http.StatusMethodNotAllowed}
	}
}

func writeJson(w http.ResponseWriter, httpStatus int, v any) error {
	w.WriteHeader(httpStatus)
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func mergeSort(items []urlStat, sortBy string) ([]urlStat, error) {
	sortBy = getSortOption(sortBy)

	if len(items) <= 1 {
		return items, nil
	}
	first, err := mergeSort(items[:len(items)/2], sortBy)
	if err != nil {
		return nil, err
	}
	second, err := mergeSort(items[len(items)/2:], sortBy)
	// The error handling below unecessary wih current implementation. Leaving it anyway for robustness.
	// mergeSort error is caused by sortOption, so it would have already been returned in the first call to mergeSort
	// This comment block can be removed if additional error handling is added to mergeSort
	if err != nil {
		return nil, err
	}
	return merge(sortBy, first, second)
}

func merge(sortBy string, first, last []urlStat) ([]urlStat, error) {
	final := []urlStat{}
	i := 0
	j := 0
	for i < len(first) && j < len(last) {
		isSorted, err := isSortedByOption(sortBy, first[i], last[j])
		if err != nil {
			return nil, err
		}

		if isSorted {
			final = append(final, first[i])
			i++
		} else {
			final = append(final, last[j])
			j++
		}
	}

	for ; i < len(first); i++ {
		final = append(final, first[i])
	}
	for ; j < len(last); j++ {
		final = append(final, last[j])
	}
	return final, nil

}

func isSortedByOption(sortByOption string, first, last urlStat) (bool, error) {
	const ERROR_INVALID_SORT_OPTION = "invalid sort option selected"

	// Do equals need to be considered? Any logic to solve ties
	switch sortByOption {
	case relevancescoreOption:
		return isSortedByRelevanceScore(first, last), nil
	case viewsOption:
		return isSortedByViewsScore(first, last), nil
	default:
		return false, fmt.Errorf("%v %v", ERROR_INVALID_SORT_OPTION, sortByOption)
	}
}

func isSortedByRelevanceScore(first, last urlStat) bool {
	return first.RelevanceScore < last.RelevanceScore
}

func isSortedByViewsScore(first, last urlStat) bool {
	return first.Views < last.Views
}
func getSortOption(sortOption string) string {
	if sortOption == "" ||
		(sortOption != relevancescoreOption && sortOption != viewsOption) {
		sortOption = relevancescoreOption
	}
	return sortOption
}
func getLimitValue(limitValueSegment url.Values) int {
	limitFilter := "limit"
	limitValue, err := strconv.Atoi(limitValueSegment.Get(limitFilter))
	if err != nil || limitValue <= 0 {
		limitValue = -1
	}
	return limitValue
}

func limitReponse(urlStatSlice []urlStat, limit int) []urlStat {
	if limit <= 0 || limit > len(urlStatSlice) {
		return urlStatSlice
	}
	return urlStatSlice[:limit]
}
