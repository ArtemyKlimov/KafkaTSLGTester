package engine

import (
	"context"
	"fmt"
	"log/slog"

	"kafkatsgltest/internal/builder"
	"kafkatsgltest/internal/config"
	"kafkatsgltest/internal/kafka"
)

// runWorker compiles its own FieldSpec (with independent rand sources),
// then generates and sends msgCount messages in batches.
// Returns the number of successfully built messages.
func runWorker(
	ctx context.Context,
	producer *kafka.Producer,
	blk config.BlockConfig,
	cfg *config.AppConfig,
	topic string,
	msgCount int,
	workerID int,
) (int, error) {
	spec, err := builder.Compile(blk.Fields, cfg.RandomWords)
	if err != nil {
		return 0, fmt.Errorf("compiling fields: %w", err)
	}

	batchSize := blk.EffectiveBatchSize(cfg.Defaults)
	sent := 0

	for sent < msgCount {
		select {
		case <-ctx.Done():
			slog.Warn("worker cancelled", "worker", workerID, "sent", sent)
			return sent, nil
		default:
		}

		end := sent + batchSize
		if end > msgCount {
			end = msgCount
		}
		for ; sent < end; sent++ {
			payload, err := builder.Build(spec)
			if err != nil {
				slog.Error("build error", "worker", workerID, "err", err)
				continue
			}
			producer.Send(topic, blk.Key, payload)
		}
	}

	slog.Debug("worker done", "worker", workerID, "sent", sent)
	return sent, nil
}
