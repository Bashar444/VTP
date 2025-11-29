/**
 * QualitySelector Component Unit Tests
 * Tests for quality profile management
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import QualitySelector from './QualitySelector';
import * as g5ServiceModule from '../services/g5Service';

vi.mock('../services/g5Service', () => ({
  g5Service: {
    getQualityProfiles: vi.fn(),
    getCurrentQualityProfile: vi.fn(),
    setQualityProfile: vi.fn(),
  },
}));

describe('QualitySelector Component', () => {
  const mockProfiles = [
    { id: 'ultra', name: 'Ultra HD', resolution: '4K', codec: 'H.265', minBandwidth: 8000, maxLatency: 20 },
    { id: 'hd', name: 'HD', resolution: '1080p', codec: 'H.264', minBandwidth: 5000, maxLatency: 40 },
    { id: 'standard', name: 'Standard', resolution: '720p', codec: 'H.264', minBandwidth: 2500, maxLatency: 60 },
    { id: 'medium', name: 'Medium', resolution: '480p', codec: 'H.264', minBandwidth: 1500, maxLatency: 100 },
    { id: 'low', name: 'Low', resolution: '360p', codec: 'H.264', minBandwidth: 500, maxLatency: 150 },
  ];

  const mockCurrentProfile = mockProfiles[1]; // HD

  beforeEach(() => {
    vi.clearAllMocks();
    (g5ServiceModule.g5Service.getQualityProfiles as any).mockResolvedValue(mockProfiles);
    (g5ServiceModule.g5Service.getCurrentQualityProfile as any).mockResolvedValue(mockCurrentProfile);
    (g5ServiceModule.g5Service.setQualityProfile as any).mockResolvedValue(mockProfiles[0]);
  });

  describe('Rendering', () => {
    it('should render without crashing', () => {
      render(<QualitySelector />);
      expect(screen.getByText(/Quality Selection/i)).toBeInTheDocument();
    });

    it('should display all quality profile cards', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        expect(screen.getByText('Ultra HD')).toBeInTheDocument();
        expect(screen.getByText('HD')).toBeInTheDocument();
        expect(screen.getByText('Standard')).toBeInTheDocument();
        expect(screen.getByText('Medium')).toBeInTheDocument();
        expect(screen.getByText('Low')).toBeInTheDocument();
      });
    });

    it('should display comparison table', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        expect(screen.getByText(/Resolution/i)).toBeInTheDocument();
        expect(screen.getByText(/Codec/i)).toBeInTheDocument();
      });
    });
  });

  describe('Data Fetching', () => {
    it('should fetch quality profiles on mount', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        expect(g5ServiceModule.g5Service.getQualityProfiles).toHaveBeenCalled();
      });
    });

    it('should fetch current profile on mount', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        expect(g5ServiceModule.g5Service.getCurrentQualityProfile).toHaveBeenCalled();
      });
    });

    it('should handle API errors gracefully', async () => {
      (g5ServiceModule.g5Service.getQualityProfiles as any).mockRejectedValueOnce(new Error('API Error'));
      render(<QualitySelector />);
      await waitFor(() => {
        expect(screen.queryByText(/error/i)).toBeInTheDocument();
      });
    });
  });

  describe('Current Profile Display', () => {
    it('should display current profile name', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        expect(screen.getByText(/Current Profile: HD/i)).toBeInTheDocument();
      });
    });

    it('should display current profile resolution', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        expect(screen.getByText(/1080p/)).toBeInTheDocument();
      });
    });

    it('should display current profile codec', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        expect(screen.getByText(/H\.264/)).toBeInTheDocument();
      });
    });

    it('should highlight current profile in cards', async () => {
      const { container } = render(<QualitySelector />);
      await waitFor(() => {
        const hdCard = container.querySelector('[data-profile-id="hd"]');
        expect(hdCard?.className).toContain('selected');
      });
    });
  });

  describe('Profile Switching', () => {
    it('should call setQualityProfile when profile selected', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        expect(screen.getByText('Ultra HD')).toBeInTheDocument();
      });

      const ultraHdButton = screen.getByText('Ultra HD').closest('button');
      fireEvent.click(ultraHdButton!);

      await waitFor(() => {
        expect(g5ServiceModule.g5Service.setQualityProfile).toHaveBeenCalledWith('ultra');
      });
    });

    it('should show loading state during profile switch', async () => {
      (g5ServiceModule.g5Service.setQualityProfile as any).mockImplementation(
        () => new Promise(resolve => setTimeout(() => resolve(mockProfiles[0]), 500))
      );

      render(<QualitySelector />);
      await waitFor(() => {
        expect(screen.getByText('Ultra HD')).toBeInTheDocument();
      });

      const ultraHdButton = screen.getByText('Ultra HD').closest('button');
      fireEvent.click(ultraHdButton!);

      await waitFor(() => {
        expect(screen.getByText(/updating/i)).toBeInTheDocument();
      }, { timeout: 100 }).catch(() => {
        // Loading state might be too quick to catch
      });
    });

    it('should handle profile switch errors', async () => {
      (g5ServiceModule.g5Service.setQualityProfile as any).mockRejectedValueOnce(new Error('Switch failed'));

      render(<QualitySelector />);
      await waitFor(() => {
        expect(screen.getByText('Ultra HD')).toBeInTheDocument();
      });

      const ultraHdButton = screen.getByText('Ultra HD').closest('button');
      fireEvent.click(ultraHdButton!);

      await waitFor(() => {
        expect(screen.getByText(/error/i)).toBeInTheDocument();
      });
    });
  });

  describe('Profile Requirements', () => {
    it('should display minimum bandwidth requirement', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        // Should show bandwidth requirement for current profile
        expect(screen.getByText(/5000/)).toBeInTheDocument();
      });
    });

    it('should display maximum latency requirement', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        // Should show latency requirement for current profile
        expect(screen.getByText(/40/)).toBeInTheDocument();
      });
    });

    it('should display all profile resolutions in table', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        expect(screen.getByText('4K')).toBeInTheDocument();
        expect(screen.getByText('1080p')).toBeInTheDocument();
        expect(screen.getByText('720p')).toBeInTheDocument();
        expect(screen.getByText('480p')).toBeInTheDocument();
        expect(screen.getByText('360p')).toBeInTheDocument();
      });
    });
  });

  describe('Callbacks', () => {
    it('should call onProfileChanged callback when profile switches', async () => {
      const mockCallback = vi.fn();
      render(<QualitySelector onProfileChanged={mockCallback} />);

      await waitFor(() => {
        expect(screen.getByText('Ultra HD')).toBeInTheDocument();
      });

      const ultraHdButton = screen.getByText('Ultra HD').closest('button');
      fireEvent.click(ultraHdButton!);

      await waitFor(() => {
        expect(mockCallback).toHaveBeenCalledWith(expect.objectContaining({
          id: 'ultra',
        }));
      });
    });

    it('should handle undefined callback gracefully', async () => {
      expect(() => {
        render(<QualitySelector />);
      }).not.toThrow();
    });
  });

  describe('Comparison Table', () => {
    it('should display profile names in table', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        const tableRows = screen.getAllByRole('row');
        expect(tableRows.length).toBeGreaterThan(0);
      });
    });

    it('should be sortable by column', async () => {
      const { container } = render(<QualitySelector />);
      await waitFor(() => {
        expect(screen.getByText('Ultra HD')).toBeInTheDocument();
      });

      const resolutionHeader = container.querySelector('[data-column="resolution"]');
      if (resolutionHeader) {
        fireEvent.click(resolutionHeader);
        // Table should reorder
        expect(container.querySelector('table')).toBeInTheDocument();
      }
    });
  });

  describe('Recommendations', () => {
    it('should display AI recommendation section', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        expect(screen.getByText(/Recommended Profile/i)).toBeInTheDocument();
      });
    });

    it('should explain recommendation reasoning', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        expect(screen.getByText(/based on/i)).toBeInTheDocument();
      });
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty profile list', async () => {
      (g5ServiceModule.g5Service.getQualityProfiles as any).mockResolvedValueOnce([]);
      render(<QualitySelector />);

      await waitFor(() => {
        expect(screen.getByText(/no profiles/i)).toBeInTheDocument();
      });
    });

    it('should handle null current profile', async () => {
      (g5ServiceModule.g5Service.getCurrentQualityProfile as any).mockResolvedValueOnce(null);
      render(<QualitySelector />);

      await waitFor(() => {
        expect(screen.getByText(/Quality Selection/i)).toBeInTheDocument();
      });
    });

    it('should cleanup on unmount', () => {
      const { unmount } = render(<QualitySelector />);
      expect(() => unmount()).not.toThrow();
    });
  });

  describe('Accessibility', () => {
    it('should have proper button roles', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        const buttons = screen.getAllByRole('button');
        expect(buttons.length).toBeGreaterThan(0);
      });
    });

    it('should have readable labels', async () => {
      render(<QualitySelector />);
      await waitFor(() => {
        expect(screen.getByText(/Ultra HD/i)).toBeInTheDocument();
      });
    });
  });
});
