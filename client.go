package main

import (
	"io"
	"io/ioutil"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type ServerConnection struct {
	user string
	host string
	session *ssh.Session
}

func NewServerConnection(user string, host string) (*ServerConnection) {
	return &ServerConnection {
		user: user,
		host: host,
	}
}

func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func SSHAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

func (c *ServerConnection) Connect() error {

	sshConfig := &ssh.ClientConfig{
		User: c.user,
		Auth: []ssh.AuthMethod{
			SSHAgent(),
		},
	}

	host := strings.Join([]string{c.host, "22"}, ":")
	connection, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to dial: %s", err))
		return err
	}

	session, err := connection.NewSession()
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to create session: %s", err))
		return err
	}

	c.session = session

	return nil
}

func (c *ServerConnection) RunCommand(command []string) error {

	session := c.session

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		session.Close()
		log.Fatal(fmt.Errorf("request for pseudo terminal failed: %s", err))
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		log.Fatal(fmt.Errorf("Unable to setup stdin for session: %v", err))
	}
	go io.Copy(stdin, os.Stdin)

	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Fatal(fmt.Errorf("Unable to setup stdout for session: %v", err))
	}
	go io.Copy(os.Stdout, stdout)

	stderr, err := session.StderrPipe()
	if err != nil {
		log.Fatal(fmt.Errorf("Unable to setup stderr for session: %v", err))
	}
	go io.Copy(os.Stderr, stderr)

	err = session.Run(strings.Join(command, " "))

	return err
}
