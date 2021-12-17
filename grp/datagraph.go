package grp

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/xcontrol"
	"github.com/ruraomsk/gena/sdb"
)

type HeadGraph struct {
	Region  int         `json:"region"`
	Area    int         `json:"area"`
	Subarea int         `json:"subarea"`
	Date    string      `json:"date"`
	Step    int         `json:"step"`
	Graph   []GraphLine `json:"graph"`
	Counts  int         `json:"count"`
}
type GraphLine struct {
	Start int `json:"start"`
	Pr    int `json:"pr"`
	Ob    int `json:"ob"`
}

func MakeStat(path string) error {
	logger.Info.Printf("working %s", path)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var hg HeadGraph
	err = json.Unmarshal(b, &hg)
	if err != nil {
		return err
	}
	logger.Info.Printf("Регион %d Район %d Подрайон %d", hg.Region, hg.Area, hg.Subarea)
	xctrl, err := sdb.GetXctrl(hg.Region, hg.Area, hg.Subarea)
	if err != nil {
		return err
	}
	cts := make([]xcontrol.Calculates, 0)
	for _, x := range xctrl.Xctrls {
		cts = append(cts, x.Calculates...)
	}
	for _, ct := range cts {
		ast := new(pudge.ArchStat)
		ast.Region = ct.Region
		ast.Area = ct.Area
		ast.ID = ct.ID
		ast.Statistics = make([]pudge.Statistic, 0)
		s := hg.Step
		period := 1
		for s <= 60*24 {
			sl := new(pudge.Statistic)
			sl.TLen = hg.Step
			sl.Period = period
			sl.Hour = s / 60
			sl.Min = s % 60
			sl.Type = 1
			sl.Datas = make([]pudge.DataStat, 0)
			for i := 1; i <= hg.Counts; i++ {
				sl.Datas = append(sl.Datas, pudge.DataStat{Chanel: i, Status: 0, Intensiv: 0})
			}
		}

	}
	return nil
}
