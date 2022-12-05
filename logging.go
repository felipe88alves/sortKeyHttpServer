package main

import (
	"context"
	"fmt"
	"time"
)

type loggingService struct {
	next service
}

func newLoggingService(next service) service {
	return &loggingService{
		next: next,
	}
}

func (lS *loggingService) getUrlStatsData(ctx context.Context) (data *urlStatData, err error) {
	defer func(start time.Time) {
		if data != nil {
			fmt.Printf("Data:%+v Error:%v took:%v\n", data.Data, err, time.Since(start))
		} else {
			fmt.Printf("Error: %v took:%v\n", err, time.Since(start))
		}
	}(time.Now())
	return lS.next.getUrlStatsData(ctx)
}
