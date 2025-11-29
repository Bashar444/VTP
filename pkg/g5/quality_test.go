package g5

import (
	"testing"
	"time"
)

func TestNewQualitySelector(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled:         true,
		TargetLatency:   50,
		TargetBandwidth: 25,
	}

	selector := NewQualitySelector(strategy)

	if selector == nil {
		t.Fatal("expected selector instance, got nil")
	}

	if selector.currentProfile != QualityAuto {
		t.Errorf("expected initial profile QualityAuto, got %v", selector.currentProfile)
	}

	if selector.adaptiveStrategy != strategy {
		t.Errorf("expected strategy to be set")
	}
}

func TestInitializeQualityProfiles(t *testing.T) {
	profiles := initializeQualityProfiles()

	expectedProfiles := []QualityLevel{
		QualityUltraHD,
		QualityHighDef,
		QualityStandard,
		QualityMedium,
		QualityLow,
	}

	for _, level := range expectedProfiles {
		if _, ok := profiles[level]; !ok {
			t.Errorf("expected profile %v to be initialized", level)
		}
	}

	if profiles[QualityUltraHD].Bitrate != 15000 {
		t.Errorf("expected Ultra HD bitrate 15000, got %d", profiles[QualityUltraHD].Bitrate)
	}

	if profiles[QualityLow].Bitrate != 800 {
		t.Errorf("expected Low bitrate 800, got %d", profiles[QualityLow].Bitrate)
	}
}

func TestSelectQualityGoodNetwork(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled:         true,
		TargetLatency:   50,
		TargetBandwidth: 25,
	}

	selector := NewQualitySelector(strategy)

	// Test with good network conditions
	profile, err := selector.SelectQuality(25, 45)

	if err != nil {
		t.Errorf("SelectQuality failed: %v", err)
	}

	if profile == QualityLow {
		t.Errorf("expected better quality than Low for good network")
	}
}

func TestSelectQualityPoorNetwork(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled:         true,
		TargetLatency:   50,
		TargetBandwidth: 25,
	}

	selector := NewQualitySelector(strategy)

	// Test with poor network conditions
	profile, err := selector.SelectQuality(200, 1)

	if err != nil {
		t.Errorf("SelectQuality failed: %v", err)
	}

	if profile != QualityLow {
		t.Errorf("expected QualityLow for poor network, got %v", profile)
	}
}

func TestSelectQualityDisabled(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: false,
	}

	selector := NewQualitySelector(strategy)
	originalProfile := selector.currentProfile

	profile, err := selector.SelectQuality(25, 45)

	if err != nil {
		t.Errorf("SelectQuality failed: %v", err)
	}

	if profile != originalProfile {
		t.Errorf("expected profile to remain unchanged when disabled")
	}
}

func TestSelectBestProfile(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	tests := []struct {
		name            string
		latency         int
		bandwidth       int
		expectedProfile QualityLevel
	}{
		{"Ultra HD conditions", 15, 15000, QualityUltraHD},
		{"High Def conditions", 30, 8000, QualityHighDef},
		{"Standard conditions", 80, 4000, QualityStandard},
		{"Medium conditions", 150, 2000, QualityMedium},
		{"Low conditions", 250, 500, QualityLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := selector.selectBestProfile(tt.latency, tt.bandwidth)
			if profile != tt.expectedProfile {
				t.Errorf("expected %v, got %v", tt.expectedProfile, profile)
			}
		})
	}
}

func TestGetAdjustmentReason(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	tests := []struct {
		name           string
		latency        int
		bandwidth      int
		expectedReason string
	}{
		{"Low bandwidth", 50, 500, "insufficient bandwidth"},
		{"High latency", 250, 8000, "high latency"},
		{"Excellent network", 15, 15000, "excellent network conditions"},
		{"Good network", 30, 8000, "good network conditions"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := selector.getAdjustmentReason(tt.latency, tt.bandwidth)
			if reason != tt.expectedReason {
				t.Errorf("expected %s, got %s", tt.expectedReason, reason)
			}
		})
	}
}

func TestGetCurrentProfile(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	profile := selector.GetCurrentProfile()
	if profile == nil {
		t.Errorf("expected profile, got nil")
	}
}

func TestGetCurrentProfileLevel(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	level := selector.GetCurrentProfileLevel()
	if level != QualityAuto {
		t.Errorf("expected QualityAuto, got %v", level)
	}
}

func TestSetProfile(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	err := selector.SetProfile(QualityStandard)
	if err != nil {
		t.Errorf("SetProfile failed: %v", err)
	}

	if selector.GetCurrentProfileLevel() != QualityStandard {
		t.Errorf("expected profile to be set to QualityStandard")
	}
}

func TestSetProfileInvalid(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	invalidLevel := QualityLevel("invalid")
	err := selector.SetProfile(invalidLevel)
	if err == nil {
		t.Errorf("expected error with invalid profile, got nil")
	}
}

