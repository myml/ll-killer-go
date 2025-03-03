/*
* Copyright (c) 2025 System233
*
* This software is released under the MIT License.
* https://opensource.org/licenses/MIT
 */
package main

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/creack/pty"
	"github.com/moby/term"
)

type Type uint8

type Channel struct {
	io   net.Conn
	argv []string
}

func (s *Channel) sendArgvEnd() error {
	err := binary.Write(s.io, binary.LittleEndian, uint32(0))
	if err != nil {
		return err
	}
	return nil
}
func (s *Channel) sendArgv(argv string) error {
	data := []byte(argv)
	err := binary.Write(s.io, binary.LittleEndian, uint32(len(data)))
	if err != nil {
		return err
	}
	err = binary.Write(s.io, binary.LittleEndian, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Channel) run() error {
	defer s.io.Close()
	if term.IsTerminal(os.Stdin.Fd()) {
		state, err := term.SetRawTerminal(os.Stdin.Fd())
		if err != nil {
			log.Printf("Failed to set raw terminal: %v", err)
		} else {
			defer term.RestoreTerminal(os.Stdin.Fd(), state)
		}
	}

	go func() {
		_, _ = io.Copy(s.io, os.Stdin)
	}()
	_, _ = io.Copy(os.Stdout, s.io)
	return nil
}
func (s *Channel) ReadArg() ([]byte, error) {
	var len uint32
	if err := binary.Read(s.io, binary.LittleEndian, &len); err != nil {
		return nil, err
	}
	if len == 0 {
		return nil, nil
	}
	var err error
	buffer := make([]byte, len)
	_, err = io.ReadFull(s.io, buffer)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}
func (s *Channel) ParseArgs() ([]string, error) {
	var args []string

	for {
		item, err := s.ReadArg()
		if err != nil {
			return nil, err
		}
		if item == nil {
			break
		}
		args = append(args, string(item))
	}
	if len(args) == 0 {
		args = append(args, DefaultShell())
	}
	return args, nil
}

func (s *Channel) serve() error {
	args, err := s.ParseArgs()
	Debug("Server:ParseArgs", err)
	if err != nil {
		return err
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	ptmx, err := pty.Start(cmd)
	Debug("Server:pty:Start", err)
	if err != nil {
		return err
	}
	defer ptmx.Close()
	go func() {
		_, _ = ptmx.ReadFrom(s.io)
	}()
	go func() {
		_, _ = ptmx.WriteTo(s.io)
	}()
	err = cmd.Wait()
	Debug("Server:cmd:Wait", err)
	return nil
}
func CreateChannel(conn net.Conn) *Channel {
	channel := &Channel{
		io: conn,
	}
	return channel
}

type ChannelFlags struct {
	Unix            string
	Server          bool
	AutoExit        bool
	Timeout         time.Duration
	connectionCount int32
	connectionTimes int32
}

func (flags *ChannelFlags) StartServer() error {
	Debug("Server")
	var signal chan int = make(chan int)
	socketPath := flags.Unix
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		Debug("Server:remove", err)
		return err
	}
	listener, err := net.Listen("unix", socketPath)
	Debug("Server:Listen", err)
	if err != nil {
		return err
	}
	defer os.Remove(socketPath)
	defer listener.Close()
	Debug("Server is listening on", socketPath)

	if flags.Timeout != 0 {
		go func() {
			time.Sleep(flags.Timeout)
			signal <- 1
		}()
	}
	go func() {
		for {
			conn, err := listener.Accept()
			Debug("Server:Accept", err)
			atomic.AddInt32(&flags.connectionTimes, 1)
			if err != nil {
				signal <- 1
				if errors.Is(err, net.ErrClosed) {
					break
				}
				continue
			}
			go func() {
				Debug("Server:handle")
				defer conn.Close()
				atomic.AddInt32(&flags.connectionCount, 1)
				channel := CreateChannel(conn)
				Debug("Server:CreateChannel")
				err := channel.serve()
				Debug("Server:Serve", err)
				atomic.AddInt32(&flags.connectionCount, -1)
				signal <- 1
			}()
		}
	}()
	for {
		<-signal
		if flags.AutoExit {
			if atomic.LoadInt32(&flags.connectionTimes) > 0 && atomic.LoadInt32(&flags.connectionCount) == 0 {
				Debug("Server:Exit:AutoExit")
				return nil
			}
		}
		if flags.Timeout != 0 {
			if atomic.LoadInt32(&flags.connectionTimes) == 0 {
				Debug("Server:Exit:Timeout")
				return nil
			}
		}
	}
}
func (flags *ChannelFlags) connect() (net.Conn, error) {
	// return net.DialTimeout("unix", flags.Unix, flags.Timeout)
	startTime := time.Now()
	for {
		conn, err := net.Dial("unix", flags.Unix)
		if err != nil {
			if flags.Timeout != 0 && time.Since(startTime) < flags.Timeout {
				time.Sleep(50 * time.Millisecond)
			} else {
				return nil, err
			}
		} else {
			return conn, nil
		}
	}
}
func (flags *ChannelFlags) StartClient(args []string) error {
	Debug("Client")
	conn, err := flags.connect()
	Debug("Client:connect", err)
	if err != nil {
		return err
	}

	defer conn.Close()

	channel := CreateChannel(conn)
	for _, item := range args {
		channel.sendArgv(item)
	}
	err = channel.sendArgvEnd()
	Debug("Client:sendArgvEnd", err)
	if err != nil {
		return err
	}
	err = channel.run()
	Debug("Client:run", err)
	return err
}

func NewChannelFlags() *ChannelFlags {
	return &ChannelFlags{
		Unix:     "",
		Server:   false,
		AutoExit: true,
		Timeout:  time.Second * 30,
	}
}
