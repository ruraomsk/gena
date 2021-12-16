package graph

type HeadGraph struct {
	Region  int         `json:"region"`
	Area    int         `json:"area"`
	Subarea int         `json:"subarea"`
	Date    string      `json:"date"`
	Graph   []GraphLine `json:"graph"`
}
type GraphLine struct {
	Start int `json:"start"`
	Pr    int `json:"pr"`
	Ob    int `json:"ob"`
}
