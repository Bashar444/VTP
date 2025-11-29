/**
 * QualitySelector Component
 * Allows users to select, override, and manage quality profiles
 */

import React, { useState, useEffect } from 'react';
import g5Service, { QualityProfile } from '@/services/g5Service';
import './QualitySelector.css';

interface QualitySelectorProps {
  onProfileChanged?: (profile: QualityProfile) => void;
}

const QualitySelectorComponent: React.FC<QualitySelectorProps> = ({ onProfileChanged }) => {
  const [profiles, setProfiles] = useState<QualityProfile[]>([]);
  const [currentProfile, setCurrentProfile] = useState<QualityProfile | null>(null);
  const [selectedProfile, setSelectedProfile] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [isChanging, setIsChanging] = useState<boolean>(false);

  // Get quality level icon
  const getQualityIcon = (level: string): string => {
    switch (level?.toLowerCase()) {
      case 'ultra_hd':
      case 'ultrahd':
        return '📺';
      case 'high_def':
      case 'highdef':
      case 'hd':
        return '📹';
      case 'standard':
      case 'std':
        return '📷';
      case 'medium':
      case 'med':
        return '📱';
      case 'low':
        return '📊';
      default:
        return '⚙️';
    }
  };

  // Get quality color
  const getQualityColor = (level: string): string => {
    switch (level?.toLowerCase()) {
      case 'ultra_hd':
      case 'ultrahd':
        return '#00ff00';
      case 'high_def':
      case 'highdef':
      case 'hd':
        return '#00dd00';
      case 'standard':
      case 'std':
        return '#ffaa00';
      case 'medium':
      case 'med':
        return '#ff6600';
      case 'low':
        return '#ff0000';
      default:
        return '#999999';
    }
  };

  // Fetch quality profiles and current selection
  const fetchProfiles = async () => {
    try {
      setError(null);
      const [profilesData, currentData] = await Promise.all([
        g5Service.getQualityProfiles(),
        g5Service.getCurrentQualityProfile(),
      ]);

      setProfiles(profilesData);
      setCurrentProfile(currentData);
      setSelectedProfile(currentData.name);
    } catch (err) {
      console.error('Error fetching quality profiles:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch quality profiles');
    } finally {
      setLoading(false);
    }
  };

  // Initial fetch
  useEffect(() => {
    fetchProfiles();
  }, []);

  // Handle profile selection
  const handleProfileSelect = async (profileName: string) => {
    if (profileName === currentProfile?.name) {
      return; // No change
    }

    setIsChanging(true);
    try {
      setError(null);
      const newProfile = await g5Service.setQualityProfile(profileName);
      setCurrentProfile(newProfile);
      setSelectedProfile(newProfile.name);
      
      if (onProfileChanged) {
        onProfileChanged(newProfile);
      }
    } catch (err) {
      console.error('Error setting quality profile:', err);
      setError(err instanceof Error ? err.message : 'Failed to set quality profile');
      setSelectedProfile(currentProfile?.name || null);
    } finally {
      setIsChanging(false);
    }
  };

  if (loading) {
    return <div className="quality-selector-container loading">Loading quality profiles...</div>;
  }

  return (
    <div className="quality-selector-container">
      <div className="selector-header">
        <h2>Video Quality</h2>
        <button onClick={fetchProfiles} className="refresh-btn" disabled={loading || isChanging}>
          ⟲ Refresh
        </button>
      </div>

      {error && <div className="error-message">{error}</div>}

      {/* Current Profile Display */}
      <div className="current-profile-section">
        <h3>Current Profile</h3>
        <div className="current-profile">
          <div className="profile-icon">{getQualityIcon(currentProfile?.level || '')}</div>
          <div className="profile-info">
            <div className="profile-name">{currentProfile?.name || 'Unknown'}</div>
            <div className="profile-details">
              <span>{currentProfile?.resolution || 'N/A'}</span>
              <span>•</span>
              <span>{currentProfile?.codec || 'N/A'}</span>
            </div>
          </div>
          <div className="profile-requirements">
            <div className="requirement">
              <label>Min Bandwidth:</label>
              <value>{currentProfile?.min_bandwidth || 0} Mbps</value>
            </div>
            <div className="requirement">
              <label>Max Latency:</label>
              <value>{currentProfile?.min_latency || 0} ms</value>
            </div>
          </div>
        </div>
      </div>

      {/* Quality Profiles Selection */}
      <div className="profiles-section">
        <h3>Available Profiles</h3>
        <div className="profiles-grid">
          {profiles.map((profile) => (
            <button
              key={profile.name}
              className={`profile-card ${selectedProfile === profile.name ? 'active' : ''} ${isChanging ? 'disabled' : ''}`}
              onClick={() => handleProfileSelect(profile.name)}
              disabled={isChanging}
              style={{
                borderColor: getQualityColor(profile.level),
              }}
            >
              <div className="profile-card-icon">{getQualityIcon(profile.level)}</div>
              <div className="profile-card-name">{profile.name}</div>
              <div className="profile-card-resolution">{profile.resolution}</div>
              <div className="profile-card-codec">{profile.codec}</div>
              <div className="profile-card-requirement">
                {profile.min_bandwidth} Mbps • {profile.min_latency} ms
              </div>
              {selectedProfile === profile.name && (
                <div className="profile-card-active-badge">✓ Active</div>
              )}
            </button>
          ))}
        </div>
      </div>

      {/* Profile Comparison */}
      <div className="comparison-section">
        <h3>Profile Comparison</h3>
        <div className="comparison-table">
          <div className="comparison-header">
            <div className="comparison-col profile-col">Profile</div>
            <div className="comparison-col resolution-col">Resolution</div>
            <div className="comparison-col codec-col">Codec</div>
            <div className="comparison-col bandwidth-col">Min BW</div>
            <div className="comparison-col latency-col">Max Latency</div>
          </div>
          {profiles.map((profile) => (
            <div
              key={profile.name}
              className={`comparison-row ${selectedProfile === profile.name ? 'active' : ''}`}
            >
              <div className="comparison-col profile-col">
                <span className="profile-icon">{getQualityIcon(profile.level)}</span>
                <span>{profile.name}</span>
              </div>
              <div className="comparison-col resolution-col">{profile.resolution}</div>
              <div className="comparison-col codec-col">{profile.codec}</div>
              <div className="comparison-col bandwidth-col">{profile.min_bandwidth} Mbps</div>
              <div className="comparison-col latency-col">{profile.min_latency} ms</div>
            </div>
          ))}
        </div>
      </div>

      {/* Auto-Adjust Recommendation */}
      <div className="recommendation-section">
        <h3>Smart Selection</h3>
        <p className="recommendation-text">
          The system automatically recommends the best quality profile based on your current network conditions.
          You can manually override this selection using the profile cards above.
        </p>
        <button className="auto-select-btn" onClick={() => fetchProfiles()}>
          🤖 Let AI Choose Best Quality
        </button>
      </div>
    </div>
  );
};

export default QualitySelectorComponent;
