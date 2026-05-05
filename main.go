package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"kafkatsgltest/internal/config"
	"kafkatsgltest/internal/engine"
	"kafkatsgltest/internal/kafka"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config file")
	verbose := flag.Bool("verbose", false, "enable debug logging")
	flag.Parse()

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	cfg, err := config.Load(*cfgPath)
	exitOnErr("loading config", err)

	saramaCfg, err := kafka.NewSaramaConfig(cfg.Kafka)
	exitOnErr("building kafka config", err)

	producer, err := kafka.NewProducer(cfg.Kafka.BrokerAddresses(), saramaCfg)
	exitOnErr("creating producer", err)
	defer func() {
		if err := producer.Close(); err != nil {
			slog.Error("closing producer", "err", err)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	eng := engine.New(cfg, producer)
	if err := eng.Run(ctx); err != nil {
		slog.Error("engine error", "err", err)
		os.Exit(1)
	}
}

func exitOnErr(op string, err error) {
	if err != nil {
		slog.Error(op, "err", err)
		os.Exit(1)
	}
}
