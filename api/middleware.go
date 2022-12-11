package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/felipe88alves/sortKeyHttpServer/types"
)

type customHandlerFunc func(w http.ResponseWriter, r *http.Request) *handlerResponse

type handlerResponse struct {
	resp       *types.ResponseUrlStats
	Err        error
	StatusCode int
}

func (e handlerResponse) Error() string {
	return e.Err.Error()
}

func middlewareHandler(f customHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var handlerResp *handlerResponse

		defer func(start time.Time) {
			if handlerResp != nil {
				logHandlerResponse(handlerResp, start)
			}
		}(time.Now())

		handlerResp = f(w, r)

		if handlerResp != nil {
			sendHttpResponse(handlerResp, w)
		}
	}
}

func sendHttpResponse(handlerResp *handlerResponse, w http.ResponseWriter) {
	if handlerResp.Err != nil {
		// errMsg := fmt.Errorf("%v %w", handlerResp.StatusCode, handlerResp.Err)
		writeJson(w, handlerResp.StatusCode, nil)
	} else {
		writeJson(w, handlerResp.StatusCode, handlerResp.resp)
	}
}

func logHandlerResponse(handlerResp *handlerResponse, start time.Time) {
	if handlerResp.Err != nil {
		log.Printf("HTTP Status Code: %d Error: %s Handler took:%v\n",
			handlerResp.StatusCode, handlerResp.Error(), time.Since(start))
	} else {
		log.Printf("HTTP Status Code: %d HTTP Response: %+v Handler took:%v\n",
			handlerResp.StatusCode, *handlerResp.resp, time.Since(start))
	}
}

func writeJson(w http.ResponseWriter, httpStatus int, v any) error {
	writeJsonHeader(w, httpStatus)
	return json.NewEncoder(w).Encode(v)
}

func writeJsonHeader(w http.ResponseWriter, httpStatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
}
