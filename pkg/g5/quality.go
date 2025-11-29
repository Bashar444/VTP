package g5

import (
	"errors"
	"sync"
	"time"
)

// QualitySelector handles streaming quality profile selection
type QualitySelector struct {
	mu                    sync.RWMutex
	currentProfile        QualityLevel
	availableProfiles     map[QualityLevel]*QualityProfile
	networkQuality        int
	adaptiveStrategy      *AdaptiveStrategy
	adjustmentHistory     []QualityAdjustment
	maxHistorySize        int
	lastAdjustmentTime    time.Time
	minAdjustmentInterval time.Duration
}

// NewQualitySelector creates a new quality selector
func NewQualitySelector(strategy *AdaptiveStrategy) *QualitySelector {
	selector := &QualitySelector{
		currentProfile:        QualityAuto,
		adaptiveStrategy:      strategy,
		adjustmentHistory:     make([]QualityAdjustment, 0),
		maxHistorySize:        100,
		minAdjustmentInterval: 2 * time.Second,
		availableProfiles:     initializeQualityProfiles(),
	}

	return selector
}

// initializeQualityProfiles initializes standard quality profiles
func initializeQualityProfiles() map[QualityLevel]*QualityProfile {
	return map[QualityLevel]*QualityProfile{
		QualityUltraHD: {
			Name:         "Ultra HD (4K)",
			Bitrate:      15000, // 15 Mbps
			Resolution:   "3840x2160",
			FPS:          60,
			Codec:        "VP9/H.265",
			MaxLatency:   20,
			MinBandwidth: 12000,
		},
		QualityHighDef: {
			Name:         "High Definition (1440p)",
			Bitrate:      8000, // 8 Mbps
			Resolution:   "2560x1440",
			FPS:          30,
			Codec:        "VP9/H.265",
			MaxLatency:   50,
			MinBandwidth: 6000,
		},
		QualityStandard: {
			Name:         "Standard (1080p)",
			Bitrate:      4000, // 4 Mbps
			Resolution:   "1920x1080",
			FPS:          30,
			Codec:        "VP8/H.264",
			MaxLatency:   100,
			MinBandwidth: 3000,
		},
		QualityMedium: {
			Name:         "Medium (720p)",
			Bitrate:      2000, // 2 Mbps
			Resolution:   "1280x720",
			FPS:          24,
			Codec:        "VP8/H.264",
			MaxLatency:   150,
			MinBandwidth: 1500,
		},
		QualityLow: {
			Name:         "Low (480p)",
			Bitrate:      800, // 800 Kbps
			Resolution:   "854x480",
			FPS:          24,
			Codec:        "VP8/H.264",
			MaxLatency:   200,
			MinBandwidth: 600,
		},
	}
}

// SelectQuality selects the best quality profile for current network conditions
func (qs *QualitySelector) SelectQuality(latency int, bandwidth int) (QualityLevel, error) {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	// Check minimum adjustment interval
	if !qs.lastAdjustmentTime.IsZero() &&
		time.Since(qs.lastAdjustmentTime) < qs.minAdjustmentInterval {
		return qs.currentProfile, nil
	}

	if qs.adaptiveStrategy == nil || !qs.adaptiveStrategy.Enabled {
		return qs.currentProfile, nil
	}

	selectedProfile := qs.selectBestProfile(latency, bandwidth)

	// Record adjustment if profile changed
	if selectedProfile != qs.currentProfile {
		adjustment := QualityAdjustment{
			FromProfile: qs.currentProfile,
			ToProfile:   selectedProfile,
			Reason:      qs.getAdjustmentReason(latency, bandwidth),
			Timestamp:   time.Now(),
		}

		qs.adjustmentHistory = append(qs.adjustmentHistory, adjustment)
		if len(qs.adjustmentHistory) > qs.maxHistorySize {
			qs.adjustmentHistory = qs.adjustmentHistory[1:]
		}

		qs.currentProfile = selectedProfile
		qs.lastAdjustmentTime = time.Now()
	}

	return selectedProfile, nil
}

// selectBestProfile selects the best quality profile based on network conditions
func (qs *QualitySelector) selectBestProfile(latency int, bandwidth int) QualityLevel {
	// Order from highest to lowest quality
	profileOrder := []QualityLevel{
		QualityUltraHD,
		QualityHighDef,
		QualityStandard,
		QualityMedium,
		QualityLow,
	}

	for _, profile := range profileOrder {
		p := qs.availableProfiles[profile]
		if p == nil {
			continue
		}

		// Check if network can support this profile
		if latency <= p.MaxLatency && bandwidth >= p.MinBandwidth {
			return profile
		}
	}

	// Fallback to lowest quality
	return QualityLow
}

