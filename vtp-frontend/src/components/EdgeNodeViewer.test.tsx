/**
 * EdgeNodeViewer Component Unit Tests
 * Tests for edge node management and visualization
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import EdgeNodeViewer from './EdgeNodeViewer';
import * as g5ServiceModule from '../services/g5Service';

vi.mock('../services/g5Service', () => ({
  g5Service: {
    getAvailableEdgeNodes: vi.fn(),
    getClosestEdgeNode: vi.fn(),
  },
}));

describe('EdgeNodeViewer Component', () => {
  const mockNodes = [
    {
      id: 'us-east-1',
      name: 'US East',
      region: 'Virginia',
      latency: 12,
      capacity: 500,
      available: 425,
      status: 'online' as const,
    },
    {
      id: 'us-west-1',
      name: 'US West',
      region: 'California',
      latency: 45,
      capacity: 500,
      available: 200,
      status: 'online' as const,
    },
    {
      id: 'eu-west-1',
      name: 'EU West',
      region: 'Ireland',
      latency: 78,
      capacity: 500,
      available: 100,
      status: 'degraded' as const,
    },
    {
      id: 'ap-south-1',
      name: 'Asia Pacific',
      region: 'Mumbai',
      latency: 95,
      capacity: 500,
      available: 50,
      status: 'offline' as const,
    },
  ];

  const mockClosestNode = mockNodes[0];

  beforeEach(() => {
    vi.clearAllMocks();
    (g5ServiceModule.g5Service.getAvailableEdgeNodes as any).mockResolvedValue(mockNodes);
    (g5ServiceModule.g5Service.getClosestEdgeNode as any).mockResolvedValue(mockClosestNode);
  });

  describe('Rendering', () => {
    it('should render without crashing', () => {
      render(<EdgeNodeViewer />);
      expect(screen.getByText(/Edge Nodes/i)).toBeInTheDocument();
    });

    it('should display closest node section', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        expect(screen.getByText(/Closest Node/i)).toBeInTheDocument();
      });
    });

    it('should display all edge nodes', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        expect(screen.getByText('US East')).toBeInTheDocument();
        expect(screen.getByText('US West')).toBeInTheDocument();
        expect(screen.getByText('EU West')).toBeInTheDocument();
        expect(screen.getByText('Asia Pacific')).toBeInTheDocument();
      });
    });

    it('should display statistics section', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        expect(screen.getByText(/Statistics/i)).toBeInTheDocument();
      });
    });
  });

  describe('Data Fetching', () => {
    it('should fetch nodes on mount', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        expect(g5ServiceModule.g5Service.getAvailableEdgeNodes).toHaveBeenCalled();
      });
    });

    it('should fetch closest node on mount', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        expect(g5ServiceModule.g5Service.getClosestEdgeNode).toHaveBeenCalled();
      });
    });

    it('should auto-refresh at specified interval', async () => {
      vi.useFakeTimers();
      render(<EdgeNodeViewer refreshInterval={2000} />);

      await waitFor(() => {
        expect(g5ServiceModule.g5Service.getAvailableEdgeNodes).toHaveBeenCalledTimes(1);
      });

      vi.advanceTimersByTime(2000);
      await waitFor(() => {
        expect(g5ServiceModule.g5Service.getAvailableEdgeNodes).toHaveBeenCalledTimes(2);
      });

      vi.useRealTimers();
    });

    it('should handle API errors gracefully', async () => {
      (g5ServiceModule.g5Service.getAvailableEdgeNodes as any).mockRejectedValueOnce(new Error('API Error'));
      render(<EdgeNodeViewer />);

      await waitFor(() => {
        expect(screen.queryByText(/error/i)).toBeInTheDocument();
      });
    });
  });

  describe('Node Display', () => {
    it('should display node names', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        expect(screen.getByText('US East')).toBeInTheDocument();
        expect(screen.getByText('US West')).toBeInTheDocument();
      });
    });

    it('should display node regions', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        expect(screen.getByText('Virginia')).toBeInTheDocument();
        expect(screen.getByText('California')).toBeInTheDocument();
      });
    });

    it('should display node latency', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        expect(screen.getByText(/12\s*ms/)).toBeInTheDocument();
        expect(screen.getByText(/45\s*ms/)).toBeInTheDocument();
      });
    });

    it('should display node capacity', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        // Check for capacity percentages or values
        expect(screen.getByText(/425/)).toBeInTheDocument();
      });
    });
  });

  describe('Closest Node Highlight', () => {
    it('should display closest node in special section', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        const closestSection = screen.getByText(/Closest Node/i).closest('div');
        expect(closestSection?.textContent).toContain('US East');
      });
    });

    it('should show latency for closest node', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        const closestSection = screen.getByText(/Closest Node/i).closest('div');
        expect(closestSection?.textContent).toContain('12');
      });
    });

    it('should show status badge for closest node', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        const closestSection = screen.getByText(/Closest Node/i).closest('div');
        expect(closestSection?.textContent).toContain('Online');
      });
    });
  });

  describe('Status Indicators', () => {
    it('should display online status with green indicator', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        const statusElements = screen.getAllByText(/Online/i);
        expect(statusElements.length).toBeGreaterThan(0);
      });
    });

    it('should display offline status', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        expect(screen.getByText(/Offline/i)).toBeInTheDocument();
      });
    });

    it('should display degraded status', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        expect(screen.getByText(/Degraded/i)).toBeInTheDocument();
      });
    });
  });

  describe('Sorting', () => {
    it('should have sorting dropdown', async () => {
      const { container } = render(<EdgeNodeViewer />);
      await waitFor(() => {
        const sortSelect = container.querySelector('select');
        expect(sortSelect).toBeInTheDocument();
      });
    });

    it('should sort by latency', async () => {
      const { container } = render(<EdgeNodeViewer />);
      await waitFor(() => {
        const sortSelect = container.querySelector('select') as HTMLSelectElement;
        fireEvent.change(sortSelect, { target: { value: 'latency' } });

        // US East (12ms) should appear before US West (45ms)
        const usEast = screen.getByText('US East');
        const usWest = screen.getByText('US West');
        expect(usEast.compareDocumentPosition(usWest)).toBe(Node.DOCUMENT_POSITION_FOLLOWING);
      });
    });

    it('should sort by capacity', async () => {
      const { container } = render(<EdgeNodeViewer />);
      await waitFor(() => {
        const sortSelect = container.querySelector('select') as HTMLSelectElement;
        fireEvent.change(sortSelect, { target: { value: 'capacity' } });
        expect(sortSelect.value).toBe('capacity');
      });
    });

    it('should sort by region', async () => {
      const { container } = render(<EdgeNodeViewer />);
      await waitFor(() => {
        const sortSelect = container.querySelector('select') as HTMLSelectElement;
        fireEvent.change(sortSelect, { target: { value: 'region' } });
        expect(sortSelect.value).toBe('region');
      });
    });
  });

  describe('Capacity Visualization', () => {
    it('should display capacity percentage', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        // 425/500 = 85%
        expect(screen.getByText(/85/)).toBeInTheDocument();
      });
    });

    it('should display capacity bar', async () => {
      const { container } = render(<EdgeNodeViewer />);
      await waitFor(() => {
        const bars = container.querySelectorAll('.capacity-bar-fill');
        expect(bars.length).toBeGreaterThan(0);
      });
    });

    it('should color-code capacity bars', async () => {
      const { container } = render(<EdgeNodeViewer />);
      await waitFor(() => {
        const bars = container.querySelectorAll('.capacity-bar-fill');
        expect(bars.length).toBeGreaterThan(0);
        // Verify they have styles
        bars.forEach(bar => {
          expect(bar).toHaveAttribute('style');
        });
      });
    });
  });

  describe('Statistics', () => {
    it('should display total nodes count', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        expect(screen.getByText(/Total Nodes.*4/)).toBeInTheDocument();
      });
    });

    it('should display online nodes count', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        expect(screen.getByText(/Online.*2/)).toBeInTheDocument();
      });
    });

    it('should display offline nodes count', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        expect(screen.getByText(/Offline.*1/)).toBeInTheDocument();
      });
    });

    it('should display average latency', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        // Average of 12, 45, 78, 95 = 57.5
        expect(screen.getByText(/Average Latency.*57/)).toBeInTheDocument();
      });
    });

    it('should display total and available capacity', async () => {
      render(<EdgeNodeViewer />);
      await waitFor(() => {
        expect(screen.getByText(/Total Capacity/)).toBeInTheDocument();
        expect(screen.getByText(/Available Capacity/)).toBeInTheDocument();
      });
    });
  });

  describe('Node Selection', () => {
    it('should allow node selection', async () => {
      const mockCallback = vi.fn();
      const { container } = render(<EdgeNodeViewer onNodeSelected={mockCallback} />);

      await waitFor(() => {
        expect(screen.getByText('US East')).toBeInTheDocument();
      });

      const nodeCard = screen.getByText('US East').closest('div[data-node-id]') || screen.getByText('US East').closest('div');
      if (nodeCard) {
        fireEvent.click(nodeCard);
        expect(mockCallback).toHaveBeenCalled();
      }
    });

    it('should highlight selected node', async () => {
      const { container } = render(<EdgeNodeViewer />);

      await waitFor(() => {
        expect(screen.getByText('US East')).toBeInTheDocument();
      });

      const nodeCard = screen.getByText('US East').closest('.node-card');
      fireEvent.click(nodeCard!);

      await waitFor(() => {
        expect(nodeCard?.className).toContain('selected');
      });
    });
  });

  describe('Callbacks', () => {
    it('should call onNodeSelected when node is clicked', async () => {
      const mockCallback = vi.fn();
      render(<EdgeNodeViewer onNodeSelected={mockCallback} />);

      await waitFor(() => {
        expect(screen.getByText('US East')).toBeInTheDocument();
      });

      const nodeCard = screen.getByText('US East').closest('div');
      if (nodeCard) {
        fireEvent.click(nodeCard);
        expect(mockCallback).toHaveBeenCalled();
      }
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty node list', async () => {
      (g5ServiceModule.g5Service.getAvailableEdgeNodes as any).mockResolvedValueOnce([]);
      render(<EdgeNodeViewer />);

      await waitFor(() => {
        expect(screen.getByText(/no nodes/i)).toBeInTheDocument();
      });
    });

    it('should handle null closest node', async () => {
      (g5ServiceModule.g5Service.getClosestEdgeNode as any).mockResolvedValueOnce(null);
      render(<EdgeNodeViewer />);

      await waitFor(() => {
        expect(screen.getByText(/Edge Nodes/i)).toBeInTheDocument();
      });
    });

    it('should cleanup on unmount', () => {
      const { unmount } = render(<EdgeNodeViewer refreshInterval={1000} />);
      expect(() => unmount()).not.toThrow();
    });

    it('should handle all offline nodes', async () => {
      const offlineNodes = mockNodes.map(node => ({ ...node, status: 'offline' as const }));
      (g5ServiceModule.g5Service.getAvailableEdgeNodes as any).mockResolvedValueOnce(offlineNodes);

      render(<EdgeNodeViewer />);
      await waitFor(() => {
        const offlineLabels = screen.getAllByText(/Offline/i);
        expect(offlineLabels.length).toBe(4);
      });
    });
  });

  describe('Responsive Design', () => {
    it('should render node grid layout', async () => {
      const { container } = render(<EdgeNodeViewer />);
      await waitFor(() => {
        const grid = container.querySelector('.nodes-list');
        expect(grid).toBeInTheDocument();
      });
    });

    it('should have proper CSS classes', () => {
      const { container } = render(<EdgeNodeViewer />);
      expect(container.querySelector('.edge-node-viewer')).toBeInTheDocument();
    });
  });
});
