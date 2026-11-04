package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
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

	logg := logger.New(config.Logger.Level, os.Stdout)

	storage, err := createStorage(config)
	if err != nil {
		log.Fatal("cant create storage, err: ", err)
	}
	calendar := app.New(logg, storage)

	server := internalhttp.NewServer(logg, calendar)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
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
		return sqlstorage.New(), nil
	}

	return nil, errors.New("unknown storage type " + config.Storage.Type)
}
