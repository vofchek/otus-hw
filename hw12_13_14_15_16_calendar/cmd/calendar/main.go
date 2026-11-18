package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
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

const codeError = 1

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

	logg, err := logger.New(config.Logger.Level, config.Logger.Output)
	if err != nil {
		log.Fatal("cant create logger, err: ", err)
	}

	code, err := startAppServer(config, logg)
	if code > 0 {
		logg.Error(err.Error())
		logg.Close()

		os.Exit(code)
	}

	logg.Close()
}

// функция запуска приложения, создана из-за того что надо закрывать файл лога.
// чтобы не дублировать logg.Close() в коде.
func startAppServer(config Config, logg *logger.Logger) (int, error) {
	storage, err := newStorage(config)
	if err != nil {
		return codeError, fmt.Errorf("cant create storage, err: %w", err)
	}

	calendar := app.New(logg, storage)

	server := internalhttp.NewServer(logg, calendar)

	// передаём адрес для сервера(лучше передавать конфиг в server.Start)
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP,
	)
	defer cancel()

	if err := storage.Connect(ctx); err != nil {
		return codeError, fmt.Errorf("cant connect storage, err: %w", err)
	}

	// закрытие подключения к бд.
	server.RegisterOnShutdown(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := storage.Close(ctx); err != nil {
			logg.Error("cant disconnect storage, err: " + err.Error())
		}
	})

	// shutdown.
	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx, net.JoinHostPort(config.HTTP.Host, config.HTTP.Port)); err != nil {
		return codeError, fmt.Errorf("failed to start http server: %w", err)
	}

	return 0, nil
}

// фабрика хранилища. создаёт storage по коду в конфиге.
func newStorage(config Config) (app.Storage, error) {
	var storage app.Storage
	switch config.Storage.Type {
	case "memory":
		storage = memorystorage.New()
	case "sql":
		storage = sqlstorage.New(config.Storage.DSN)
	default:
		return nil, fmt.Errorf("unknown storage type %s", config.Storage.Type)
	}

	return storage, nil
}
