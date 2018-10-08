package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/go-errors/errors"
	"github.com/howeyc/gopass"
	"github.com/jroimartin/gocui"
	"github.com/nsf/termbox-go"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"time"
)

type AuthType uint8

const (
	Interactive AuthType = iota
)

// Job is what is run on a server by the troupe client.
type Job struct {
	SSHAddress     string
	SSHUsername    string
	Pipeline       string
	LoggerFile     string
	Logger         io.ReadWriter
	authentication ssh.AuthMethod
	buffer         *bytes.Buffer
	sshClient      *ssh.Client
	view           *gocui.View
	pid            string
}

// Client stores information about servers and which pipelines to run on the servers.
type Client struct {
	Jobs []*Job
}

// NewClient creates a new troupe sshClient.
func NewClient() *Client {
	return &Client{}
}

// AddJob adds a job to the troupe sshClient.
func (c *Client) AddJob(username, server, pipeline, logger string, authentication AuthType) error {
	var (
		job Job
	)
	// Configure the authentication method for the ssh connection.
	switch authentication {
	case Interactive:
		fmt.Printf("password requested for authentication to %s >", server)
		pass, err := gopass.GetPasswd()
		if err != nil {
			return err
		}
		job.authentication = ssh.Password(string(pass))
	default:
		return errors.New("unknown authentication method")
	}

	// Configure the other aspects of the ssh connection.
	job.SSHUsername = username
	job.SSHAddress = server

	// The buffer is used to output the logs to the screen later.
	job.buffer = bytes.NewBuffer([]byte{})
	job.Logger = job.buffer

	job.LoggerFile = logger

	job.Pipeline = pipeline

	c.Jobs = append(c.Jobs, &job)
	return nil
}

func (c *Client) Start() error {
	var (
		err      error
		managers []gocui.Manager
	)

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer g.Close()

	jobErr := make(chan error, len(c.Jobs))
	f := func(j *Job, idx int) (func(gui *gocui.Gui) error) {
		return func(gui *gocui.Gui) error {
			maxX, maxY := gui.Size()
			height := maxY / len(c.Jobs)
			if v, err := gui.SetView(fmt.Sprintf("%s@%s", j.SSHUsername, j.SSHAddress), 0, (1+(idx+1)*height)-height, maxX-1, (idx+1)*height-1); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}
				v.Title = fmt.Sprintf("%s@%s", j.SSHUsername, j.SSHAddress)
				v.Autoscroll = true
				j.view = v
				fmt.Fprintf(j.view, " ==attached== \n")
				j.view.Highlight = false
			} else {
				return err
			}
			return nil
		}

	}
	for i, job := range c.Jobs {
		managers = append(managers, gocui.ManagerFunc(f(job, i)))
	}

	g.SetManager(managers...)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	for _, job := range c.Jobs {
		cc := &ssh.ClientConfig{
			User:    job.SSHUsername,
			Timeout: 0,
			Auth: []ssh.AuthMethod{
				job.authentication,
			},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
		}
		job.sshClient, err = ssh.Dial("tcp", job.SSHAddress, cc)
		if err != nil {
			return err
		}

		go job.Run(jobErr)
	}

	//go func() {
	//	for _ := range jobErr {
	//		//if err != nil {
	//		//	g.Update(func(gui *gocui.Gui) error {
	//		//		return nil
	//		//	})
	//		//	//fmt.Println(errors.Wrap(err, 1).ErrorStack())
	//		//}
	//	}
	//}()

	go func() {
		for {
			termbox.Interrupt()
			time.Sleep(time.Second)
		}
	}()

	g.BgColor = gocui.ColorDefault
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}

	return nil
}

