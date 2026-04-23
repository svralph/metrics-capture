package test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"bov-2/metricscapture/metrics"
)

type mockTabularDataClient struct {
	mu         sync.Mutex
	callCount  int
	failOnCall map[int]error
}

func (m *mockTabularDataClient) TabularDataCaptureUpload(_ context.Context, _ metrics.TabularUploadRequest) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if err, shouldFail := m.failOnCall[m.callCount]; shouldFail {
		return "", err
	}
	return fmt.Sprintf("file-%d", m.callCount), nil
}

func (m *mockTabularDataClient) Calls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func TestFlushNowSuccessEmptiesQueue(t *testing.T) {
	mockClient := &mockTabularDataClient{
		failOnCall: map[int]error{},
	}
	uploader := metrics.NewDataClientUploader(
		mockClient,
		"part-1",
		"rdk:component:sensor",
		"metricscapture",
		"DoCommandFlush",
		[]string{"metrics"},
	)
	svc := metrics.NewService(metrics.Config{
		MaxQueueSize: 10,
		BatchSize:    2,
	}, uploader)

	svc.Emit("brain.do_command_total", 1, map[string]string{"command": "start"})
	svc.Emit("brain.auto_mode_loop_total", 1, nil)
	svc.Emit("brain.move_cancelled_total", 1, map[string]string{"reason": "bumper"})

	res, err := svc.FlushNow(context.Background())
	if err != nil {
		t.Fatalf("expected flush success, got error: %v", err)
	}
	if res.UploadedCount != 3 {
		t.Fatalf("expected uploaded_count=3, got %d", res.UploadedCount)
	}
	if got := len(res.FileIDs); got != 2 {
		t.Fatalf("expected 2 file IDs for 2 batches, got %d", got)
	}

	stats := svc.Stats()
	if stats.QueueDepth != 0 {
		t.Fatalf("expected queue_depth=0, got %d", stats.QueueDepth)
	}
	if stats.UploadFailCount != 0 {
		t.Fatalf("expected upload_fail_count=0, got %d", stats.UploadFailCount)
	}
	if stats.UploadedCount != 3 {
		t.Fatalf("expected uploaded_count=3 in stats, got %d", stats.UploadedCount)
	}
	if mockClient.Calls() != 2 {
		t.Fatalf("expected 2 upload calls, got %d", mockClient.Calls())
	}
}

func TestFlushNowFailureRequeuesBatch(t *testing.T) {
	mockClient := &mockTabularDataClient{
		failOnCall: map[int]error{
			1: errors.New("upload failed"),
		},
	}
	uploader := metrics.NewDataClientUploader(
		mockClient,
		"part-1",
		"rdk:component:sensor",
		"metricscapture",
		"DoCommandFlush",
		[]string{"metrics"},
	)
	svc := metrics.NewService(metrics.Config{
		MaxQueueSize: 10,
		BatchSize:    2,
	}, uploader)

	svc.Emit("brain.do_command_total", 1, nil)
	svc.Emit("brain.auto_mode_loop_total", 1, nil)
	svc.Emit("brain.move_cancelled_total", 1, nil)

	res, err := svc.FlushNow(context.Background())
	if err == nil {
		t.Fatalf("expected flush error, got success")
	}
	if res.UploadedCount != 0 {
		t.Fatalf("expected uploaded_count=0 on failure, got %d", res.UploadedCount)
	}

	stats := svc.Stats()
	if stats.QueueDepth != 3 {
		t.Fatalf("expected queue_depth=3 after requeue, got %d", stats.QueueDepth)
	}
	if stats.UploadFailCount != 1 {
		t.Fatalf("expected upload_fail_count=1, got %d", stats.UploadFailCount)
	}
	if stats.UploadedCount != 0 {
		t.Fatalf("expected uploaded_count=0 in stats, got %d", stats.UploadedCount)
	}
	if stats.LastUploadError == "" {
		t.Fatalf("expected last_upload_error to be set")
	}
	if mockClient.Calls() != 1 {
		t.Fatalf("expected 1 upload call before failure, got %d", mockClient.Calls())
	}
}