// getAdjustmentReason returns the reason for quality adjustment
func (qs *QualitySelector) getAdjustmentReason(latency int, bandwidth int) string {
	if bandwidth < 600 {
		return "insufficient bandwidth"
	}
	if latency > 200 {
		return "high latency"
	}
	if bandwidth > 12000 && latency < 20 {
		return "excellent network conditions"
	}
	if bandwidth > 6000 && latency < 50 {
		return "good network conditions"
	}
	if bandwidth < 3000 || latency > 100 {
		return "poor network conditions"
	}
	return "network condition change"
}

// GetCurrentProfile returns the current quality profile
func (qs *QualitySelector) GetCurrentProfile() *QualityProfile {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	if profile, ok := qs.availableProfiles[qs.currentProfile]; ok {
		// Return a copy
		p := *profile
		return &p
	}
	return nil
}

// GetCurrentProfileLevel returns the current quality level
func (qs *QualitySelector) GetCurrentProfileLevel() QualityLevel {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	return qs.currentProfile
}

// SetProfile manually sets the quality profile
func (qs *QualitySelector) SetProfile(level QualityLevel) error {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	if _, ok := qs.availableProfiles[level]; !ok {
		return errors.New("invalid quality level")
	}

	if qs.currentProfile != level {
		adjustment := QualityAdjustment{
			FromProfile: qs.currentProfile,
			ToProfile:   level,
			Reason:      "manual selection",
			Timestamp:   time.Now(),
		}

		qs.adjustmentHistory = append(qs.adjustmentHistory, adjustment)
		if len(qs.adjustmentHistory) > qs.maxHistorySize {
			qs.adjustmentHistory = qs.adjustmentHistory[1:]
		}

		qs.currentProfile = level
		qs.lastAdjustmentTime = time.Now()
	}

	return nil
}

// GetAvailableProfiles returns all available quality profiles
func (qs *QualitySelector) GetAvailableProfiles() map[QualityLevel]*QualityProfile {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	// Return a copy
	profiles := make(map[QualityLevel]*QualityProfile)
	for level, profile := range qs.availableProfiles {
		p := *profile
		profiles[level] = &p
	}
	return profiles
}

// GetAdjustmentHistory returns the quality adjustment history
func (qs *QualitySelector) GetAdjustmentHistory() []QualityAdjustment {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	// Return a copy
	history := make([]QualityAdjustment, len(qs.adjustmentHistory))
	copy(history, qs.adjustmentHistory)
	return history
}

// ClearAdjustmentHistory clears the adjustment history
func (qs *QualitySelector) ClearAdjustmentHistory() {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.adjustmentHistory = make([]QualityAdjustment, 0)
}

// AddProfile adds a custom quality profile
func (qs *QualitySelector) AddProfile(level QualityLevel, profile *QualityProfile) error {
	if profile == nil {
		return errors.New("profile cannot be nil")
	}

	qs.mu.Lock()
	defer qs.mu.Unlock()

	qs.availableProfiles[level] = profile
	return nil
}

// RemoveProfile removes a quality profile
func (qs *QualitySelector) RemoveProfile(level QualityLevel) error {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	if qs.currentProfile == level {
		return errors.New("cannot remove currently active profile")
	}

	delete(qs.availableProfiles, level)
	return nil
}

// GetRecommendedProfile returns the recommended quality profile for given conditions
func (qs *QualitySelector) GetRecommendedProfile(latency int, bandwidth int) *QualityProfile {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	recommended := qs.selectBestProfile(latency, bandwidth)
	if profile, ok := qs.availableProfiles[recommended]; ok {
		p := *profile
		return &p
	}
	return nil
}

// CanSwitchToProfile checks if switching to a profile is possible with current network
func (qs *QualitySelector) CanSwitchToProfile(level QualityLevel, latency int, bandwidth int) bool {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	profile, ok := qs.availableProfiles[level]
	if !ok {
		return false
	}

	return latency <= profile.MaxLatency && bandwidth >= profile.MinBandwidth
}

// GetStatistics returns quality adjustment statistics
func (qs *QualitySelector) GetStatistics() map[string]interface{} {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	// Count adjustments by type
	upgrades := 0
	downgrades := 0

	profileOrder := map[QualityLevel]int{
		QualityUltraHD:  5,
		QualityHighDef:  4,
		QualityStandard: 3,
		QualityMedium:   2,
		QualityLow:      1,
	}

	for _, adj := range qs.adjustmentHistory {
		fromScore := profileOrder[adj.FromProfile]
		toScore := profileOrder[adj.ToProfile]
		if toScore > fromScore {
			upgrades++
		} else if toScore < fromScore {
			downgrades++
		}
	}

	return map[string]interface{}{
		"currentProfile":        qs.currentProfile,
		"totalAdjustments":      len(qs.adjustmentHistory),
		"qualityUpgrades":       upgrades,
		"qualityDowngrades":     downgrades,
		"lastAdjustmentTime":    qs.lastAdjustmentTime,
		"availableProfileCount": len(qs.availableProfiles),
	}
}
