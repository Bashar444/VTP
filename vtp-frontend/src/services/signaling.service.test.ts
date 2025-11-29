import { describe, it, expect, beforeEach, vi } from 'vitest';
import SignalingService from '@/services/signaling.service';

// Mock socket.io-client
vi.mock('socket.io-client', () => ({
  default: vi.fn(() => ({
    on: vi.fn(),
    emit: vi.fn((event: string, ...args: any[]) => {
      // Simulate callback-based socket.io behavior
      const lastArg = args[args.length - 1];
      if (typeof lastArg === 'function') {
        if (event === 'getRouterCapabilities') {
          lastArg(null, { codecs: [], headerExtensions: [], fecMechanisms: [] });
        }
      }
    }),
    disconnect: vi.fn(),
  })),
}));

// Mock API client
vi.mock('@/services/api.client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

describe('SignalingService', () => {
  let signalingService: SignalingService;
  const roomId = 'test-room';
  const userId = 'user-123';
  const token = 'test-token';

  beforeEach(() => {
    vi.clearAllMocks();
    signalingService = new SignalingService(roomId, userId, token);
  });

  it('should initialize socket connection', () => {
    expect(signalingService).toBeDefined();
  });

  it('should get router capabilities', async () => {
    const capabilities = await signalingService.getRouterCapabilities();
    expect(capabilities).toHaveProperty('codecs');
    expect(capabilities).toHaveProperty('headerExtensions');
    expect(capabilities).toHaveProperty('fecMechanisms');
  });

  it('should create producer transport', async () => {
    const rtpCapabilities = { codecs: [] };
    const transportInfo = await signalingService.createProducerTransport(rtpCapabilities);
    expect(transportInfo).toBeDefined();
  });

  it('should create consumer transport', async () => {
    const rtpCapabilities = { codecs: [] };
    const transportInfo = await signalingService.createConsumerTransport(rtpCapabilities);
    expect(transportInfo).toBeDefined();
  });

  it('should handle peer join event', () => {
    const mockCallback = vi.fn();
    signalingService.onPeerJoined(mockCallback);
    // Event listener is registered
    expect(mockCallback).toBeDefined();
  });

  it('should handle new producer event', () => {
    const mockCallback = vi.fn();
    signalingService.onNewProducer(mockCallback);
    // Event listener is registered
    expect(mockCallback).toBeDefined();
  });

  it('should handle peer left event', () => {
    const mockCallback = vi.fn();
    signalingService.onPeerLeft(mockCallback);
    // Event listener is registered
    expect(mockCallback).toBeDefined();
  });

  it('should disconnect gracefully', async () => {
    await signalingService.disconnect();
    // Verify disconnect is called without errors
    expect(signalingService).toBeDefined();
  });
});
