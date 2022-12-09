package utils

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/armon/go-socks5"
	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	PemFileContent string
	RemoteEndpoint string
	SocksV5Address string
	SSHUsername    string
	SSHPort        string
}

func (config *SSHConfig) StartSocksV5Server() {
	// References: https://gist.github.com/afdalwahyu/4c70868c84e68676c86e1a54b410655d
	sshConf := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.Password("SSH_PASSWORD_HERE")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshConn, err := ssh.Dial("tcp", config.RemoteEndpoint, sshConf)
	if err != nil {
		fmt.Println("error tunnel to server: ", err)
		return
	}
	defer func(sshConn *ssh.Client) {
		err := sshConn.Close()
		if err != nil {
			log.Warnf("Error occurred when trying to close the SSH Connection. Error: %s\n", err)
		}
	}(sshConn)

	log.Infoln("Connected to ssh server")

	go func() {
		conf := &socks5.Config{
			Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return sshConn.Dial(network, addr)
			},
		}

		serverSocks, err := socks5.New(conf)
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := serverSocks.ListenAndServe("tcp", config.SocksV5Address); err != nil {
			log.Fatalf("Failed to create socks5 server %s", err)
		}
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	return
}
