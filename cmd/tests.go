package main

import (
	"context"
	"fmt"
	"github.com/D1sordxr/packages/logger"
	"github.com/google/uuid"
	"math/rand"
	"time"
)

func main() {
	cfg := logger.Config{
		LogLevel:         "INFO",
		IsFormatted:      false,
		ContextLogFields: []string{"trace_id", "span_id"},
		CallerSkip:       1,
		BufferSize:       100,
	}

	log := logger.New(cfg).ToAsync()
	baseLog := logger.BaseLog{
		Operation: "test",
		Data:      logger.Data{},
	}

	log.Info("Starting test.")
	log.Infof("Hello, %d", 1)
	log.Infow("Hello, 2", "fieldA", "valueA", "fieldB", "valueB")

	log.Error("Starting test.")
	log.Errorf("Hello, %d", 1)
	log.Errorw("Hello, 2", "fieldA", "valueA", "fieldB", "valueB")

	for i := 0; i < 5; i++ {
		newData := logger.Data{
			FieldA: uuid.New().String(),
			FieldB: rand.Float64(),
			FieldC: rand.Int(),
		}
		baseLog.Data = newData
		log.Infof(fmt.Sprintf("Hello, %d", i), baseLog)
		time.Sleep(1 * time.Millisecond)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer log.Shutdown(ctx)

	log.Info("Shutdown complete.")
}
