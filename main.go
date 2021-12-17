package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/gena/grp"
	"github.com/ruraomsk/gena/sdb"
	"github.com/ruraomsk/gena/setup"
)

func init() {
	setup.Set = new(setup.Setup)
	if _, err := toml.DecodeFile("config.toml", &setup.Set); err != nil {
		fmt.Println("Can't load config file - ", err.Error())
	}
}

var err error

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	err = logger.Init("log")
	if err != nil {
		fmt.Println("Error opening logger subsystem ", err.Error())
		return
	}
	err = sdb.InitDataBase()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}
	if len(os.Args) > 1 {
		if strings.Contains(os.Args[1], "create") {
			if len(os.Args[2]) == 0 {
				fmt.Println("Нужно указать имя для сохранения")
				return
			}
			//Для отладки
			hg := new(grp.HeadGraph)
			hg.Area = 1
			hg.Region = 2
			hg.Subarea = 3
			hg.Step = 5
			hg.Date = time.Now().Format("2006-01-02")
			hg.Graph = make([]grp.GraphLine, 0)
			st := 0
			rand.Seed(time.Now().UnixNano())
			for st < 60*24 {
				p := int(rand.Intn(1000))
				o := int(rand.Intn(1000))
				hg.Graph = append(hg.Graph, grp.GraphLine{Start: st, Pr: p, Ob: o})
				st += 30
			}
			h, _ := json.Marshal(hg)
			path := setup.Set.PathData + os.Args[2] + ".json"
			fmt.Println(path)
			ioutil.WriteFile(path, h, 0644)
			fmt.Printf("file %s created\n", path)
			return
		}
		if strings.Contains(os.Args[1], "make") {
			if len(os.Args[2]) == 0 {
				fmt.Println("Нужно указать имя для исполнения")
				return
			}
			path := setup.Set.PathData + os.Args[1] + ".json"
			err = grp.MakeStat(path)
			if err != nil {
				logger.Info.Print(err.Error())
			}
			return
		}
	}
	fmt.Println("Start generator work...")
	logger.Info.Println("Start generator work...")
	path := setup.Set.PathData
	dirs, _ := ioutil.ReadDir(path)
	for _, dir := range dirs {
		if dir.IsDir() {
			continue
		}
		if strings.HasSuffix(dir.Name(), ".json") {
			err = grp.MakeStat(path + dir.Name())
			if err != nil {
				logger.Info.Printf(err.Error())
			}
		}
	}
	fmt.Println("Generator done.")
	logger.Info.Println("Generator done.")
}
