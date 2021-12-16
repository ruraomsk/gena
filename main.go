package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ruraomsk/TLServer/logger"
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
	if len(os.Args) > 1 {
		if strings.Contains(os.Args[1], "create") {
			return
		}
		if strings.Contains(os.Args[1], "save") {
			return
		}
		if strings.Contains(os.Args[1], "update") {
			if len(os.Args[2]) == 0 || len(os.Args[3]) == 0 {
				fmt.Println("Нужно запускать с параметрами номер региона имя файла копии базы")
				return
			}
			return
		}

	}
	fmt.Println("Start generator work...")
	logger.Info.Println("Start generator work...")

	fmt.Println("Generator done.")
	logger.Info.Println("Generator done.")
}
