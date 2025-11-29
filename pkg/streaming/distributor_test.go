package streaming

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

// Distributor Tests

func TestNewLiveDistributor(t *testing.T) {
	distributor := NewLiveDistributor("rec-001", 50, 30*time.Second)

	if distributor == nil {
		t.Error("Expected non-nil distributor")
	}
	if distributor.RecordingID != "rec-001" {
		t.Errorf("Expected recording ID 'rec-001', got %s", distributor.RecordingID)
	}
	if distributor.maxViewers != 50 {
		t.Errorf("Expected max viewers 50, got %d", distributor.maxViewers)
	}
	if len(distributor.DistributionProfiles) != 4 {
		t.Errorf("Expected 4 profiles, got %d", len(distributor.DistributionProfiles))
	}
}

func TestEnqueueSegment(t *testing.T) {
	distributor := NewLiveDistributor("rec-001", 50, 30*time.Second)
	segment := &VideoSegment{
		SegmentID:   "seg-001",
		RecordingID: "rec-001",
		Bitrate:     "1000",
		SequenceNum: 1,
		DurationMs:  2000,
		FilePath:    "/tmp/seg-001.ts",
		FileSize:    1024000,
		CreatedTime: time.Now(),
		ExpiresTime: time.Now().Add(30 * time.Second),
		IsKeyFrame:  true,
	}

	err := distributor.EnqueueSegment(segment)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if distributor.segmentQueue.Length() != 1 {
		t.Errorf("Expected queue length 1, got %d", distributor.segmentQueue.Length())
	}
}

func TestJoinViewer(t *testing.T) {
	distributor := NewLiveDistributor("rec-001", 50, 30*time.Second)

	session, err := distributor.JoinViewer("viewer-001", "Medium")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if session == nil {
		t.Error("Expected non-nil session")
		return
	}
	if session.ViewerID != "viewer-001" {
		t.Errorf("Expected viewer ID 'viewer-001', got %s", session.ViewerID)
	}
	if session.CurrentBitrate != "Medium" {
		t.Errorf("Expected bitrate 'Medium', got %s", session.CurrentBitrate)
	}

	sessions := distributor.GetAllViewerSessions()
	if len(sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessions))
	}
}

func TestLeaveViewer(t *testing.T) {
	distributor := NewLiveDistributor("rec-001", 50, 30*time.Second)

	_, _ = distributor.JoinViewer("viewer-001", "Low")
	sessions := distributor.GetAllViewerSessions()
	if len(sessions) != 1 {
		t.Errorf("Expected 1 session before leave, got %d", len(sessions))
	}

	err := distributor.LeaveViewer("viewer-001")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	sessions = distributor.GetAllViewerSessions()
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions after leave, got %d", len(sessions))
	}
}

func TestMaxViewersLimit(t *testing.T) {
	distributor := NewLiveDistributor("rec-001", 2, 30*time.Second)

	_, _ = distributor.JoinViewer("viewer-001", "Low")
	_, _ = distributor.JoinViewer("viewer-002", "Low")

	_, err := distributor.JoinViewer("viewer-003", "Low")
	if err == nil {
		t.Error("Expected error when max viewers reached")
	}
}

func TestDeliverSegment(t *testing.T) {
	distributor := NewLiveDistributor("rec-001", 50, 30*time.Second)

	session, _ := distributor.JoinViewer("viewer-001", "Medium")
	segment := &VideoSegment{
		SegmentID:   "seg-001",
		RecordingID: "rec-001",
		Bitrate:     "Medium",
		SequenceNum: 1,
		DurationMs:  2000,
		FilePath:    "/tmp/seg-001.ts",
		FileSize:    2048000,
		CreatedTime: time.Now(),
		ExpiresTime: time.Now().Add(30 * time.Second),
		IsKeyFrame:  true,
	}

	distributor.EnqueueSegment(segment)

	err := distributor.DeliverSegment("viewer-001", "seg-001")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if session.SegmentsReceived != 1 {
		t.Errorf("Expected 1 segment received, got %d", session.SegmentsReceived)
	}
	if session.BytesReceived != 2048000 {
		t.Errorf("Expected bytes received 2048000, got %d", session.BytesReceived)
	}
}

