package main

import (
	"flag"
	"fmt"
	"github.com/liuhaogui/nat-ddns/service/server"
)

var ss = server.ServerService{}

func main() {
	fmt.Println("server start>>>>>>>>>>>>>>>>>>>>>")

	var config string
	flag.StringVar(&config, "config", "./config.json", "--config ./config.json")

	flag.Parse()

	ss.ServerInit(config)

	ss.ServerStart()

	fmt.Println("server end<<<<<<<<<<<<<<<<<<<<<<<")
}
