package clicks

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/shehansuraweera/linkr/internal/domain"
	"github.com/shehansuraweera/linkr/internal/repository"
)

// Flusher is satisfied by *repository.ClickRepo and any test double.
type Flusher interface {
	FlushBatch(ctx context.Context, batch []domain.ClickEvent) error
}

// Pipeline buffers click events and flushes them in batches.
// The redirect handler enqueues non-blocking; workers drain and write to Postgres.
type Pipeline struct {
	queue         chan domain.ClickEvent
	clicks        Flusher
	batchSize     int
	flushInterval time.Duration
	workers       int
	logger        *slog.Logger
	wg            sync.WaitGroup
}

func NewPipeline(
	clicks *repository.ClickRepo,
	bufferSize, batchSize, workers int,
	flushInterval time.Duration,
	logger *slog.Logger,
) *Pipeline {
	return NewPipelineWithRepo(clicks, bufferSize, batchSize, workers, flushInterval, logger)
}

// NewPipelineWithRepo accepts the Flusher interface, enabling test doubles.
func NewPipelineWithRepo(
	clicks Flusher,
	bufferSize, batchSize, workers int,
	flushInterval time.Duration,
	logger *slog.Logger,
) *Pipeline {
	return &Pipeline{
		queue:         make(chan domain.ClickEvent, bufferSize),
		clicks:        clicks,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		workers:       workers,
		logger:        logger,
	}
}

// Enqueue adds an event to the queue without blocking.
// If the buffer is full the event is dropped — redirect latency is never sacrificed for analytics.
func (p *Pipeline) Enqueue(e domain.ClickEvent) bool {
	select {
	case p.queue <- e:
		return true
	default:
		p.logger.Warn("click queue full, dropping event", "link_id", e.LinkID)
		return false
	}
}

// Start launches the worker goroutines.
func (p *Pipeline) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// Stop closes the queue and waits for all workers to flush their final batches.
func (p *Pipeline) Stop() {
	close(p.queue)
	p.wg.Wait()
}

func (p *Pipeline) worker() {
	defer p.wg.Done()
	batch := make([]domain.ClickEvent, 0, p.batchSize)
	ticker := time.NewTicker(p.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case e, ok := <-p.queue:
			if !ok {
				// Channel closed — drain remaining and exit.
				if len(batch) > 0 {
					p.flush(batch)
				}
				return
			}
			batch = append(batch, e)
			if len(batch) >= p.batchSize {
				p.flush(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				p.flush(batch)
				batch = batch[:0]
			}
		}
	}
}

func (p *Pipeline) flush(batch []domain.ClickEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := p.clicks.FlushBatch(ctx, batch); err != nil {
		p.logger.Error("click flush failed", "err", err, "batch_size", len(batch))
	}
}
