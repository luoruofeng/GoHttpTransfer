package server

import (
	"fmt"
	"net/http"

	cnf "github.com/luoruofeng/httpchunktransfer/config"
)

func Start() {
	http.HandleFunc("/info", info)
	http.HandleFunc("/hrange", hrange)
	http.ListenAndServe(fmt.Sprintf(":%v", cnf.C.Server.Port), nil)
}
