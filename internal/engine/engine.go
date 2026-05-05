package engine

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"kafkatsgltest/internal/config"
	"kafkatsgltest/internal/kafka"
)

// Engine orchestrates all blocks in parallel worker pools.
type Engine struct {
	cfg      *config.AppConfig
	producer *kafka.Producer
}

func New(cfg *config.AppConfig, producer *kafka.Producer) *Engine {
	return &Engine{cfg: cfg, producer: producer}
}

// Run launches a worker pool for each block and waits for all to finish.
// Returns an error only if a block cannot be started (e.g. bad field config).
func (e *Engine) Run(ctx context.Context) error {
	start := time.Now()
	var totalSent atomic.Int64
	var wg sync.WaitGroup

	for i, blk := range e.cfg.Blocks {
		blk := blk

		workers := blk.EffectiveWorkers(e.cfg.Defaults)
		total := blk.Count
		perWorker := total / workers
		remainder := total % workers

		topic := blk.Topic
		if topic == "" {
			topic = e.cfg.Kafka.Topic
		}

		slog.Info("starting block",
			"index", i,
			"count", total,
			"workers", workers,
			"topic", topic,
			"key", blk.Key)

		for w := 0; w < workers; w++ {
			count := perWorker
			if w == 0 {
				count += remainder
			}
			if count == 0 {
				continue
			}

			wg.Add(1)
			go func(workerID, msgCount int) {
				defer wg.Done()
				sent, err := runWorker(ctx, e.producer, blk, e.cfg, topic, msgCount, workerID)
				totalSent.Add(int64(sent))
				if err != nil {
					slog.Error("worker error", "block", i, "worker", workerID, "err", err)
				}
			}(w, count)
		}
	}

	wg.Wait()
	elapsed := time.Since(start)
	n := totalSent.Load()
	slog.Info("done",
		"sent", n,
		"elapsed", elapsed.Round(time.Millisecond),
		"msg_per_sec", fmt.Sprintf("%.0f", float64(n)/elapsed.Seconds()))
	return nil
}
