package common

import (
	"log"
	"os"
	"path/filepath"

	cnf "github.com/luoruofeng/httpchunktransfer/config"
)

func InitDir() {
	if _, err := os.Stat(cnf.C.Filesys.Savedir); err == nil || os.IsExist(err) {
		log.Println("remove save dir:", cnf.C.Filesys.Savedir)
		os.RemoveAll(filepath.Join(cnf.C.Filesys.Savedir, "/"))
	}

	os.Mkdir(cnf.C.Filesys.Savedir, 0755)

	if _, err := os.Stat(cnf.C.Filesys.Rootpath); err == nil || os.IsExist(err) {
		if err != nil && os.IsNotExist(err) {
			os.Mkdir(cnf.C.Filesys.Rootpath, 0755)
			log.Println("create root path dir")
		} else {
			log.Println("root path is already exist")
		}
	}

}
