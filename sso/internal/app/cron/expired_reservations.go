package cron

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"sso/sso/internal/lib/logger/sl"
	"sso/sso/internal/repository"
)

// ExpiredReservationsWorker - cron worker для отмены истёкших резервирований
type ExpiredReservationsWorker struct {
	log       *slog.Logger
	dbRepo    repository.TransactionRepository
	interval  time.Duration
	batchSize int
	stopCh    chan struct{}
	wg        sync.WaitGroup
	isRunning bool
	mu        sync.Mutex
}

// NewExpiredReservationsWorker создаёт новый worker
func NewExpiredReservationsWorker(
	log *slog.Logger,
	dbRepo repository.TransactionRepository,
	interval time.Duration,
	batchSize int,
) *ExpiredReservationsWorker {
	return &ExpiredReservationsWorker{
		log:       log.With(slog.String("component", "expired_reservations_worker")),
		dbRepo:    dbRepo,
		interval:  interval,
		batchSize: batchSize,
		stopCh:    make(chan struct{}),
	}
}

// Start запускает worker в фоновом режиме
func (w *ExpiredReservationsWorker) Start() {
	w.mu.Lock()
	if w.isRunning {
		w.mu.Unlock()
		return
	}
	w.isRunning = true
	w.mu.Unlock()

	w.wg.Add(1)
	go w.run()

	w.log.Info("expired reservations worker started",
		slog.Duration("interval", w.interval),
		slog.Int("batch_size", w.batchSize),
	)
}

// Stop останавливает worker и ждёт завершения
func (w *ExpiredReservationsWorker) Stop() {
	w.mu.Lock()
	if !w.isRunning {
		w.mu.Unlock()
		return
	}
	w.mu.Unlock()

	close(w.stopCh)
	w.wg.Wait()

	w.mu.Lock()
	w.isRunning = false
	w.mu.Unlock()

	w.log.Info("expired reservations worker stopped")
}

func (w *ExpiredReservationsWorker) run() {
	defer w.wg.Done()

	// Первый запуск сразу
	w.processExpiredReservations()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.processExpiredReservations()
		}
	}
}

func (w *ExpiredReservationsWorker) processExpiredReservations() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Получаем список истёкших резервирований
	ids, err := w.dbRepo.GetExpiredReservations(ctx, w.batchSize)
	if err != nil {
		w.log.Error("failed to get expired reservations", sl.Err(err))
		return
	}

	if len(ids) == 0 {
		return
	}

	w.log.Info("found expired reservations", slog.Int("count", len(ids)))

	// 2. Отменяем каждое резервирование
	var successCount, failCount int

	for _, id := range ids {
		_, err := w.dbRepo.CancelExpiredReservation(ctx, id)
		if err != nil {
			w.log.Error("failed to cancel expired reservation",
				slog.String("reservation_id", id),
				sl.Err(err),
			)
			failCount++
			continue
		}

		w.log.Debug("cancelled expired reservation",
			slog.String("reservation_id", id),
		)
		successCount++
	}

	w.log.Info("expired reservations processing completed",
		slog.Int("success", successCount),
		slog.Int("failed", failCount),
	)
}