func (j *Job) Run(e chan error) {
	for j.view == nil {
		time.Sleep(time.Second)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	session, err := j.sshClient.NewSession()
	if err != nil {
		j.view.FgColor = gocui.ColorBlack
		j.view.BgColor = gocui.ColorRed
		e <- err
		return
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		j.view.BgColor = gocui.ColorRed
		fmt.Fprintf(j.view, "unable to setup stdout for session: %v", err)
		e <- fmt.Errorf("unable to setup stdout for session: %v", err)
		return
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		j.view.BgColor = gocui.ColorRed
		fmt.Fprintf(j.view, "request for pseudo terminal failed: %s", err)
		e <- fmt.Errorf("request for pseudo terminal failed: %s", err)
		return
	}

	go func() {
		// Skip the first line.
		s := bufio.NewScanner(bufio.NewReader(stdout))
		s.Scan()
		for s.Scan() {
			_, err = j.view.Write(append(s.Bytes(), '\n'))
			if err != nil {
				panic(err)
			}
		}
	}()

	fmt.Fprintf(j.view, ">>> updating boogie\n")

	// Ensure the latest boogie is on the server.
	err = session.Run("go get -v -u github.com/hscells/boogie/cmd/boogie")
	if err != nil {
		j.view.BgColor = gocui.ColorRed
		fmt.Fprintf(j.view, "%s", err)
		e <- err
		return
	}

	fmt.Fprintf(j.view, ">>> copying %s\n", j.Pipeline)

	// Ensure the latest pipeline is on the server.
	f, err := os.Open(j.Pipeline)
	if err != nil {
		j.view.BgColor = gocui.ColorRed
		fmt.Fprintf(j.view, "%s", err)
		e <- err
		return
	}
	defer f.Close()

	// Create a new SCP client
	client := scp.NewClient(j.SSHAddress, &ssh.ClientConfig{
		User:    j.SSHUsername,
		Timeout: 0,
		Auth: []ssh.AuthMethod{
			j.authentication,
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	})

	// Connect to the remote server
	err = client.Connect()
	if err != nil {
		j.view.BgColor = gocui.ColorRed
		fmt.Println("Couldn't establish a connection to the remote server ", err)
		e <- err
		return
	}

	// Close client connection after the file has been copied
	defer client.Close()

	client.CopyFile(f, fmt.Sprintf("~/%s", j.Pipeline), "0655")
	if err != nil {
		j.view.BgColor = gocui.ColorRed
		fmt.Fprintf(j.view, "copy failure\n")
		fmt.Fprintf(j.view, "%s", errors.Wrap(err, 1).ErrorStack())
		e <- err
		return
	}

	session, err = j.sshClient.NewSession()
	if err != nil {
		j.view.FgColor = gocui.ColorBlack
		j.view.BgColor = gocui.ColorRed
		e <- err
		return
	}
	defer session.Close()

	stdout, err = session.StdoutPipe()
	if err != nil {
		j.view.BgColor = gocui.ColorRed
		fmt.Fprintf(j.view, "unable to setup stdout for session: %v", err)
		e <- fmt.Errorf("unable to setup stdout for session: %v", err)
		return
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		j.view.BgColor = gocui.ColorRed
		fmt.Fprintf(j.view, "request for pseudo terminal failed: %s", err)
		e <- fmt.Errorf("request for pseudo terminal failed: %s", err)
		return
	}

	go func() {
		// Skip the first line.
		s := bufio.NewScanner(bufio.NewReader(stdout))
		s.Scan()
		for s.Scan() {
			_, err = j.view.Write(append(s.Bytes(), '\n'))
			if err != nil {
				panic(err)
			}
		}
	}()

	// Run boogie with the pipeline.
	err = session.Run(fmt.Sprintf("boogie --pipeline %s --logger %s", j.Pipeline, j.LoggerFile))
	if err != nil {
		j.view.BgColor = gocui.ColorRed
		fmt.Fprintf(j.view, "request for pseudo terminal failed: %s", err)
		e <- fmt.Errorf("request for pseudo terminal failed: %s", err)
		return
	}
	j.view.FgColor = gocui.ColorBlack
	j.view.BgColor = gocui.ColorGreen
	fmt.Fprintf(j.view, ">>> compete\n")
	return
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
