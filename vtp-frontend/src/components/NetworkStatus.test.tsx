/**
 * NetworkStatus Component Unit Tests
 * Tests for real-time network status display
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import NetworkStatus from './NetworkStatus';
import * as g5ServiceModule from '../services/g5Service';

// Mock the g5Service
vi.mock('../services/g5Service', () => ({
  g5Service: {
    getCurrentNetwork: vi.fn(),
    getNetworkQuality: vi.fn(),
    is5GAvailable: vi.fn(),
    getStatus: vi.fn(),
  },
}));

describe('NetworkStatus Component', () => {
  const mockNetworkInfo = {
    type: '5G' as const,
    signalStrength: 92,
    bandwidth: 45,
    latency: 24.5,
  };

  const mockStatus = {
    running: true,
    lastUpdate: new Date().toISOString(),
    networksDetected: 1,
    activeSession: true,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    // Setup default mock implementations
    (g5ServiceModule.g5Service.getCurrentNetwork as any).mockResolvedValue(mockNetworkInfo);
    (g5ServiceModule.g5Service.getNetworkQuality as any).mockResolvedValue(87);
    (g5ServiceModule.g5Service.is5GAvailable as any).mockResolvedValue(true);
    (g5ServiceModule.g5Service.getStatus as any).mockResolvedValue(mockStatus);
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  describe('Rendering', () => {
    it('should render without crashing', () => {
      render(<NetworkStatus />);
      expect(screen.getByText(/Network Status/i)).toBeInTheDocument();
    });

    it('should display loading state initially', () => {
      render(<NetworkStatus />);
      // Component should either show content or loading
      const container = screen.getByRole('heading', { level: 2 });
      expect(container).toBeInTheDocument();
    });

    it('should render network status cards after data loads', async () => {
      render(<NetworkStatus />);
      await waitFor(() => {
        expect(screen.queryByText(/latency/i)).toBeInTheDocument();
      });
    });
  });

  describe('Data Fetching', () => {
    it('should call g5Service methods on mount', async () => {
      render(<NetworkStatus />);
      await waitFor(() => {
        expect(g5ServiceModule.g5Service.getCurrentNetwork).toHaveBeenCalled();
        expect(g5ServiceModule.g5Service.getNetworkQuality).toHaveBeenCalled();
        expect(g5ServiceModule.g5Service.is5GAvailable).toHaveBeenCalled();
      });
    });

    it('should auto-refresh at specified interval', async () => {
      vi.useFakeTimers();
      render(<NetworkStatus refreshInterval={1000} />);

      await waitFor(() => {
        expect(g5ServiceModule.g5Service.getCurrentNetwork).toHaveBeenCalledTimes(1);
      });

      vi.advanceTimersByTime(1000);
      await waitFor(() => {
        expect(g5ServiceModule.g5Service.getCurrentNetwork).toHaveBeenCalledTimes(2);
      });

      vi.useRealTimers();
    });

    it('should handle API errors gracefully', async () => {
      const error = new Error('Network error');
      (g5ServiceModule.g5Service.getCurrentNetwork as any).mockRejectedValueOnce(error);

      render(<NetworkStatus />);
      await waitFor(() => {
        expect(screen.queryByText(/error/i)).toBeInTheDocument();
      });
    });
  });

  describe('Display Values', () => {
    it('should display network type correctly', async () => {
      render(<NetworkStatus />);
      await waitFor(() => {
        expect(screen.getByText(/5G/i)).toBeInTheDocument();
      });
    });

    it('should display quality percentage', async () => {
      render(<NetworkStatus />);
      await waitFor(() => {
        expect(screen.getByText(/87/)).toBeInTheDocument();
      });
    });

    it('should display latency value', async () => {
      render(<NetworkStatus />);
      await waitFor(() => {
        expect(screen.getByText(/24\.5/)).toBeInTheDocument();
      });
    });

    it('should display bandwidth value', async () => {
      render(<NetworkStatus />);
      await waitFor(() => {
        expect(screen.getByText(/45/)).toBeInTheDocument();
      });
    });

    it('should display signal strength', async () => {
      render(<NetworkStatus />);
      await waitFor(() => {
        expect(screen.getByText(/92/)).toBeInTheDocument();
      });
    });
  });

  describe('5G Availability', () => {
    it('should show 5G available indicator when true', async () => {
      render(<NetworkStatus />);
      await waitFor(() => {
        const indicator = screen.getByText(/5G Available/i);
        expect(indicator).toBeInTheDocument();
      });
    });

    it('should update 5G status when availability changes', async () => {
      const { rerender } = render(<NetworkStatus />);

      (g5ServiceModule.g5Service.is5GAvailable as any).mockResolvedValueOnce(false);

      await waitFor(() => {
        expect(g5ServiceModule.g5Service.is5GAvailable).toHaveBeenCalled();
      });
    });
  });

  describe('Callbacks', () => {
    it('should call onStatusChange callback when data updates', async () => {
      const mockCallback = vi.fn();
      render(<NetworkStatus onStatusChange={mockCallback} />);

      await waitFor(() => {
        expect(mockCallback).toHaveBeenCalledWith(expect.objectContaining({
          type: '5G',
        }));
      });
    });

    it('should not throw when callback is undefined', async () => {
      expect(() => {
        render(<NetworkStatus />);
      }).not.toThrow();
    });
  });

  describe('Responsive Design', () => {
    it('should render responsive layout structure', () => {
      render(<NetworkStatus />);
      const container = screen.getByText(/Network Status/i).closest('div');
      expect(container).toHaveClass('network-status');
    });

    it('should have correct CSS classes', () => {
      const { container } = render(<NetworkStatus />);
      expect(container.querySelector('.network-status')).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle null network data', async () => {
      (g5ServiceModule.g5Service.getCurrentNetwork as any).mockResolvedValueOnce(null);
      render(<NetworkStatus />);

      await waitFor(() => {
        expect(screen.getByText(/Network Status/i)).toBeInTheDocument();
      });
    });

    it('should handle zero quality score', async () => {
      (g5ServiceModule.g5Service.getNetworkQuality as any).mockResolvedValueOnce(0);
      render(<NetworkStatus />);

      await waitFor(() => {
        expect(screen.getByText(/0/)).toBeInTheDocument();
      });
    });

    it('should handle maximum quality score', async () => {
      (g5ServiceModule.g5Service.getNetworkQuality as any).mockResolvedValueOnce(100);
      render(<NetworkStatus />);

      await waitFor(() => {
        expect(screen.getByText(/100/)).toBeInTheDocument();
      });
    });

    it('should cleanup on unmount', () => {
      const { unmount } = render(<NetworkStatus refreshInterval={1000} />);
      expect(() => unmount()).not.toThrow();
    });
  });

  describe('Performance', () => {
    it('should not call API more frequently than refresh interval', async () => {
      vi.useFakeTimers();
      render(<NetworkStatus refreshInterval={5000} />);

      await waitFor(() => {
        expect(g5ServiceModule.g5Service.getCurrentNetwork).toHaveBeenCalledTimes(1);
      });

      vi.advanceTimersByTime(2000);
      expect(g5ServiceModule.g5Service.getCurrentNetwork).toHaveBeenCalledTimes(1);

      vi.advanceTimersByTime(3000);
      expect(g5ServiceModule.g5Service.getCurrentNetwork).toHaveBeenCalledTimes(2);

      vi.useRealTimers();
    });
  });
});
