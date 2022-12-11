package api

import (
	"context"
	"errors"
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

func (s *apiServer) Start(listenAddr string) error {
	http.HandleFunc("/", middlewareHandler(s.handleRawStats))
	http.HandleFunc(fmt.Sprintf("/%s/", sortkeyPath), middlewareHandler(s.handleSortKey))
	return http.ListenAndServe(listenAddr, nil)
}

func (s *apiServer) handleRawStats(w http.ResponseWriter, r *http.Request) *handlerResponse {
	if r.URL.Path != "/" {
		// Returning nil, since the "/" pattern will always be called
		// A proper Mux would allow for the appropriate BadRequest Status Code Response
		return nil
	}
	urlStats, err := s.svc.getUrlStatsData((context.Background()))
	if err != nil {
		if errStatusCode, errStrconv := strconv.Atoi(err.Error()); errStrconv != nil {
			return &handlerResponse{Err: err, StatusCode: http.StatusInternalServerError}
		} else {
			return &handlerResponse{Err: err, StatusCode: errStatusCode}
		}
	}

	switch r.Method {
	case http.MethodGet:
		jsonReturnMsg := types.ResponseUrlStats{
			SortedUrlStats: &urlStats.Data,
			Count:          len(urlStats.Data),
		}
		return &handlerResponse{resp: &jsonReturnMsg, StatusCode: http.StatusOK}

	default:
		return &handlerResponse{
			Err:        errors.New(http.StatusText(http.StatusMethodNotAllowed)),
			StatusCode: http.StatusMethodNotAllowed}
	}
}

func (s *apiServer) handleSortKey(w http.ResponseWriter, r *http.Request) *handlerResponse {
	urlPathSegments := strings.Split(r.URL.Path, fmt.Sprintf("%s/", sortkeyPath))
	if len(urlPathSegments) == 1 || len(urlPathSegments) > 2 {
		return &handlerResponse{
			Err:        errors.New(http.StatusText(http.StatusBadRequest)),
			StatusCode: http.StatusBadRequest}
	}
	urlPathSegments = strings.Split(urlPathSegments[1], "/")
	if len(urlPathSegments) != 1 || urlPathSegments[0] == "" {
		return &handlerResponse{
			Err:        errors.New(http.StatusText(http.StatusBadRequest)),
			StatusCode: http.StatusBadRequest}
	}

	urlStats, err := s.svc.getUrlStatsData((context.Background()))
	if err != nil {
		if errStatusCode, errStrconv := strconv.Atoi(err.Error()); errStrconv != nil {
			return &handlerResponse{Err: err, StatusCode: http.StatusInternalServerError}
		} else {
			return &handlerResponse{Err: err, StatusCode: errStatusCode}
		}
	}

	switch r.Method {
	case http.MethodGet:
		sortOption := urlPathSegments[0]
		urlStatResponse, err := mergeSort(&urlStats.Data, sortOption)
		if err != nil {
			return &handlerResponse{Err: err, StatusCode: http.StatusInternalServerError}
		}

		urlStatResponse, err = limitReponse(urlStatResponse, r.URL.Query())
		if err != nil {
			return &handlerResponse{Err: err, StatusCode: http.StatusInternalServerError}
		}

		jsonReturnMsg := types.ResponseUrlStats{
			SortedUrlStats: urlStatResponse,
			Count:          len(*urlStatResponse),
		}
		return &handlerResponse{resp: &jsonReturnMsg, StatusCode: http.StatusOK}
	default:
		return &handlerResponse{
			Err:        errors.New(http.StatusText(http.StatusMethodNotAllowed)),
			StatusCode: http.StatusMethodNotAllowed}
	}
}
