package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  Server  `yaml:"server"`
	Filesys Filesys `yaml:"filesys"`
}

type Server struct {
	Port     int    `yaml:"port", envconfig:"SERVER_PORT"`
	Host     string `yaml:"host", envconfig:"SERVER_HOST"`
	Protocal string `yaml:"protocal", envconfig:"SERVER_PROTOCAL"`
}

type Filesys struct {
	Chunksize    uint64 `yaml:"chunksize", envconfig:"FILESYS_CHUNKSIZE"`
	Blocksize    uint64 `yaml:"blocksize", envconfig:"FILESYS_BLOCKSIZE"`
	Savedir      string `yaml:"savedir", envconfig:"FILESYS_SAVEDIR"`
	Rootpath     string `yaml:"rootpath", envconfig:"FILESYS_ROOTPATH"`
	Compress     bool   `yaml:"compress", envconfig:"FILESYS_COMPRESS"`
	Testfilename string `yaml:"testfilename", envconfig:"FILESYS_TESTFILENAME"`
}

var C Config

func init() {
	fmt.Println("read config.yaml")
	f, err := os.Open("./config.yaml")
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&C)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
}
