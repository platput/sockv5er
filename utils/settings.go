package utils

import (
	"os"
	"strconv"
)

type Reader interface {
	Read() *Settings
}

type ConfigFileData struct{}
type ENVData struct{}

type Settings struct {
	AccessKeyId string
	SecretKey   string
	SocksV5Port int
}

func (s *ConfigFileData) Read() *Settings {
	return &Settings{}
}

func (s *ENVData) Read() (*Settings, error) {
	accessKeyId := os.Getenv("AccessKeyId")
	secretKey := os.Getenv("SecretKey")
	socksV5PortString := os.Getenv("SocksV5Port")
	socksV5Port, err := strconv.Atoi(socksV5PortString)
	return &Settings{
		AccessKeyId: accessKeyId,
		SecretKey:   secretKey,
		SocksV5Port: socksV5Port,
	}, err
}
