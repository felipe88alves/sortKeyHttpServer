package main

import (
	"context"
	"fmt"
	"time"
)

type LoggingService struct {
	next Service
}

func NewLoggingService(next Service) Service {
	return &LoggingService{
		next: next,
	}
}

func (lS *LoggingService) getUrlStatsData(ctx context.Context) (data *UrlStatData, err error) {
	defer func(start time.Time) {
		if data != nil {
			fmt.Printf("Data:%+v Error:%v took:%v\n", data.Data, err, time.Since(start))
		} else {
			fmt.Printf("Error: %v took:%v\n", err, time.Since(start))
		}
	}(time.Now())
	return lS.next.getUrlStatsData(ctx)
}