func TestSwitchBitrate(t *testing.T) {
	distributor := NewLiveDistributor("rec-001", 50, 30*time.Second)
	distributor.JoinViewer("viewer-001", "Low")

	err := distributor.SwitchBitrate("viewer-001", "High")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	session, _ := distributor.GetViewerSession("viewer-001")
	if session.CurrentBitrate != "High" {
		t.Errorf("Expected bitrate 'High', got %s", session.CurrentBitrate)
	}
}

func TestUpdateViewerBuffer(t *testing.T) {
	distributor := NewLiveDistributor("rec-001", 50, 30*time.Second)
	distributor.JoinViewer("viewer-001", "Low")

	err := distributor.UpdateViewerBuffer("viewer-001", 75.0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	session, _ := distributor.GetViewerSession("viewer-001")
	if session.BufferHealth != 75.0 {
		t.Errorf("Expected buffer health 75.0, got %.1f", session.BufferHealth)
	}
	if session.ConnectionQuality != "good" {
		t.Errorf("Expected connection quality 'good', got %s", session.ConnectionQuality)
	}
}

func TestGetNextSegment(t *testing.T) {
	distributor := NewLiveDistributor("rec-001", 50, 30*time.Second)
	distributor.JoinViewer("viewer-001", "Low")

	segment := &VideoSegment{
		SegmentID:   "seg-001",
		RecordingID: "rec-001",
		Bitrate:     "Low",
		SequenceNum: 1,
		DurationMs:  2000,
		FilePath:    "/tmp/seg-001.ts",
		FileSize:    1024000,
		CreatedTime: time.Now(),
		ExpiresTime: time.Now().Add(30 * time.Second),
		IsKeyFrame:  true,
	}

	distributor.EnqueueSegment(segment)

	nextSegment, err := distributor.GetNextSegment("viewer-001")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if nextSegment == nil {
		t.Error("Expected non-nil segment")
		return
	}
	if nextSegment.SegmentID != "seg-001" {
		t.Errorf("Expected segment ID 'seg-001', got %s", nextSegment.SegmentID)
	}
}

func TestGetDistributionStats(t *testing.T) {
	distributor := NewLiveDistributor("rec-001", 50, 30*time.Second)

	distributor.JoinViewer("viewer-001", "Low")
	distributor.JoinViewer("viewer-002", "Medium")

	stats := distributor.GetDistributionStats()

	if stats.ActiveViewers != 2 {
		t.Errorf("Expected 2 active viewers, got %d", stats.ActiveViewers)
	}
	if distributor.RecordingID != "rec-001" {
		t.Errorf("Expected recording ID 'rec-001', got %s", distributor.RecordingID)
	}
}

func TestDistributorClose(t *testing.T) {
	distributor := NewLiveDistributor("rec-001", 50, 30*time.Second)
	distributor.JoinViewer("viewer-001", "Low")

	err := distributor.Close()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !distributor.closed {
		t.Error("Expected distributor to be marked as closed")
	}

	sessions := distributor.GetAllViewerSessions()
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions after close, got %d", len(sessions))
	}
}

// Distribution Service Tests

func TestNewDistributionService(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	service := NewDistributionService(4, logger)

	if service == nil {
		t.Error("Expected non-nil service")
	}
	if service.workerCount != 4 {
		t.Errorf("Expected 4 workers, got %d", service.workerCount)
	}
}

func TestCreateDistributor(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	service := NewDistributionService(2, logger)
	defer service.Stop()

	distributor, err := service.CreateDistributor("rec-001", 50, 30*time.Second)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if distributor == nil {
		t.Error("Expected non-nil distributor")
	}
}

func TestStartLiveStream(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	service := NewDistributionService(2, logger)
	defer service.Stop()

	distributor, err := service.StartLiveStream("rec-001", 100)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if distributor == nil {
		t.Error("Expected non-nil distributor")
	}
}

func TestGetDistributor(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	service := NewDistributionService(2, logger)
	defer service.Stop()

	service.CreateDistributor("rec-001", 50, 30*time.Second)

	distributor, err := service.GetDistributor("rec-001")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if distributor == nil {
		t.Error("Expected non-nil distributor")
	}
}

