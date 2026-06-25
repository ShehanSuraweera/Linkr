package clicks_test

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shehansuraweera/linkr/internal/clicks"
	"github.com/shehansuraweera/linkr/internal/domain"
)

// fakeClickRepo records flushed batches without hitting a real DB.
type fakeClickRepo struct {
	mu      sync.Mutex
	batches [][]domain.ClickEvent
	total   int64
}

func (f *fakeClickRepo) FlushBatch(_ context.Context, batch []domain.ClickEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make([]domain.ClickEvent, len(batch))
	copy(cp, batch)
	f.batches = append(f.batches, cp)
	atomic.AddInt64(&f.total, int64(len(batch)))
	return nil
}

func newTestPipeline(repo *fakeClickRepo, bufSize, batchSize int, flush time.Duration) *clicks.Pipeline {
	return clicks.NewPipelineWithRepo(repo, bufSize, batchSize, 2, flush, slog.Default())
}

func TestPipeline_FlushOnBatchSize(t *testing.T) {
	repo := &fakeClickRepo{}
	p := newTestPipeline(repo, 1000, 5, 10*time.Second) // flush every 5 events
	p.Start()

	for i := 0; i < 10; i++ {
		p.Enqueue(domain.ClickEvent{LinkID: 1, At: time.Now()})
	}

	p.Stop()

	if atomic.LoadInt64(&repo.total) != 10 {
		t.Errorf("expected 10 clicks flushed, got %d", repo.total)
	}
}

func TestPipeline_FlushOnTicker(t *testing.T) {
	repo := &fakeClickRepo{}
	p := newTestPipeline(repo, 1000, 100, 50*time.Millisecond) // batch of 100, but ticker fires first
	p.Start()

	p.Enqueue(domain.ClickEvent{LinkID: 1, At: time.Now()})
	time.Sleep(150 * time.Millisecond)
	p.Stop()

	if atomic.LoadInt64(&repo.total) != 1 {
		t.Errorf("expected 1 click flushed via ticker, got %d", repo.total)
	}
}

func TestPipeline_DropsWhenFull(t *testing.T) {
	repo := &fakeClickRepo{}
	// Buffer of 2, batch of 1000 (so nothing flushes during the test), flush interval huge.
	p := newTestPipeline(repo, 2, 1000, 10*time.Second)
	// Do NOT start workers — queue stays full after 2 events.

	ok1 := p.Enqueue(domain.ClickEvent{LinkID: 1, At: time.Now()})
	ok2 := p.Enqueue(domain.ClickEvent{LinkID: 1, At: time.Now()})
	dropped := p.Enqueue(domain.ClickEvent{LinkID: 1, At: time.Now()}) // should drop

	if !ok1 || !ok2 {
		t.Error("first two enqueues should succeed")
	}
	if dropped {
		t.Error("third enqueue should drop (queue full), but returned true")
	}
}

func TestPipeline_DrainOnStop(t *testing.T) {
	repo := &fakeClickRepo{}
	p := newTestPipeline(repo, 1000, 500, 10*time.Second) // batch of 500, we send fewer
	p.Start()

	const n = 37
	for i := 0; i < n; i++ {
		p.Enqueue(domain.ClickEvent{LinkID: 1, At: time.Now()})
	}
	p.Stop() // must drain remaining events before returning

	if atomic.LoadInt64(&repo.total) != n {
		t.Errorf("expected %d clicks after drain, got %d", n, repo.total)
	}
}

func TestPipeline_ConcurrentEnqueue_RaceDetector(t *testing.T) {
	repo := &fakeClickRepo{}
	p := newTestPipeline(repo, 10_000, 100, 50*time.Millisecond)
	p.Start()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				p.Enqueue(domain.ClickEvent{LinkID: int64(j % 5), At: time.Now()})
			}
		}()
	}
	wg.Wait()
	p.Stop()

	// All 5000 events should have been flushed (buffer is large enough).
	if atomic.LoadInt64(&repo.total) != 5000 {
		t.Errorf("expected 5000 clicks, got %d", repo.total)
	}
}