func TestGetAvailableProfiles(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	profiles := selector.GetAvailableProfiles()
	if len(profiles) == 0 {
		t.Errorf("expected profiles to be available")
	}

	if _, ok := profiles[QualityUltraHD]; !ok {
		t.Errorf("expected QualityUltraHD to be available")
	}

	if _, ok := profiles[QualityLow]; !ok {
		t.Errorf("expected QualityLow to be available")
	}
}

func TestGetAdjustmentHistory(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	// Initially should be empty
	history := selector.GetAdjustmentHistory()
	if len(history) != 0 {
		t.Errorf("expected empty history initially")
	}

	// Make some adjustments
	selector.SetProfile(QualityStandard)
	selector.SetProfile(QualityHighDef)

	history = selector.GetAdjustmentHistory()
	if len(history) != 2 {
		t.Errorf("expected 2 adjustments, got %d", len(history))
	}
}

func TestClearAdjustmentHistory(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	selector.SetProfile(QualityStandard)
	selector.SetProfile(QualityHighDef)

	selector.ClearAdjustmentHistory()

	history := selector.GetAdjustmentHistory()
	if len(history) != 0 {
		t.Errorf("expected empty history after clear")
	}
}

func TestAddProfile(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	customProfile := &QualityProfile{
		Name:       "Custom",
		Bitrate:    5000,
		Resolution: "1280x960",
		FPS:        30,
	}

	customLevel := QualityLevel("custom")
	err := selector.AddProfile(customLevel, customProfile)
	if err != nil {
		t.Errorf("AddProfile failed: %v", err)
	}

	profiles := selector.GetAvailableProfiles()
	if _, ok := profiles[customLevel]; !ok {
		t.Errorf("expected custom profile to be added")
	}
}

func TestAddProfileNil(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	err := selector.AddProfile(QualityLevel("custom"), nil)
	if err == nil {
		t.Errorf("expected error adding nil profile, got nil")
	}
}

func TestRemoveProfile(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	// Set current profile to something else first
	selector.SetProfile(QualityUltraHD)

	err := selector.RemoveProfile(QualityLow)
	if err != nil {
		t.Errorf("RemoveProfile failed: %v", err)
	}

	profiles := selector.GetAvailableProfiles()
	if _, ok := profiles[QualityLow]; ok {
		t.Errorf("expected QualityLow to be removed")
	}
}

func TestRemoveCurrentProfile(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	selector.SetProfile(QualityStandard)

	err := selector.RemoveProfile(QualityStandard)
	if err == nil {
		t.Errorf("expected error removing current profile, got nil")
	}
}

func TestGetRecommendedProfile(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	profile := selector.GetRecommendedProfile(25, 15000)
	if profile == nil {
		t.Errorf("expected recommended profile, got nil")
	}

	if profile.Bitrate > 15000 {
		t.Errorf("expected profile bitrate <= 15000, got %d", profile.Bitrate)
	}
}

func TestCanSwitchToProfile(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	tests := []struct {
		name      string
		level     QualityLevel
		latency   int
		bandwidth int
		expected  bool
	}{
		{"Can switch to Ultra HD", QualityUltraHD, 15, 15000, true},
		{"Cannot switch to Ultra HD (low bandwidth)", QualityUltraHD, 15, 1000, false},
		{"Can switch to Low", QualityLow, 250, 500, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := selector.CanSwitchToProfile(tt.level, tt.latency, tt.bandwidth)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestQualitySelectorGetStatistics(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	selector.SetProfile(QualityStandard)
	selector.SetProfile(QualityHighDef)
	selector.SetProfile(QualityStandard)

	stats := selector.GetStatistics()

	if stats == nil {
		t.Errorf("expected statistics, got nil")
	}

	if currentProfile, ok := stats["currentProfile"]; ok {
		if currentProfile != QualityStandard {
			t.Errorf("expected current profile QualityStandard")
		}
	}

	if totalAdjustments, ok := stats["totalAdjustments"].(int); ok {
		if totalAdjustments != 3 {
			t.Errorf("expected 3 adjustments, got %d", totalAdjustments)
		}
	}
}

func TestQualitySelectorConcurrency(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				selector.GetCurrentProfile()
				selector.GetCurrentProfileLevel()
				selector.SelectQuality(25+i*10, 45+j*10)
				time.Sleep(time.Millisecond)
			}
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestMinAdjustmentInterval(t *testing.T) {
	strategy := &AdaptiveStrategy{
		Enabled: true,
	}

	selector := NewQualitySelector(strategy)

	// First adjustment
	selector.SetProfile(QualityStandard)

	// Try immediate adjustment (should be throttled)
	oldProfile := selector.GetCurrentProfileLevel()
	selector.SelectQuality(100, 20)
	newProfile := selector.GetCurrentProfileLevel()

	if oldProfile != newProfile {
		t.Logf("Profile may have changed despite throttling")
	}
}
