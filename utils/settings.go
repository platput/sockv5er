package utils

import (
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type Reader interface {
	Read() *Settings
}

type ConfigFileData struct{}
type ENVData struct{}

type Settings struct {
	AccessKeyId       string
	SecretKey         string
	SocksV5Host       string
	SocksV5Port       string
	GeoLocationFile   string
	PrivateKeyPath    string
	SSHKnownHostsPath string
	SSHUserName       string
	SSHPort           string
	TrackingFilepath  string
}

func (s *ConfigFileData) Read() *Settings {
	return &Settings{}
}

func (s *ENVData) Read() *Settings {
	accessKeyId := os.Getenv("ACCESS_KEY_ID")
	secretKey := os.Getenv("SECRET_KEY")
	socksV5Host := os.Getenv("SOCKS_V5_HOST")
	if socksV5Host == "" {
		socksV5Host = "127.0.0.1"
	}
	socksV5Port := os.Getenv("SOCKS_V5_PORT")
	geoLocationFile := os.Getenv("GEO_LOCATION_FILE")
	privateKeyPath := os.Getenv("PRIVATE_KEY_PATH")
	sshKnownHostsPath := os.Getenv("SSH_KNOWN_HOSTS_PATH")
	if geoLocationFile == "" {
		geoLocationFile = filepath.Join("assets", "IP2LOCATION-LITE-DB1.IPV6.BIN")
	}
	sshUsername := os.Getenv("SSH_USERNAME")
	if sshUsername == "" {
		sshUsername = "ec2-user"
	}
	if sshKnownHostsPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Getting users home directory failed with error: %s. Please set the SSH known hosts file path in the environment to continue.\n", err)
		}
		sshKnownHostsPath = filepath.Join(homeDir, ".ssh/known_hosts")
	}
	sshPort := os.Getenv("SSH_PORT")
	if sshPort == "" {
		sshPort = "22"
	}
	return &Settings{
		AccessKeyId:       accessKeyId,
		SecretKey:         secretKey,
		SocksV5Host:       socksV5Host,
		SocksV5Port:       socksV5Port,
		GeoLocationFile:   geoLocationFile,
		PrivateKeyPath:    privateKeyPath,
		SSHKnownHostsPath: sshKnownHostsPath,
		SSHUserName:       sshUsername,
		SSHPort:           sshPort,
	}
}
