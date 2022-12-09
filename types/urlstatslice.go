package types

import "fmt"

type UrlStatSlice []*UrlStat

func (u *UrlStatSlice) String() string {
	var s []UrlStat
	for _, d := range *u {
		s = append(s, *d)
	}
	return fmt.Sprintf("%+v", s)
}
