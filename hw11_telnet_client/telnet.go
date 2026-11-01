package main

import (
	"errors"
	"io"
	"net"
	"time"
)

var (
	ErrClientClosed       = errors.New("client closed")
	ErrClientNotConnected = errors.New("client not connected")
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &GoClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

type GoClient struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	conn    net.Conn
}

func (client *GoClient) Connect() error {
	conn, err := net.DialTimeout("tcp", client.address, client.timeout)
	if err != nil {
		return err
	}

	client.conn = conn
	return nil
}

func (client *GoClient) Close() error {
	if client.conn == nil {
		return ErrClientNotConnected
	}
	err := client.conn.Close()
	client.conn = nil

	return err
}

func (client *GoClient) Send() error {
	if client.conn == nil {
		return ErrClientClosed
	}

	buffer := make([]byte, 1024)
	n, err := client.in.Read(buffer)
	if err != nil {
		return err
	}

	_, err = client.conn.Write(buffer[:n])
	if err != nil {
		return err
	}

	return nil
}

func (client *GoClient) Receive() error {
	if client.conn == nil {
		return ErrClientClosed
	}

	buffer := make([]byte, 1024)
	n, err := client.conn.Read(buffer)
	if err != nil {
		return err
	}

	_, err = client.out.Write(buffer[:n])
	if err != nil {
		return err
	}

	return nil
}
