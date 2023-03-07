package utils

import (
	"bytes"
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
	PrivateKey         []byte
	KnownHostsFilepath string
	SSHHost            string
	SSHPort            string
	SSHUsername        string
	SocksV5IP          string
	SocksV5Port        string
}

func (config *SSHConfig) StartSocksV5Server() {
	// References:
	// 1. https://gist.github.com/afdalwahyu/4c70868c84e68676c86e1a54b410655d
	// 2. https://pkg.go.dev/golang.org/x/crypto/ssh#PublicKeys
	// 3. https://stackoverflow.com/questions/45441735/ssh-handshake-complains-about-missing-host-key
	sshConn, err := config.connectToSSH()
	if err != nil {
		log.Fatalf("SSH Connection failed with error: %s\n", err)
	}
	defer func(sshConn *ssh.Client) {
		err := sshConn.Close()
		if err != nil {
			log.Warnf("Error occurred when trying to close the SSH Connection. Error: %s\n", err)
		}
	}(sshConn)
	log.Infoln("Connected to ssh server")
	ch := make(chan os.Signal)

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
		socksV5Address := fmt.Sprintf("%s:%s", config.SocksV5IP, config.SocksV5Port)
		if err := serverSocks.ListenAndServe("tcp", socksV5Address); err != nil {
			log.Fatalf("Failed to create socks5 server %s\n", err)
		}
	}()
	log.Infoln("Started SocksV5 server.")
	log.Infoln("Press CTRL+C to stop SocksV5 server and exit!")

	go func(ch chan os.Signal) {
		exitSignal := <-ch
		if exitSignal == syscall.SIGINT || exitSignal == syscall.SIGTERM {
			config.cleanup()
			os.Exit(0)
		}
	}(ch)

	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	return
}

func (config *SSHConfig) cleanup() {
	log.Infoln("Stopping SocksV5 Server")
	log.Infoln("Terminating EC2 Instance.")
	session, err := config.GetNewSSHSession()
	log.Warnf("Cleaning up resources fauled with err: %s", err)
	defer func(session *ssh.Session) {
		err := session.Close()
		if err != nil {

		}
	}(session)
	commandsToExecute := []string{"sudo shutdown now", ""}
	config.IssueCommandsViaSSH(session, commandsToExecute)
	log.Infoln("All clean up done without any errors.")
	log.Infoln("Exiting...")
}

func (config *SSHConfig) connectToSSH() (*ssh.Client, error) {
	signer, err := ssh.ParsePrivateKey(config.PrivateKey)
	if err != nil {
		log.Fatalf("Unable to parse private key: %v\n", err)
	}
	sshConf := &ssh.ClientConfig{
		User:            config.SSHUsername,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	hostWithPort := fmt.Sprintf("%s:%s", config.SSHHost, config.SSHPort)
	sshConn, err := ssh.Dial("tcp", hostWithPort, sshConf)
	return sshConn, err
}

func (config *SSHConfig) GetNewSSHSession() (*ssh.Session, error) {
	client, err := config.connectToSSH()
	if err != nil {
		return nil, err
	}
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (config *SSHConfig) IssueCommandsViaSSH(session *ssh.Session, commandsToExecute []string) {
	defer func(session *ssh.Session) {
		err := session.Close()
		if err != nil {
			log.Warnf("Closing SSH Session failed with error: %s\n", err)
		}
	}(session)
	for i := range commandsToExecute {
		var output bytes.Buffer
		session.Stdout = &output
		err := session.Run(commandsToExecute[i])
		if err != nil {
			log.Warnf("Executing command `%s` failed with error: %s\n", commandsToExecute[i], err)
		}
		log.Infof("`%s` returned `%s` as output.\n", commandsToExecute[i], output.String())
	}
}
