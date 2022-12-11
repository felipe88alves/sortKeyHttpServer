package api

import (
	"context"
	"log"
	"time"

	"github.com/felipe88alves/sortKeyHttpServer/types"
)

type loggingService struct {
	next service
}

func NewLoggingService(next service) service {
	return &loggingService{
		next: next,
	}
}

func (l *loggingService) getUrlStatsData(ctx context.Context) (data *types.UrlStatData, err error) {
	defer func(start time.Time) {
		if data != nil {
			log.Printf("Input Data:%+v Error:%v took:%v\n", data.Data.String(), err, time.Since(start))
		} else {
			log.Printf("Error: %v took:%v\n", err, time.Since(start))
		}
	}(time.Now())
	return l.next.getUrlStatsData(ctx)
}
