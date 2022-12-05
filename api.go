package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	sortkeyPath = "sortkey"

	relevancescoreOption = "relevanceScore"
	viewsOption          = "views"
	limitFilterOption    = "limit"
)

type apiServer struct {
	svc service
}

func newApiServer(svc service) *apiServer {
	return &apiServer{
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

func (s *apiServer) start(listenAddr string) error {
	http.HandleFunc("/", makeHttpHandler(s.handleRawStats))
	http.HandleFunc(fmt.Sprintf("/%s/", sortkeyPath), makeHttpHandler(s.handleSortKey))
	return http.ListenAndServe(listenAddr, nil)
}

func (s *apiServer) handleRawStats(w http.ResponseWriter, r *http.Request) error {
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
			SortedUrlStats: &urlStats.Data,
			Count:          len(urlStats.Data),
		}
		writeJson(w, http.StatusOK, jsonReturnMsg)

		return nil

	default:
		return apiError{Err: http.StatusText(http.StatusMethodNotAllowed), Status: http.StatusMethodNotAllowed}
	}
}

func (s *apiServer) handleSortKey(w http.ResponseWriter, r *http.Request) error {
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
		urlStatResponse, err := mergeSort(&urlStats.Data, sortOption)
		if err != nil {
			return apiError{Err: err.Error(), Status: http.StatusInternalServerError}
		}

		limitValue := getLimitValue(r.URL.Query())
		urlStatResponse, err = limitReponse(urlStatResponse, limitValue)
		if err != nil {
			return apiError{Err: err.Error(), Status: http.StatusInternalServerError}
		}

		jsonReturnMsg := responseUrlStats{
			SortedUrlStats: urlStatResponse,
			Count:          len(*urlStatResponse),
		}
		writeJson(w, http.StatusOK, jsonReturnMsg)
		return nil
	default:
		return apiError{Err: http.StatusText(http.StatusMethodNotAllowed), Status: http.StatusMethodNotAllowed}
	}
}

func writeJson(w http.ResponseWriter, httpStatus int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	return json.NewEncoder(w).Encode(v)
}

func mergeSort(items *urlStatSlice, sortBy string) (*urlStatSlice, error) {
	if items == nil {
		return nil, fmt.Errorf("null pointer exception. Found when sorting Url Data")
	}
	sortBy = getSortOption(sortBy)

	if len(*items) <= 1 {
		return items, nil
	}
	first := (*items)[:len(*items)/2]
	firstPtr, err := mergeSort(&first, sortBy)
	if err != nil {
		return nil, err
	}
	second := (*items)[len(*items)/2:]
	secondPtr, err := mergeSort(&second, sortBy)
	// The error handling below is unecessary wih current implementation. LDecision was made to leave it in for robustness.
	// DETAILS ON DECISION: mergeSort error is caused by sortOption, so it would have already been returned in the first call to mergeSort.
	// This comment block can be removed if additional error handling is added to mergeSort
	if err != nil {
		return nil, err
	}
	return merge(sortBy, firstPtr, secondPtr)
}

func merge(sortBy string, first, last *urlStatSlice) (*urlStatSlice, error) {
	final := new(urlStatSlice)
	i := 0
	j := 0
	for i < len(*first) && j < len(*last) {
		isSorted, err := isSortedByOption(sortBy, (*first)[i], (*last)[j])
		if err != nil {
			return nil, err
		}

		if isSorted {
			*final = append(*final, (*first)[i])
			i++
		} else {
			*final = append(*final, (*last)[j])
			j++
		}
	}

	for ; i < len(*first); i++ {
		*final = append(*final, (*first)[i])
	}
	for ; j < len(*last); j++ {
		*final = append(*final, (*last)[j])
	}
	return final, nil

}

func isSortedByOption(sortByOption string, first, last *urlStat) (bool, error) {
	const errInvalidSortOption = "invalid sort option selected"

	// Do equals need to be considered? Any logic to solve ties
	switch sortByOption {
	case relevancescoreOption:
		return isSortedByRelevanceScoreAscending(first, last)
	case viewsOption:
		return isSortedByViewsScoreAscending(first, last)
	default:
		return false, fmt.Errorf("%v %v", errInvalidSortOption, sortByOption)
	}
}

func isSortedByRelevanceScoreAscending(first, last *urlStat) (bool, error) {
	if first == nil || last == nil {
		return false, fmt.Errorf("null pointer exception. Found when sorting Views in ascending order")
	}
	// WARNING: Empty RelevanceScore is treated as normal 0 value
	return first.RelevanceScore < last.RelevanceScore, nil
}

func isSortedByViewsScoreAscending(first, last *urlStat) (bool, error) {
	if first == nil || last == nil {
		return false, fmt.Errorf("null pointer exception. Found when sorting Views in ascending order")
	}
	// WARNING: Empty Views are treated as normal 0 value
	return first.Views < last.Views, nil
}

func getSortOption(sortOption string) string {
	if sortOption == "" ||
		(sortOption != relevancescoreOption && sortOption != viewsOption) {
		sortOption = relevancescoreOption
	}
	return sortOption
}
func getLimitValue(limitValueSegment url.Values) int {
	limitValue, err := strconv.Atoi(limitValueSegment.Get(limitFilterOption))
	if err != nil || limitValue <= 0 {
		limitValue = -1
	}
	return limitValue
}

func limitReponse(u *urlStatSlice, limit int) (*urlStatSlice, error) {
	if u == nil {
		return nil, fmt.Errorf("null pointer exception. Found when filtering response using Limit Option")
	}
	if limit <= 0 || limit > len(*u) {
		return u, nil
	}
	*u = (*u)[:limit]
	return u, nil
}
