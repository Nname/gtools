package ssh

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"time"
)

type Client struct {
	Timeout  time.Duration
	Host     string
	User     string
	Password string
	Port     int
}

func (c *Client) Connections() (*ssh.Client, *ssh.Session, error) {
	config := &ssh.ClientConfig{
		Timeout:         time.Second * c.Timeout,
		User:            c.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // not secure
		Auth:            []ssh.AuthMethod{ssh.Password(c.Password)},
	}
	dial, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), config)
	if err != nil {
		return nil, nil, err
	}
	session, err := dial.NewSession()
	if err != nil {
		return nil, nil, err
	}
	return dial, session, nil
}

func (c *Client) RunCommand(command string) (string, error) {
	connections, session, err := c.Connections()
	if err != nil {
		return "", err
	}
	defer session.Close()
	defer connections.Close()
	output, err := session.CombinedOutput(command)
	return string(output), err
}

func (c *Client) RunCommands(commands []string) ([]string, error) {
	var result []string
	var resultErr error
	for i := 0; i < len(commands); i++ {
		output, err := c.RunCommand(commands[i])
		if err != nil {
			resultErr = err
		}
		result = append(result, output)
	}
	return result, resultErr
}

func (c *Client) RunCommandPty(command string) {
	connections, session, err := c.Connections()
	if err != nil {
		return
	}
	defer session.Close()
	defer connections.Close()
	session.RequestPty("xtrem", 120, 120, ssh.TerminalModes{})
	Stdout, err := session.StdoutPipe()
	if err != nil {
		fmt.Println("StdoutPipe Error, ", err)
	}
	Stderr, err := session.StderrPipe()
	if err != nil {
		fmt.Println("StderrPipe Error, ", err)
	}
	StartErr := session.Start(command)
	if err != nil {
		fmt.Println("Start Error, ", StartErr)
	}
	StdoutScanner := bufio.NewScanner(Stdout)
	for StdoutScanner.Scan() {
		fmt.Println(StdoutScanner.Text())
	}
	StderrScanner := bufio.NewScanner(Stderr)
	for StderrScanner.Scan() {
		fmt.Println(StderrScanner.Text())
	}
	if WaitErr := session.Wait(); err != nil {
		fmt.Println("WaitErr, ", WaitErr)
	}
}
