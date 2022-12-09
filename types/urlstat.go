package types

type UrlStat struct {
	Url            string  `json:"url,omitempty"`
	Views          int     `json:"views,omitempty"`
	RelevanceScore float32 `json:"relevanceScore,omitempty"`
}
