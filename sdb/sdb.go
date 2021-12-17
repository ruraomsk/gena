package sdb

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/xcontrol"
	"github.com/ruraomsk/gena/setup"
)

var ConDB *sql.DB
var err error

func InitDataBase() error {
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	ConDB, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		return err
	}
	return nil
}

func GetXctrl(region, area, subarea int) (*xcontrol.State, error) {
	xctrl := new(xcontrol.State)
	w := fmt.Sprintf("select state from public.xctrl where region=%d and area=%d and subarea=%d;", region, area, subarea)
	rows, err := ConDB.Query(w)
	if err != nil {
		return nil, err
	}
	found := false
	for rows.Next() {
		var s []byte
		err = rows.Scan(&s)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal(s, &xctrl)
		found = true
	}
	rows.Close()
	if found {
		return xctrl, nil
	}
	return nil, fmt.Errorf("нет такого XCTRL")
}

func WriteStat(rs *pudge.ArchStat) error {
	w := fmt.Sprintf("select count(*) from public.statistics where date='%s' and region=%d and area=%d and id=%d;",
		rs.Date.Format("2006-01-02"), rs.Region, rs.Area, rs.ID)
	rows, err := ConDB.Query(w)
	if err != nil {
		return err
	}
	var count int
	for rows.Next() {
		rows.Scan(&count)
	}
	rows.Close()
	js, _ := json.Marshal(&rs)
	if count == 0 {
		w = fmt.Sprintf("INSERT INTO public.statistics(region, area, id, date, stat) VALUES (%d, %d, %d, '%s', '%s');",
			rs.Region, rs.Area, rs.ID, rs.Date.Format("2006-01-02"), string(js))
	} else {
		w = fmt.Sprintf("Update public.statistics set stat='%s' where date='%s' and region=%d and area=%d and id=%d;",
			string(js), rs.Date.Format("2006-01-02"), rs.Region, rs.Area, rs.ID)

	}
	_, err = ConDB.Exec(w)
	return err
}
