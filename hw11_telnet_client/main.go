package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

var timeout time.Duration

func init() {
	flag.DurationVar(&timeout, "timeout", time.Second*10, "--timeout")
}

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("set address")
	}
	address := flag.Args()[0]

	port := ""
	if len(flag.Args()) >= 2 {
		port = flag.Args()[1]
	}

	ctx := context.Background()
	signalContext, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	inBuffer := &bytes.Buffer{}
	rwc := io.NopCloser(inBuffer)

	telnet := NewTelnetClient(net.JoinHostPort(address, port), timeout, rwc, os.Stdout)
	err := telnet.Connect()
	if err != nil {
		stop()
		log.Fatal("connection err: ", err)
	}
	defer telnet.Close()
	printServiceMessagef("...Connected to %s:%s", address, port)

	// конкурентно читаем
	go startReceiving(signalContext, stop, telnet)

	// конкурентно пишем
	go startSending(signalContext, stop, telnet, inBuffer)

	<-signalContext.Done()

	stop()

	printServiceMessagef("...Telnet exit")
}

// возвращает канал в который будет происходить запись данных из os.Stdin.
// выделено в отдельную функцию из-за создания канала.
func makeStdinGenerator(done <-chan struct{}) chan string {
	stdin := make(chan string)

	go func() {
		defer close(stdin)
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			select {
			case <-done:
				return
			default:
			}

			stdin <- scanner.Text() + "\n"
		}
		if scanner.Err() != nil {
			fmt.Println("highly unlikely")
		}
	}()

	return stdin
}

func startReceiving(signalContext context.Context, stop context.CancelFunc, telnet TelnetClient) {
	recTicker := time.NewTicker(time.Millisecond * 100)
	defer func() {
		stop()
		recTicker.Stop()
	}()

	for {
		select {
		case <-signalContext.Done():
			return
		case <-recTicker.C:
			select {
			case <-signalContext.Done():
				return
			default:
			}

			err := telnet.Receive()
			if err != nil {
				if errors.Is(err, io.EOF) {
					printServiceMessagef("...Connection was closed by peer")
				} else if !errors.Is(err, ErrClientClosed) && !errors.Is(err, net.ErrClosed) {
					printServiceMessagef("err on receive: %v", err)
				}
				return
			}
		}
	}
}

func startSending(signalContext context.Context, stop context.CancelFunc, telnet TelnetClient, in io.Writer) {
	stdin := makeStdinGenerator(signalContext.Done())
	defer stop()

	for {
		select {
		case <-signalContext.Done():
			return
		case line, ok := <-stdin:
			if !ok {
				printServiceMessagef("...EOF")
				return
			}

			select {
			case <-signalContext.Done():
				return
			default:
			}

			io.WriteString(in, line)
			err := telnet.Send()
			if err != nil {
				if errors.Is(err, io.EOF) {
					printServiceMessagef("...EOF")
				} else if !errors.Is(err, ErrClientClosed) && !errors.Is(err, net.ErrClosed) {
					printServiceMessagef("err on send: %v", err)
				}
				return
			}
		}
	}
}

func printServiceMessagef(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", v...)
}
