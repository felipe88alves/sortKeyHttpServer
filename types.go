package main

type UrlStatData struct {
	Data []urlStat `json:"data,omitempty"`
}
type urlStat struct {
	Url            string  `json:"url,omitempty"`
	Views          int     `json:"views,omitempty"`
	RelevanceScore float32 `json:"relevanceScore,omitempty"`
}

type responseUrlStats struct {
	SortedUrlStats []urlStat `json:"data"`
	Count          int       `json:"count"`
}
