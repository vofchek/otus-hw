package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/app"
	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config := NewConfig()
	err := config.ParseFile(configFile)
	if err != nil {
		log.Fatal("cant parse config, err: ", err)
	}

	writer := io.Writer(os.Stdout)
	if path := config.Logger.Output; path != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			log.Fatal("cant create log dir, err: ", err)
		}

		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			log.Fatal("cant open log file, err: ", err)
		}
		defer f.Close()

		writer = io.MultiWriter(os.Stdout, f)
	}

	logg := logger.New(config.Logger.Level, writer)

	storage, err := createStorage(config)
	if err != nil {
		log.Fatal("cant create storage, err: ", err)
	}
	calendar := app.New(logg, storage)

	server := internalhttp.NewServer(logg, calendar)

	// передаём адрес для сервера(лучше передавать конфиг в server.Start)
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP,
	)
	defer cancel()

	if err := storage.Connect(ctx); err != nil {
		logg.Error("cant connect storage, err: " + err.Error())
		return
	}

	server.RegisterOnShutdown(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		if err := storage.Close(ctx); err != nil {
			logg.Error("cant disconnect storage, err: " + err.Error())
		}
	})

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx, net.JoinHostPort(config.HTTP.Host, config.HTTP.Port)); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}

func createStorage(config Config) (app.Storage, error) {
	switch config.Storage.Type {
	case "memory":
		return memorystorage.New(), nil
	case "sql":
		return sqlstorage.New(config.Storage.DSN), nil
	}

	return nil, errors.New("unknown storage type " + config.Storage.Type)
}