func TestJoinStream(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	service := NewDistributionService(2, logger)
	defer service.Stop()

	service.StartLiveStream("rec-001", 50)

	session, err := service.JoinStream("rec-001", "viewer-001", "Low")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if session == nil {
		t.Error("Expected non-nil session")
	}
}

func TestLeaveStream(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	service := NewDistributionService(2, logger)
	defer service.Stop()

	service.StartLiveStream("rec-001", 50)
	service.JoinStream("rec-001", "viewer-001", "Low")

	err := service.LeaveStream("rec-001", "viewer-001")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestAdaptViewerQuality(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	service := NewDistributionService(2, logger)
	defer service.Stop()

	service.StartLiveStream("rec-001", 50)
	service.JoinStream("rec-001", "viewer-001", "Medium")

	// Simulate poor buffer - should downgrade
	bitrate, err := service.AdaptViewerQuality("rec-001", "viewer-001", 25.0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if bitrate != "Low" {
		t.Errorf("Expected downgrade to 'Low', got %s", bitrate)
	}
}

func TestGetStreamViewers(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	service := NewDistributionService(2, logger)
	defer service.Stop()

	service.StartLiveStream("rec-001", 50)
	service.JoinStream("rec-001", "viewer-001", "Low")
	service.JoinStream("rec-001", "viewer-002", "Medium")

	viewers, err := service.GetStreamViewers("rec-001")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(viewers) != 2 {
		t.Errorf("Expected 2 viewers, got %d", len(viewers))
	}
}

func TestGetStreamStatistics(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	service := NewDistributionService(2, logger)
	defer service.Stop()

	service.StartLiveStream("rec-001", 50)
	service.JoinStream("rec-001", "viewer-001", "Low")

	stats, err := service.GetStreamStatistics("rec-001")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if stats.ActiveViewers != 1 {
		t.Errorf("Expected 1 active viewer, got %d", stats.ActiveViewers)
	}
}

func TestEndLiveStream(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	service := NewDistributionService(2, logger)
	defer service.Stop()

	service.StartLiveStream("rec-001", 50)
	service.JoinStream("rec-001", "viewer-001", "Low")

	err := service.EndLiveStream("rec-001")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, err = service.GetDistributor("rec-001")
	if err == nil {
		t.Error("Expected error when getting closed stream")
	}
}

func TestEnableCDN(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	service := NewDistributionService(2, logger)
	defer service.Stop()

	service.EnableCDN("https://cdn.example.com")

	if !service.cdnEnabled {
		t.Error("Expected CDN to be enabled")
	}
	if service.cdnEndpoint != "https://cdn.example.com" {
		t.Errorf("Expected CDN endpoint 'https://cdn.example.com', got %s", service.cdnEndpoint)
	}
}

func TestGetMetrics(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	service := NewDistributionService(2, logger)
	defer service.Stop()

	service.StartLiveStream("rec-001", 50)
	service.JoinStream("rec-001", "viewer-001", "Low")

	metrics := service.GetMetrics()

	if metrics.TotalDistributors != 1 {
		t.Errorf("Expected 1 distributor, got %d", metrics.TotalDistributors)
	}
	if metrics.TotalActiveViewers != 1 {
		t.Errorf("Expected 1 active viewer, got %d", metrics.TotalActiveViewers)
	}
}

// Benchmark Tests

func BenchmarkJoinViewer(b *testing.B) {
	distributor := NewLiveDistributor("rec-001", 10000, 30*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = distributor.JoinViewer(fmt.Sprintf("viewer-%d", i%100), "Low")
	}
}

func BenchmarkEnqueueSegment(b *testing.B) {
	distributor := NewLiveDistributor("rec-001", 50, 30*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		segment := &VideoSegment{
			SegmentID:   fmt.Sprintf("seg-%d", i),
			RecordingID: "rec-001",
			Bitrate:     "Low",
			SequenceNum: int64(i),
			DurationMs:  2000,
			FilePath:    "/tmp/test.ts",
			FileSize:    1024000,
			CreatedTime: time.Now(),
			ExpiresTime: time.Now().Add(30 * time.Second),
			IsKeyFrame:  i%10 == 0,
		}
		_ = distributor.EnqueueSegment(segment)
	}
}
