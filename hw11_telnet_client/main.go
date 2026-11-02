package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

var timeout time.Duration

func init() {
	flag.DurationVar(&timeout, "timeout", time.Second*10, "--timeout 10s")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "telnet - простой клиент telnet\n\n")
		fmt.Fprintf(os.Stderr, "Использование:\n  telnet [флаги] [адрес] [порт]\n\n")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		log.Fatal("no address")
	}
	address := flag.Args()[0]

	if len(flag.Args()) < 2 {
		flag.Usage()
		log.Fatal("no port")
	}
	port := flag.Args()[1]

	ctx := context.Background()
	signalContext, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	telnet := NewTelnetClient(net.JoinHostPort(address, port), timeout, os.Stdin, os.Stdout)
	err := telnet.Connect()
	if err != nil {
		stop()
		log.Fatal("connection err: ", err)
	}
	defer telnet.Close()
	printServiceMessagef("...Connected to %s:%s\n", address, port)

	// конкурентно читаем
	go func() {
		defer stop()

		err := telnet.Receive()
		if err != nil {
			if !errors.Is(err, ErrClientClosed) && !errors.Is(err, net.ErrClosed) {
				printServiceMessagef("err on receive: %v\n", err)
			}
			return
		}

		printServiceMessagef("...Connection was closed by peer\n")
	}()

	// конкурентно пишем
	go func() {
		defer stop()

		err := telnet.Send()
		if err != nil {
			if !errors.Is(err, ErrClientClosed) && !errors.Is(err, net.ErrClosed) {
				printServiceMessagef("err on send: %v\n", err)
			}
			return
		}
		printServiceMessagef("...EOF\n")
	}()

	// ждём прекращения работы
	<-signalContext.Done()

	printServiceMessagef("...Telnet exit\n")
}

func printServiceMessagef(format string, v ...any) {
	fmt.Fprintf(os.Stderr, format, v...)
}
