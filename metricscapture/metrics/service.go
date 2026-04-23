package metrics

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	MaxQueueSize int
	BatchSize    int
}

type Service struct {
	queue *boundedQueue

	uploader Uploader

	maxQueueSize int
	batchSize    int

	mu             sync.Mutex
	lastUploadAt   time.Time
	lastUploadErr  string
	uploadedCount  uint64
	uploadFailCount uint64
}

func NewService(cfg Config, uploader Uploader) *Service {
	maxQueue := cfg.MaxQueueSize
	if maxQueue <= 0 {
		maxQueue = 1000
	}
	batch := cfg.BatchSize
	if batch <= 0 {
		batch = 100
	}
	return &Service{
		queue:        newBoundedQueue(maxQueue),
		uploader:     uploader,
		maxQueueSize: maxQueue,
		batchSize:    batch,
	}
}

func (s *Service) Emit(name string, value float64, tags map[string]string) {
	if strings.TrimSpace(name) == "" {
		return
	}
	s.queue.enqueue(Metric{
		Name:      name,
		Value:     value,
		Timestamp: time.Now().UTC(),
		Tags:      tags,
	})
}

func (s *Service) FlushNow(ctx context.Context) (FlushResult, error) {
	if s.uploader == nil {
		return FlushResult{}, fmt.Errorf("uploader is not configured")
	}
	out := FlushResult{}

	for {
		batch := s.queue.popN(s.batchSize)
		if len(batch) == 0 {
			break
		}

		fileIDs, err := s.uploader.Upload(ctx, batch)
		if err != nil {
			// Put failed batch back to queue front for next attempt.
			s.queue.prepend(batch)
			atomic.AddUint64(&s.uploadFailCount, 1)
			s.setLastUpload(time.Time{}, err.Error())
			return out, err
		}

		atomic.AddUint64(&s.uploadedCount, uint64(len(batch)))
		out.UploadedCount += len(batch)
		out.FileIDs = append(out.FileIDs, fileIDs...)
	}

	s.setLastUpload(time.Now().UTC(), "")
	return out, nil
}

func (s *Service) Stats() Stats {
	s.mu.Lock()
	lastAt := s.lastUploadAt
	lastErr := s.lastUploadErr
	s.mu.Unlock()

	return Stats{
		QueueDepth:      s.queue.len(),
		DroppedCount:    s.queue.droppedCount(),
		UploadedCount:   atomic.LoadUint64(&s.uploadedCount),
		UploadFailCount: atomic.LoadUint64(&s.uploadFailCount),
		LastUploadAt:    lastAt,
		LastUploadError: lastErr,
		MaxQueueSize:    s.maxQueueSize,
		BatchSize:       s.batchSize,
	}
}

func (s *Service) setLastUpload(at time.Time, uploadErr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !at.IsZero() {
		s.lastUploadAt = at
	}
	s.lastUploadErr = uploadErr
}
