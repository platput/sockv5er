package utils

import (
	log "github.com/sirupsen/logrus"
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
	AccessKeyId       string
	SecretKey         string
	SocksV5Port       int
	GeoLocationFile   string
	PrivateKeyPath    string
	SSHKnownHostsPath string
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
	privateKeyPath := os.Getenv("PrivateKeyPath")
	sshKnownHostsPath := os.Getenv("SSHKnownHostsPath")
	if geoLocationFile == "" {
		geoLocationFile = filepath.Join("assets", "IP2LOCATION-LITE-DB1.IPV6.BIN")
	}
	if sshKnownHostsPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Getting users home directory failed with error: %s. Please set the SSH known hosts file path in the environment to continue.\n", err)
		}
		sshKnownHostsPath = filepath.Join(homeDir, ".ssh/known_hosts")
	}
	return &Settings{
		AccessKeyId:       accessKeyId,
		SecretKey:         secretKey,
		SocksV5Port:       socksV5Port,
		GeoLocationFile:   geoLocationFile,
		PrivateKeyPath:    privateKeyPath,
		SSHKnownHostsPath: sshKnownHostsPath,
	}, err
}
