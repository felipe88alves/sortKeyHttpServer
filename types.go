package main

import "fmt"

type urlStatData struct {
	Data urlStatSlice `json:"data,omitempty"`
}

type urlStatSlice []*urlStat

func (u urlStatSlice) String() string {
	var s []urlStat
	for _, d := range u {
		s = append(s, *d)
	}
	return fmt.Sprintf("%+v", s)
}

type urlStat struct {
	Url            string  `json:"url,omitempty"`
	Views          int     `json:"views,omitempty"`
	RelevanceScore float32 `json:"relevanceScore,omitempty"`
}

type responseUrlStats struct {
	SortedUrlStats *urlStatSlice `json:"data"`
	Count          int           `json:"count"`
}
