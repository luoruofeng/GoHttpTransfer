package common

import (
	"strconv"

	cnf "github.com/luoruofeng/httpchunktransfer/config"
)

func GetAddr() string {
	return cnf.C.Server.Protocal + "://" + cnf.C.Server.Host + ":" + strconv.Itoa(cnf.C.Server.Port)
}
