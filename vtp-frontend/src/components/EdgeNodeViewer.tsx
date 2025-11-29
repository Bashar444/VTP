/**
 * EdgeNodeViewer Component
 * Displays available edge nodes, their regions, latency, capacity, and health status
 */

import React, { useState, useEffect } from 'react';
import g5Service, { EdgeNode } from '@/services/g5Service';
import './EdgeNodeViewer.css';

interface EdgeNodeViewerProps {
  autoRefresh?: boolean;
  refreshInterval?: number;
}

const EdgeNodeViewerComponent: React.FC<EdgeNodeViewerProps> = ({
  autoRefresh = true,
  refreshInterval = 10000,
}) => {
  const [nodes, setNodes] = useState<EdgeNode[]>([]);
  const [closestNode, setClosestNode] = useState<EdgeNode | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  const [sortBy, setSortBy] = useState<'latency' | 'capacity' | 'region'>('latency');

  // Get status color
  const getStatusColor = (status: string): string => {
    switch (status?.toLowerCase()) {
      case 'online':
        return '#00ff00';
      case 'offline':
        return '#ff0000';
      case 'degraded':
        return '#ffaa00';
      case 'maintenance':
        return '#ff6600';
      default:
        return '#999999';
    }
  };

  // Get status badge
  const getStatusBadge = (status: string): string => {
    switch (status?.toLowerCase()) {
      case 'online':
        return '🟢';
      case 'offline':
        return '🔴';
      case 'degraded':
        return '🟡';
      case 'maintenance':
        return '🔧';
      default:
        return '⚪';
    }
  };

  // Get region emoji
  const getRegionEmoji = (region: string): string => {
    const regionMap: { [key: string]: string } = {
      'us-east': '🗽',
      'us-west': '🏔️',
      'us-central': '🌾',
      'us-south': '⛱️',
      'eu-west': '🏰',
      'eu-central': '🕌',
      'asia-southeast': '🏯',
      'asia-east': '🏮',
      'asia-south': '🕍',
      'australia': '🦘',
      'canada': '🍁',
      'sa-east': '🦙',
    };

    const normalizedRegion = region?.toLowerCase() || '';
    return regionMap[normalizedRegion] || '🌍';
  };

  // Calculate capacity usage percentage
  const getCapacityUsage = (node: EdgeNode): number => {
    if (node.capacity === 0) return 0;
    return Math.round(((node.capacity - node.available) / node.capacity) * 100);
  };

  // Sort nodes
  const getSortedNodes = (): EdgeNode[] => {
    const sorted = [...nodes];
    switch (sortBy) {
      case 'latency':
        return sorted.sort((a, b) => (a.latency || 0) - (b.latency || 0));
      case 'capacity':
        return sorted.sort((a, b) => getCapacityUsage(b) - getCapacityUsage(a));
      case 'region':
        return sorted.sort((a, b) => a.region.localeCompare(b.region));
      default:
        return sorted;
    }
  };

  // Fetch data
  const fetchData = async () => {
    try {
      setError(null);
      const [allNodes, closest] = await Promise.all([
        g5Service.getAvailableEdgeNodes(),
        g5Service.getClosestEdgeNode().catch(() => null),
      ]);

      setNodes(allNodes);
      if (closest) {
        setClosestNode(closest);
        setSelectedNode(closest.id);
      }
    } catch (err) {
      console.error('Error fetching edge nodes:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch edge nodes');
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
    if (!autoRefresh) return;
    const interval = setInterval(fetchData, refreshInterval);
    return () => clearInterval(interval);
  }, [autoRefresh, refreshInterval]);

  if (loading && nodes.length === 0) {
    return <div className="edge-node-viewer loading">Loading edge nodes...</div>;
  }

  const sortedNodes = getSortedNodes();

  return (
    <div className="edge-node-viewer">
      <div className="viewer-header">
        <h2>Edge Nodes</h2>
        <div className="header-controls">
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as any)}
            className="sort-select"
          >
            <option value="latency">Sort by Latency</option>
            <option value="capacity">Sort by Capacity</option>
            <option value="region">Sort by Region</option>
          </select>
          <button onClick={fetchData} className="refresh-btn" disabled={loading}>
            {loading ? '⟳ Loading...' : '⟲ Refresh'}
          </button>
        </div>
      </div>

      {error && <div className="error-message">{error}</div>}

      {/* Closest Node Highlight */}
      {closestNode && (
        <div className="closest-node-section">
          <h3>🎯 Recommended Node</h3>
          <div className="closest-node-card">
            <div className="node-header">
              <div className="node-title">
                <span className="region-emoji">{getRegionEmoji(closestNode.region)}</span>
                <div className="node-info">
                  <div className="node-name">{closestNode.id}</div>
                  <div className="node-region">{closestNode.region}</div>
                </div>
              </div>
              <div className="node-status">
                <span className="status-badge">{getStatusBadge(closestNode.status)}</span>
                <span className="status-text">{closestNode.status}</span>
              </div>
            </div>
            <div className="node-metrics">
              <div className="metric">
                <label>Latency</label>
                <value>{closestNode.latency?.toFixed(1) || 'N/A'} ms</value>
              </div>
              <div className="metric">
                <label>Capacity</label>
                <value>{closestNode.capacity} cores</value>
              </div>
              <div className="metric">
                <label>Available</label>
                <value>{closestNode.available} cores</value>
              </div>
              <div className="metric">
                <label>Usage</label>
                <value>{getCapacityUsage(closestNode)}%</value>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* All Nodes List */}
      <div className="nodes-section">
        <h3>All Nodes ({nodes.length})</h3>
        <div className="nodes-list">
          {sortedNodes.length === 0 ? (
            <div className="no-nodes">No edge nodes available</div>
          ) : (
            sortedNodes.map((node) => (
              <div
                key={node.id}
                className={`node-card ${selectedNode === node.id ? 'selected' : ''} ${node.status?.toLowerCase() === 'offline' ? 'offline' : ''}`}
                onClick={() => setSelectedNode(node.id)}
              >
                <div className="node-card-header">
                  <div className="node-card-title">
                    <span className="region-emoji">{getRegionEmoji(node.region)}</span>
                    <div>
                      <div className="node-card-name">{node.id}</div>
                      <div className="node-card-region">{node.region}</div>
                    </div>
                  </div>
                  <div className="node-card-status">
                    <span className="status-badge">{getStatusBadge(node.status)}</span>
                  </div>
                </div>

                <div className="node-card-metrics">
                  <div className="metric-group">
                    <label>Latency</label>
                    <value className="latency-value">{node.latency?.toFixed(1) || 'N/A'} ms</value>
                  </div>
                  <div className="metric-group">
                    <label>Capacity</label>
                    <value>{node.capacity}</value>
                  </div>
                  <div className="metric-group">
                    <label>Available</label>
                    <value className="available-value">{node.available}</value>
                  </div>
                </div>

                <div className="capacity-bar">
                  <div className="capacity-bar-fill" style={{ width: `${getCapacityUsage(node)}%` }}></div>
                </div>
                <div className="capacity-text">
                  {getCapacityUsage(node)}% Used ({node.capacity - node.available}/{node.capacity} cores)
                </div>

                {selectedNode === node.id && (
                  <div className="node-card-badge">Selected</div>
                )}
              </div>
            ))
          )}
        </div>
      </div>

      {/* Node Statistics */}
      <div className="statistics-section">
        <h3>Statistics</h3>
        <div className="stats-grid">
          <div className="stat-card">
            <div className="stat-label">Total Nodes</div>
            <div className="stat-value">{nodes.length}</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Online Nodes</div>
            <div className="stat-value" style={{ color: '#00ff00' }}>
              {nodes.filter((n) => n.status?.toLowerCase() === 'online').length}
            </div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Offline Nodes</div>
            <div className="stat-value" style={{ color: '#ff6666' }}>
              {nodes.filter((n) => n.status?.toLowerCase() === 'offline').length}
            </div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Avg Latency</div>
            <div className="stat-value">
              {nodes.length > 0
                ? (nodes.reduce((sum, n) => sum + (n.latency || 0), 0) / nodes.length).toFixed(1)
                : 'N/A'} ms
            </div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Total Capacity</div>
            <div className="stat-value">
              {nodes.reduce((sum, n) => sum + n.capacity, 0)}
            </div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Total Available</div>
            <div className="stat-value" style={{ color: '#66ff66' }}>
              {nodes.reduce((sum, n) => sum + n.available, 0)}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default EdgeNodeViewerComponent;
