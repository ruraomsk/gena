package grp

import (
	"encoding/json"
	"io/ioutil"
	"sort"
	"time"

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
	sort.Slice(hg.Graph, func(i, j int) bool {
		return hg.Graph[i].Start < hg.Graph[j].Start
	})
	if hg.Graph[0].Start != 0 {
		//Должна быть всегда запись с времени 0
		hg.Graph = append(hg.Graph, GraphLine{Start: 0, Pr: hg.Graph[0].Pr, Ob: hg.Graph[0].Ob})
		sort.Slice(hg.Graph, func(i, j int) bool {
			return hg.Graph[i].Start < hg.Graph[j].Start
		})
	}
	if hg.Graph[len(hg.Graph)-1].Start != 24*60 {
		//Должна быть всегда запись с времени 24:00
		hg.Graph = append(hg.Graph, GraphLine{Start: 24 * 60, Pr: hg.Graph[len(hg.Graph)-1].Pr, Ob: hg.Graph[len(hg.Graph)-1].Ob})
		sort.Slice(hg.Graph, func(i, j int) bool {
			return hg.Graph[i].Start < hg.Graph[j].Start
		})
	}
	table := make(map[int]GraphLine)
	s := hg.Step
	for s <= 60*24 {
		//Вычисляем новые значения
		p := 60 / hg.Step
		l := hg.Graph[0]
		flag := false
		for _, g := range hg.Graph {
			if g.Start == s {
				table[s] = GraphLine{Start: s, Pr: g.Pr / p, Ob: g.Ob / p}
				// logger.Debug.Printf("%v", table[s])
				break
			}
			if s > g.Start {
				l = g
				flag = true
				continue
			}
			if s < g.Start && flag {
				// (x-l.Srart)/(g.Start-l.Start)=(y-g.Pr)/(g.Pr-l.Pr)
				// (x-l.Srart)/(g.Start-l.Start)=(y-g.Ob)/(g.Ob-l.Ob)
				pr := (((g.Pr - l.Pr) * (s - l.Start)) / (g.Start - l.Start)) + l.Pr
				ob := (((g.Ob - l.Ob) * (s - l.Start)) / (g.Start - l.Start)) + l.Ob
				table[s] = GraphLine{Start: s, Pr: pr / p, Ob: ob / p}
				// logger.Debug.Printf("%v", table[s])
				break
			}

		}
		s += hg.Step
	}

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
		ast.Date, _ = time.Parse("2006-01-02", hg.Date)
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
			count := 0
			for _, n := range ct.ChanL {
				if n == 0 {
					continue
				}
				if n > count {
					count = n
				}
			}
			for _, n := range ct.ChanR {
				if n == 0 {
					continue
				}
				if n > count {
					count = n
				}
			}

			for i := 1; i <= count; i++ {
				sl.Datas = append(sl.Datas, pudge.DataStat{Chanel: i, Status: 0, Intensiv: 0})
			}
			p := len(ct.ChanL)
			if p != 0 {
				for _, n := range ct.ChanL {
					if n == 0 {
						continue
					}
					sl.Datas[n-1].Intensiv = table[s].Pr / p
				}
			}
			p = len(ct.ChanR)
			if p != 0 {
				for _, n := range ct.ChanR {
					if n == 0 {
						continue
					}
					sl.Datas[n-1].Intensiv = table[s].Ob / p
				}
			}
			ast.Statistics = append(ast.Statistics, *sl)
			s += hg.Step
		}
		sdb.WriteStat(ast)
		logger.Info.Printf("Записана статистика для %d %d %d ", ast.Region, ast.Area, ast.ID)
	}
	return nil
}
