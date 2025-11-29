/**
 * 5G Service Integration
 * Communicates with Go backend API endpoints for 5G module functionality
 */

import axios, { AxiosInstance } from 'axios';

// API Response Types
export interface NetworkStatus {
  type: string;
  latency: number;
  bandwidth: number;
  connected: boolean;
  signal_strength?: number;
}

export interface AdapterStatus {
  is_started: boolean;
  is_healthy: boolean;
  active_session_id: string;
  current_network?: NetworkStatus;
  last_status_update: string;
}

export interface EdgeNode {
  id: string;
  region: string;
  status: string;
  capacity: number;
  available: number;
  latency?: number;
}

export interface QualityProfile {
  name: string;
  level: string;
  min_bandwidth: number;
  min_latency: number;
  codec: string;
  resolution: string;
}

export interface SessionMetrics {
  session_id: string;
  start_time: string;
  duration: number;
  samples_count: number;
  average_latency: number;
  average_bandwidth: number;
  quality_level: string;
  codec: string;
}

export interface GlobalMetrics {
  total_sessions: number;
  active_sessions: number;
  average_quality: number;
  total_bytes_transferred: number;
  uptime_seconds: number;
}

export interface DetectionResult {
  detected_type: string;
  quality_score: number;
  latency: number;
  bandwidth: number;
  signal_strength: number;
}

/**
 * 5G Service - REST API client for backend communication
 */
class G5Service {
  private api: AxiosInstance;
  private baseURL: string;

  constructor(baseURL: string = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1/5g') {
    this.baseURL = baseURL;
    this.api = axios.create({
      baseURL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add response interceptor for error handling
    this.api.interceptors.response.use(
      (response) => response,
      (error) => {
        console.error('5G Service API Error:', error);
        return Promise.reject(error);
      }
    );
  }

  /**
   * Get current adapter status
   */
  async getStatus(): Promise<AdapterStatus> {
    try {
      const response = await this.api.get<AdapterStatus>('/status');
      return response.data;
    } catch (error) {
      console.error('Failed to get adapter status:', error);
      throw error;
    }
  }

  /**
   * Get current network information
   */
  async getCurrentNetwork(): Promise<NetworkStatus> {
    try {
      const response = await this.api.get<NetworkStatus>('/network/current');
      return response.data;
    } catch (error) {
      console.error('Failed to get current network:', error);
      throw error;
    }
  }

  /**
   * Get current network quality (0-100)
   */
  async getNetworkQuality(): Promise<number> {
    try {
      const response = await this.api.get<{ quality: number }>('/network/quality');
      return response.data.quality;
    } catch (error) {
      console.error('Failed to get network quality:', error);
      throw error;
    }
  }

  /**
   * Check if 5G is available
   */
  async is5GAvailable(): Promise<boolean> {
    try {
      const response = await this.api.get<{ available: boolean }>('/network/5g-available');
      return response.data.available;
    } catch (error) {
      console.error('Failed to check 5G availability:', error);
      throw error;
    }
  }

  /**
   * Get all available edge nodes
   */
  async getAvailableEdgeNodes(): Promise<EdgeNode[]> {
    try {
      const response = await this.api.get<EdgeNode[]>('/edge/nodes');
      return response.data;
    } catch (error) {
      console.error('Failed to get edge nodes:', error);
      throw error;
    }
  }

  /**
   * Get the closest edge node based on latency
   */
  async getClosestEdgeNode(): Promise<EdgeNode> {
    try {
      const response = await this.api.get<EdgeNode>('/edge/closest');
      return response.data;
    } catch (error) {
      console.error('Failed to get closest edge node:', error);
      throw error;
    }
  }

  /**
   * Get edge node by ID
   */
  async getEdgeNode(nodeId: string): Promise<EdgeNode> {
    try {
      const response = await this.api.get<EdgeNode>(`/edge/nodes/${nodeId}`);
      return response.data;
    } catch (error) {
      console.error(`Failed to get edge node ${nodeId}:`, error);
      throw error;
    }
  }

  /**
   * Get available quality profiles
   */
  async getQualityProfiles(): Promise<QualityProfile[]> {
    try {
      const response = await this.api.get<QualityProfile[]>('/quality/profiles');
      return response.data;
    } catch (error) {
      console.error('Failed to get quality profiles:', error);
      throw error;
    }
  }

  /**
   * Get current quality profile
   */
  async getCurrentQualityProfile(): Promise<QualityProfile> {
    try {
      const response = await this.api.get<QualityProfile>('/quality/current');
      return response.data;
    } catch (error) {
      console.error('Failed to get current quality profile:', error);
      throw error;
    }
  }

  /**
   * Set quality profile
   */
  async setQualityProfile(profileName: string): Promise<QualityProfile> {
    try {
      const response = await this.api.post<QualityProfile>('/quality/set', { profile: profileName });
      return response.data;
    } catch (error) {
      console.error('Failed to set quality profile:', error);
      throw error;
    }
  }

  /**
   * Get session metrics
   */
  async getSessionMetrics(sessionId?: string): Promise<SessionMetrics> {
    try {
      const response = await this.api.get<SessionMetrics>('/metrics/session', {
        params: sessionId ? { session_id: sessionId } : undefined,
      });
      return response.data;
    } catch (error) {
      console.error('Failed to get session metrics:', error);
      throw error;
    }
  }

  /**
   * Get global metrics
   */
  async getGlobalMetrics(): Promise<GlobalMetrics> {
    try {
      const response = await this.api.get<GlobalMetrics>('/metrics/global');
      return response.data;
    } catch (error) {
      console.error('Failed to get global metrics:', error);
      throw error;
    }
  }

  /**
   * Detect network type and quality
   */
  async detectNetwork(): Promise<DetectionResult> {
    try {
      const response = await this.api.post<DetectionResult>('/network/detect', {});
      return response.data;
    } catch (error) {
      console.error('Failed to detect network:', error);
      throw error;
    }
  }

  /**
   * Start a session
   */
  async startSession(sessionId: string): Promise<{ status: string; session_id: string }> {
    try {
      const response = await this.api.post<{ status: string; session_id: string }>('/session/start', {
        session_id: sessionId,
      });
      return response.data;
    } catch (error) {
      console.error('Failed to start session:', error);
      throw error;
    }
  }

  /**
   * End current session
   */
  async endSession(): Promise<{ status: string }> {
    try {
      const response = await this.api.post<{ status: string }>('/session/end', {});
      return response.data;
    } catch (error) {
      console.error('Failed to end session:', error);
      throw error;
    }
  }

  /**
   * Record metric for current session
   */
  async recordMetric(name: string, value: number | string): Promise<{ status: string }> {
    try {
      const response = await this.api.post<{ status: string }>('/metrics/record', {
        metric_name: name,
        value,
      });
      return response.data;
    } catch (error) {
      console.error('Failed to record metric:', error);
      throw error;
    }
  }

  /**
   * Get health check status
   */
  async healthCheck(): Promise<{ status: string; healthy: boolean }> {
    try {
      const response = await this.api.get<{ status: string; healthy: boolean }>('/health');
      return response.data;
    } catch (error) {
      console.error('Health check failed:', error);
      throw error;
    }
  }

  /**
   * Set API base URL (for dynamic endpoint switching)
   */
  setBaseURL(url: string): void {
    this.baseURL = url;
    this.api.defaults.baseURL = url;
  }

  /**
   * Get current base URL
   */
  getBaseURL(): string {
    return this.baseURL;
  }
}

// Export singleton instance
export default new G5Service();
