package main

import (
	"log"
	"os"
	"time"

	"github.com/luoruofeng/httpchunktransfer/client"
	"github.com/luoruofeng/httpchunktransfer/common"
	"github.com/luoruofeng/httpchunktransfer/config"
	"github.com/luoruofeng/httpchunktransfer/server"
)

func main() {
	common.InitDir()
	go func() {
		time.Sleep(time.Second)
		client.Query(config.C.Filesys.Testfilename)
		os.Exit(1)
	}()
	log.Println("server start at port:", config.C.Server.Port)
	server.Start()
}
