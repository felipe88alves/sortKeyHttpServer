package types

type ResponseUrlStats struct {
	SortedUrlStats *UrlStatSlice `json:"data"`
	Count          int           `json:"count"`
}
