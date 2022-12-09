package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/felipe88alves/sortKeyHttpServer/types"
)

const (
	sortkeyPath = "sortkey"
)

type apiServer struct {
	svc service
}

func NewApiServer(svc service) *apiServer {
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

func (s *apiServer) Start(listenAddr string) error {
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
		jsonReturnMsg := types.ResponseUrlStats{
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

		urlStatResponse, err = limitReponse(urlStatResponse, r.URL.Query())
		if err != nil {
			return apiError{Err: err.Error(), Status: http.StatusInternalServerError}
		}

		jsonReturnMsg := types.ResponseUrlStats{
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
