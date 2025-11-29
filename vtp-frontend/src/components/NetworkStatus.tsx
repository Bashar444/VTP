/**
 * NetworkStatus Component
 * Displays current network type, signal quality, 5G availability, and network metrics
 */

import React, { useState, useEffect } from 'react';
import g5Service, { AdapterStatus, NetworkStatus } from '@/services/g5Service';
import './NetworkStatus.css';

interface NetworkStatusProps {
  refreshInterval?: number; // milliseconds, default 5000
}

const NetworkStatusComponent: React.FC<NetworkStatusProps> = ({ refreshInterval = 5000 }) => {
  const [status, setStatus] = useState<AdapterStatus | null>(null);
  const [networkInfo, setNetworkInfo] = useState<NetworkStatus | null>(null);
  const [quality, setQuality] = useState<number>(0);
  const [is5GAvailable, setIs5GAvailable] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());

  // Get network type color
  const getNetworkTypeColor = (type: string): string => {
    switch (type?.toLowerCase()) {
      case '5g':
        return '#00ff00'; // Green
      case '4g':
      case 'lte':
        return '#ffaa00'; // Orange
      case 'wifi':
        return '#0099ff'; // Blue
      default:
        return '#999999'; // Gray
    }
  };

  // Get quality indicator color
  const getQualityColor = (qualityScore: number): string => {
    if (qualityScore >= 80) return '#00ff00'; // Green - Excellent
    if (qualityScore >= 60) return '#ffaa00'; // Orange - Good
    if (qualityScore >= 40) return '#ff6600'; // Orange-red - Fair
    return '#ff0000'; // Red - Poor
  };

  // Get status indicator
  const getStatusIndicator = (isHealthy: boolean): React.ReactNode => {
    return (
      <div className={`status-indicator ${isHealthy ? 'healthy' : 'unhealthy'}`}>
        <span className="status-dot"></span>
        {isHealthy ? 'Healthy' : 'Unhealthy'}
      </div>
    );
  };

  // Fetch data
  const fetchData = async () => {
    try {
      setError(null);
      const [adapterStatus, netInfo, qualityScore, is5g] = await Promise.all([
        g5Service.getStatus(),
        g5Service.getCurrentNetwork(),
        g5Service.getNetworkQuality(),
        g5Service.is5GAvailable(),
      ]);

      setStatus(adapterStatus);
      setNetworkInfo(netInfo);
      setQuality(qualityScore);
      setIs5GAvailable(is5g);
      setLastUpdate(new Date());
    } catch (err) {
      console.error('Error fetching network status:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch network status');
    } finally {
      setLoading(false);
    }
  };

  // Initial fetch
  useEffect(() => {
    fetchData();
  }, []);

  // Set up refresh interval
  useEffect(() => {
    const interval = setInterval(fetchData, refreshInterval);
    return () => clearInterval(interval);
  }, [refreshInterval]);

  if (loading && !status) {
    return <div className="network-status-container loading">Loading network status...</div>;
  }

  return (
    <div className="network-status-container">
      <div className="network-status-header">
        <h2>Network Status</h2>
        <button onClick={fetchData} className="refresh-button" disabled={loading}>
          {loading ? 'Refreshing...' : 'Refresh'}
        </button>
      </div>

      {error && <div className="error-message">{error}</div>}

      <div className="network-status-grid">
        {/* Network Type Card */}
        <div className="status-card network-type-card">
          <div className="card-label">Network Type</div>
          <div className="network-type" style={{ borderColor: getNetworkTypeColor(networkInfo?.type || 'Unknown') }}>
            <span className="network-badge" style={{ backgroundColor: getNetworkTypeColor(networkInfo?.type || 'Unknown') }}>
              {networkInfo?.type || 'N/A'}
            </span>
          </div>
          <div className="card-subtext">{networkInfo?.connected ? 'Connected' : 'Disconnected'}</div>
        </div>

        {/* Quality Score Card */}
        <div className="status-card quality-card">
          <div className="card-label">Quality Score</div>
          <div className="quality-circle" style={{ borderColor: getQualityColor(quality) }}>
            <span className="quality-value">{quality}</span>
            <span className="quality-unit">%</span>
          </div>
          <div className="card-subtext quality-label">
            {quality >= 80 ? 'Excellent' : quality >= 60 ? 'Good' : quality >= 40 ? 'Fair' : 'Poor'}
          </div>
        </div>

        {/* 5G Availability Card */}
        <div className="status-card 5g-card">
          <div className="card-label">5G Available</div>
          <div className={`availability-badge ${is5GAvailable ? 'available' : 'unavailable'}`}>
            {is5GAvailable ? '✓ Yes' : '✗ No'}
          </div>
          <div className="card-subtext">
            {is5GAvailable ? '5G network detected' : 'Not on 5G network'}
          </div>
        </div>

        {/* Status Card */}
        <div className="status-card health-card">
          <div className="card-label">Service Status</div>
          {status && getStatusIndicator(status.is_healthy)}
          <div className="card-subtext">
            {status?.is_started ? 'Service running' : 'Service stopped'}
          </div>
        </div>
      </div>

      {/* Metrics Section */}
      <div className="metrics-section">
        <h3>Network Metrics</h3>
        <div className="metrics-grid">
          <div className="metric-item">
            <label>Latency</label>
            <value className="metric-value">{networkInfo?.latency?.toFixed(2) || 'N/A'} ms</value>
          </div>
          <div className="metric-item">
            <label>Bandwidth</label>
            <value className="metric-value">{networkInfo?.bandwidth?.toFixed(2) || 'N/A'} Mbps</value>
          </div>
          <div className="metric-item">
            <label>Signal Strength</label>
            <value className="metric-value">{networkInfo?.signal_strength?.toFixed(1) || 'N/A'} dBm</value>
          </div>
          <div className="metric-item">
            <label>Active Session</label>
            <value className="metric-value">{status?.active_session_id || 'None'}</value>
          </div>
        </div>
      </div>

      {/* Last Update */}
      <div className="last-update">
        Last updated: {lastUpdate.toLocaleTimeString()}
      </div>
    </div>
  );
};

export default NetworkStatusComponent;
