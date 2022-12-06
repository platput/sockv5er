package utils

import (
	"os"
	"path/filepath"
	"strconv"
)

type Reader interface {
	Read() *Settings
}

type ConfigFileData struct{}
type ENVData struct{}

type Settings struct {
	AccessKeyId     string
	SecretKey       string
	SocksV5Port     int
	GeoLocationFile string
}

func (s *ConfigFileData) Read() *Settings {
	return &Settings{}
}

func (s *ENVData) Read() (*Settings, error) {
	accessKeyId := os.Getenv("AccessKeyId")
	secretKey := os.Getenv("SecretKey")
	socksV5PortString := os.Getenv("SocksV5Port")
	socksV5Port, err := strconv.Atoi(socksV5PortString)
	geoLocationFile := os.Getenv("GeoLocationFile")
	if geoLocationFile == "" {
		geoLocationFile = filepath.Join("assets", "IP2LOCATION-LITE-DB1.IPV6.BIN")
	}
	return &Settings{
		AccessKeyId:     accessKeyId,
		SecretKey:       secretKey,
		SocksV5Port:     socksV5Port,
		GeoLocationFile: geoLocationFile,
	}, err
}
