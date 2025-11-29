import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { StreamingControls } from '@/components/streaming/StreamingControls';

describe('StreamingControls', () => {
  const mockFunctions = {
    onToggleAudio: vi.fn(),
    onToggleVideo: vi.fn(),
    onLeave: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render all control buttons', () => {
    render(
      <StreamingControls
        isAudioEnabled={true}
        isVideoEnabled={true}
        onToggleAudio={mockFunctions.onToggleAudio}
        onToggleVideo={mockFunctions.onToggleVideo}
        onLeave={mockFunctions.onLeave}
      />
    );

    expect(screen.getByTitle('Mute')).toBeInTheDocument();
    expect(screen.getByTitle('Stop camera')).toBeInTheDocument();
    expect(screen.getByTitle('Leave call')).toBeInTheDocument();
  });

  it('should toggle audio when button is clicked', async () => {
    mockFunctions.onToggleAudio.mockResolvedValue(undefined);

    render(
      <StreamingControls
        isAudioEnabled={true}
        isVideoEnabled={true}
        onToggleAudio={mockFunctions.onToggleAudio}
        onToggleVideo={mockFunctions.onToggleVideo}
        onLeave={mockFunctions.onLeave}
      />
    );

    const audioButton = screen.getByTitle('Mute');
    fireEvent.click(audioButton);

    await waitFor(() => {
      expect(mockFunctions.onToggleAudio).toHaveBeenCalledWith(false);
    });
  });

  it('should toggle video when button is clicked', async () => {
    mockFunctions.onToggleVideo.mockResolvedValue(undefined);

    render(
      <StreamingControls
        isAudioEnabled={true}
        isVideoEnabled={true}
        onToggleAudio={mockFunctions.onToggleAudio}
        onToggleVideo={mockFunctions.onToggleVideo}
        onLeave={mockFunctions.onLeave}
      />
    );

    const videoButton = screen.getByTitle('Stop camera');
    fireEvent.click(videoButton);

    await waitFor(() => {
      expect(mockFunctions.onToggleVideo).toHaveBeenCalledWith(false);
    });
  });

  it('should call onLeave when leave button is clicked', () => {
    render(
      <StreamingControls
        isAudioEnabled={true}
        isVideoEnabled={true}
        onToggleAudio={mockFunctions.onToggleAudio}
        onToggleVideo={mockFunctions.onToggleVideo}
        onLeave={mockFunctions.onLeave}
      />
    );

    const leaveButton = screen.getByTitle('Leave call');
    fireEvent.click(leaveButton);

    expect(mockFunctions.onLeave).toHaveBeenCalled();
  });

  it('should show disabled state when isLoading is true', () => {
    render(
      <StreamingControls
        isAudioEnabled={true}
        isVideoEnabled={true}
        onToggleAudio={mockFunctions.onToggleAudio}
        onToggleVideo={mockFunctions.onToggleVideo}
        onLeave={mockFunctions.onLeave}
        isLoading={true}
      />
    );

    const buttons = screen.getAllByRole('button');
    buttons.forEach(button => {
      expect(button).toBeDisabled();
    });
  });

  it('should reflect audio disabled state visually', () => {
    const { rerender } = render(
      <StreamingControls
        isAudioEnabled={true}
        isVideoEnabled={true}
        onToggleAudio={mockFunctions.onToggleAudio}
        onToggleVideo={mockFunctions.onToggleVideo}
        onLeave={mockFunctions.onLeave}
      />
    );

    let audioButton = screen.getByTitle('Mute');
    expect(audioButton).toHaveClass('bg-blue-600');

    rerender(
      <StreamingControls
        isAudioEnabled={false}
        isVideoEnabled={true}
        onToggleAudio={mockFunctions.onToggleAudio}
        onToggleVideo={mockFunctions.onToggleVideo}
        onLeave={mockFunctions.onLeave}
      />
    );

    audioButton = screen.getByTitle('Unmute');
    expect(audioButton).toHaveClass('bg-red-600');
  });
});
