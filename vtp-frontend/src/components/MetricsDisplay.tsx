/**
 * MetricsDisplay Component
 * Displays real-time 5G metrics including session metrics, global metrics, and historical trends
 * Features:
 * - Real-time metric updates with auto-refresh
 * - Session metrics: latency, bandwidth, packet loss, codec, resolution, quality level
 * - Global metrics: average latency, total bandwidth, network quality, user count
 * - Responsive grid layout
 * - Loading states and error handling
 */

import React, { useState, useEffect } from 'react';
import { g5Service, SessionMetrics, GlobalMetrics } from '../services/g5Service';
import './MetricsDisplay.css';

interface MetricsDisplayProps {
  refreshInterval?: number; // in milliseconds, default 5000
  onMetricsUpdate?: (session: SessionMetrics, global: GlobalMetrics) => void;
}

interface MetricTrend {
  value: number;
  timestamp: number;
  trend?: 'up' | 'down' | 'stable';
}

const MetricsDisplay: React.FC<MetricsDisplayProps> = ({
  refreshInterval = 5000,
  onMetricsUpdate
}) => {
  // State
  const [sessionMetrics, setSessionMetrics] = useState<SessionMetrics | null>(null);
  const [globalMetrics, setGlobalMetrics] = useState<GlobalMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());
  
  // Trend tracking for metrics
  const [latencyTrend, setLatencyTrend] = useState<MetricTrend[]>([]);
  const [bandwidthTrend, setBandwidthTrend] = useState<MetricTrend[]>([]);
  const [qualityTrend, setQualityTrend] = useState<MetricTrend[]>([]);

  // Fetch metrics
  const fetchMetrics = async () => {
    try {
      setError(null);
      const [sessionData, globalData] = await Promise.all([
        g5Service.getSessionMetrics(),
        g5Service.getGlobalMetrics()
      ]);

      setSessionMetrics(sessionData);
      setGlobalMetrics(globalData);
      setLastUpdate(new Date());

      // Update trends
      if (sessionData) {
        setLatencyTrend(prev => {
          const updated = [...prev, {
            value: sessionData.latency,
            timestamp: Date.now()
          }];
          return updated.slice(-20); // Keep last 20 readings
        });

        setBandwidthTrend(prev => {
          const updated = [...prev, {
            value: sessionData.bandwidth,
            timestamp: Date.now()
          }];
          return updated.slice(-20);
        });

        setQualityTrend(prev => {
          const updated = [...prev, {
            value: sessionData.qualityLevel,
            timestamp: Date.now()
          }];
          return updated.slice(-20);
        });
      }

      if (onMetricsUpdate && sessionData && globalData) {
        onMetricsUpdate(sessionData, globalData);
      }

      setLoading(false);
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to fetch metrics';
      setError(errorMsg);
      setLoading(false);
    }
  };

  // Auto-refresh effect
  useEffect(() => {
    fetchMetrics();
    const interval = setInterval(fetchMetrics, refreshInterval);
    return () => clearInterval(interval);
  }, [refreshInterval]);

  // Helper functions
  const getTrendIcon = (current: number, previous: number | undefined): string => {
    if (!previous) return '→';
    if (current > previous) return '↑';
    if (current < previous) return '↓';
    return '→';
  };

  const getTrendColor = (current: number, previous: number | undefined, isGood: boolean): string => {
    if (!previous) return '#888';
    const trendUp = current > previous;
    return (isGood && trendUp) || (!isGood && !trendUp) ? '#66ff66' : '#ff6666';
  };

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
  };

  const formatBandwidth = (value: number): string => {
    if (value >= 1000) {
      return (value / 1000).toFixed(1) + ' Mbps';
    }
    return value.toFixed(1) + ' Kbps';
  };

  const getMetricClass = (value: number, thresholds: { good: number; warning: number }, isReverse?: boolean): string => {
    if (isReverse) {
      if (value <= thresholds.good) return 'metric-good';
      if (value <= thresholds.warning) return 'metric-warning';
      return 'metric-poor';
    } else {
      if (value >= thresholds.good) return 'metric-good';
      if (value >= thresholds.warning) return 'metric-warning';
      return 'metric-poor';
    }
  };

  if (loading && !sessionMetrics) {
    return <div className="metrics-display loading">Loading metrics...</div>;
  }

  const prevLatency = latencyTrend.length > 1 ? latencyTrend[latencyTrend.length - 2]?.value : undefined;
  const prevBandwidth = bandwidthTrend.length > 1 ? bandwidthTrend[bandwidthTrend.length - 2]?.value : undefined;
  const prevQuality = qualityTrend.length > 1 ? qualityTrend[qualityTrend.length - 2]?.value : undefined;

  return (
    <div className="metrics-display">
      {error && (
        <div className="error-message">
          ⚠️ {error}
        </div>
      )}

      {/* Session Metrics Section */}
      <div className="metrics-section session-metrics">
        <div className="section-header">
          <h3>📊 Session Metrics</h3>
          <div className="update-info">
            Last updated: {lastUpdate.toLocaleTimeString()}
          </div>
        </div>

        {sessionMetrics ? (
          <div className="metrics-grid">
            {/* Latency */}
            <div className={`metric-card ${getMetricClass(sessionMetrics.latency, { good: 50, warning: 100 }, true)}`}>
              <div className="metric-header">
                <span className="metric-label">Latency</span>
                <span className="metric-trend" style={{ color: getTrendColor(sessionMetrics.latency, prevLatency, false) }}>
                  {getTrendIcon(sessionMetrics.latency, prevLatency)}
                </span>
              </div>
              <div className="metric-value">{sessionMetrics.latency.toFixed(1)}ms</div>
              <div className="metric-subtext">Lower is better</div>
              <div className="metric-bar">
                <div className="metric-bar-fill" style={{ width: `${Math.min(sessionMetrics.latency / 2, 100)}%` }}></div>
              </div>
            </div>

            {/* Bandwidth */}
            <div className={`metric-card ${getMetricClass(sessionMetrics.bandwidth, { good: 5000, warning: 2000 })}`}>
              <div className="metric-header">
                <span className="metric-label">Bandwidth</span>
                <span className="metric-trend" style={{ color: getTrendColor(sessionMetrics.bandwidth, prevBandwidth, true) }}>
                  {getTrendIcon(sessionMetrics.bandwidth, prevBandwidth)}
                </span>
              </div>
              <div className="metric-value">{formatBandwidth(sessionMetrics.bandwidth)}</div>
              <div className="metric-subtext">Current throughput</div>
              <div className="metric-bar">
                <div className="metric-bar-fill" style={{ width: `${Math.min((sessionMetrics.bandwidth / 10000) * 100, 100)}%` }}></div>
              </div>
            </div>

            {/* Packet Loss */}
            <div className={`metric-card ${getMetricClass(sessionMetrics.packetLoss, { good: 1, warning: 5 }, true)}`}>
              <div className="metric-header">
                <span className="metric-label">Packet Loss</span>
                <span className="metric-icon">📦</span>
              </div>
              <div className="metric-value">{sessionMetrics.packetLoss.toFixed(2)}%</div>
              <div className="metric-subtext">Loss rate</div>
              <div className="metric-bar">
                <div className="metric-bar-fill" style={{ width: `${Math.min(sessionMetrics.packetLoss * 10, 100)}%`, background: sessionMetrics.packetLoss > 5 ? '#ff6666' : '#66ff66' }}></div>
              </div>
            </div>

            {/* Quality Level */}
            <div className={`metric-card ${getMetricClass(sessionMetrics.qualityLevel, { good: 80, warning: 60 })}`}>
              <div className="metric-header">
                <span className="metric-label">Quality Level</span>
                <span className="metric-trend" style={{ color: getTrendColor(sessionMetrics.qualityLevel, prevQuality, true) }}>
                  {getTrendIcon(sessionMetrics.qualityLevel, prevQuality)}
                </span>
              </div>
              <div className="metric-value">{sessionMetrics.qualityLevel.toFixed(0)}%</div>
              <div className="metric-subtext">Overall quality</div>
              <div className="metric-bar">
                <div className="metric-bar-fill" style={{ width: `${sessionMetrics.qualityLevel}%` }}></div>
              </div>
            </div>

            {/* Codec */}
            <div className="metric-card">
              <div className="metric-header">
                <span className="metric-label">Codec</span>
                <span className="metric-icon">🎞️</span>
              </div>
              <div className="metric-value">{sessionMetrics.codec}</div>
              <div className="metric-subtext">Video codec</div>
            </div>

            {/* Resolution */}
            <div className="metric-card">
              <div className="metric-header">
                <span className="metric-label">Resolution</span>
                <span className="metric-icon">📺</span>
              </div>
              <div className="metric-value">{sessionMetrics.resolution}</div>
              <div className="metric-subtext">Current resolution</div>
            </div>
          </div>
        ) : (
          <div className="no-data">No session metrics available</div>
        )}
      </div>

      {/* Global Metrics Section */}
      <div className="metrics-section global-metrics">
        <div className="section-header">
          <h3>🌍 Global Metrics</h3>
        </div>

        {globalMetrics ? (
          <div className="global-grid">
            <div className="global-card">
              <div className="global-label">Average Latency</div>
              <div className="global-value">{globalMetrics.averageLatency.toFixed(1)}ms</div>
              <div className="global-subtext">Across all sessions</div>
            </div>

            <div className="global-card">
              <div className="global-label">Total Bandwidth</div>
              <div className="global-value">{formatBandwidth(globalMetrics.totalBandwidth)}</div>
              <div className="global-subtext">Network capacity</div>
            </div>

            <div className="global-card">
              <div className="global-label">Network Quality</div>
              <div className="global-value">{globalMetrics.networkQuality.toFixed(0)}%</div>
              <div className="global-subtext">Overall health</div>
            </div>

            <div className="global-card">
              <div className="global-label">Active Users</div>
              <div className="global-value">{globalMetrics.activeUsers}</div>
              <div className="global-subtext">Current sessions</div>
            </div>

            <div className="global-card">
              <div className="global-label">Peak Bandwidth</div>
              <div className="global-value">{formatBandwidth(globalMetrics.peakBandwidth)}</div>
              <div className="global-subtext">Maximum observed</div>
            </div>

            <div className="global-card">
              <div className="global-label">Uptime</div>
              <div className="global-value">{(globalMetrics.uptime / 3600).toFixed(1)}h</div>
              <div className="global-subtext">Service availability</div>
            </div>
          </div>
        ) : (
          <div className="no-data">No global metrics available</div>
        )}
      </div>

      {/* Charts Section */}
      <div className="metrics-section charts-section">
        <div className="section-header">
          <h3>📈 Trends</h3>
        </div>

        <div className="charts-grid">
          {/* Latency Trend */}
          <div className="chart-card">
            <div className="chart-title">Latency Trend</div>
            <div className="chart-container">
              <div className="trend-chart">
                <div className="chart-area">
                  {latencyTrend.length > 1 && (
                    <svg viewBox="0 0 300 100" preserveAspectRatio="none" className="sparkline">
                      <polyline
                        points={latencyTrend
                          .map((t, i) => {
                            const x = (i / (latencyTrend.length - 1)) * 300;
                            const y = 100 - Math.min((t.value / 200) * 100, 100);
                            return `${x},${y}`;
                          })
                          .join(' ')}
                        fill="none"
                        stroke="url(#gradient)"
                        strokeWidth="2"
                      />
                      <defs>
                        <linearGradient id="gradient" x1="0%" y1="0%" x2="0%" y2="100%">
                          <stop offset="0%" stopColor="#66ff66" stopOpacity="0.8" />
                          <stop offset="100%" stopColor="#66ff66" stopOpacity="0.2" />
                        </linearGradient>
                      </defs>
                    </svg>
                  )}
                </div>
              </div>
              <div className="trend-stat">
                Current: <strong>{sessionMetrics?.latency.toFixed(1)}ms</strong>
              </div>
            </div>
          </div>

          {/* Bandwidth Trend */}
          <div className="chart-card">
            <div className="chart-title">Bandwidth Trend</div>
            <div className="chart-container">
              <div className="trend-chart">
                <div className="chart-area">
                  {bandwidthTrend.length > 1 && (
                    <svg viewBox="0 0 300 100" preserveAspectRatio="none" className="sparkline">
                      <polyline
                        points={bandwidthTrend
                          .map((t, i) => {
                            const x = (i / (bandwidthTrend.length - 1)) * 300;
                            const y = 100 - Math.min((t.value / 10000) * 100, 100);
                            return `${x},${y}`;
                          })
                          .join(' ')}
                        fill="none"
                        stroke="url(#gradientBw)"
                        strokeWidth="2"
                      />
                      <defs>
                        <linearGradient id="gradientBw" x1="0%" y1="0%" x2="0%" y2="100%">
                          <stop offset="0%" stopColor="#ffaa00" stopOpacity="0.8" />
                          <stop offset="100%" stopColor="#ffaa00" stopOpacity="0.2" />
                        </linearGradient>
                      </defs>
                    </svg>
                  )}
                </div>
              </div>
              <div className="trend-stat">
                Current: <strong>{formatBandwidth(sessionMetrics?.bandwidth || 0)}</strong>
              </div>
            </div>
          </div>

          {/* Quality Trend */}
          <div className="chart-card">
            <div className="chart-title">Quality Trend</div>
            <div className="chart-container">
              <div className="trend-chart">
                <div className="chart-area">
                  {qualityTrend.length > 1 && (
                    <svg viewBox="0 0 300 100" preserveAspectRatio="none" className="sparkline">
                      <polyline
                        points={qualityTrend
                          .map((t, i) => {
                            const x = (i / (qualityTrend.length - 1)) * 300;
                            const y = 100 - t.value;
                            return `${x},${y}`;
                          })
                          .join(' ')}
                        fill="none"
                        stroke="url(#gradientQl)"
                        strokeWidth="2"
                      />
                      <defs>
                        <linearGradient id="gradientQl" x1="0%" y1="0%" x2="0%" y2="100%">
                          <stop offset="0%" stopColor="#667eea" stopOpacity="0.8" />
                          <stop offset="100%" stopColor="#667eea" stopOpacity="0.2" />
                        </linearGradient>
                      </defs>
                    </svg>
                  )}
                </div>
              </div>
              <div className="trend-stat">
                Current: <strong>{sessionMetrics?.qualityLevel.toFixed(0)}%</strong>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Refresh Info */}
      <div className="metrics-footer">
        <span>Auto-refresh: Every {(refreshInterval / 1000).toFixed(0)}s</span>
        <button onClick={fetchMetrics} disabled={loading} className="refresh-metrics-btn">
          {loading ? 'Updating...' : '🔄 Refresh Now'}
        </button>
      </div>
    </div>
  );
};

export default MetricsDisplay;
