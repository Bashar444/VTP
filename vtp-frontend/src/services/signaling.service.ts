import io from 'socket.io-client';
import { api } from './api.client';

interface RouterCapabilities {
  codecs: any[];
  headerExtensions: any[];
  fecMechanisms: any[];
}

interface TransportOptions {
  id: string;
  iceParameters: any;
  iceCandidates: any[];
  dtlsParameters: any;
}

interface ConsumerInfo {
  id: string;
  producerId: string;
  kind: string;
  rtpParameters: any;
}

export default class SignalingService {
  private socket: any;
  private roomId: string;
  private userId: string;
  private token: string;
  private baseUrl: string;
  private routerCapabilities: RouterCapabilities | null = null;
  private joined = false;

  constructor(roomId: string, userId: string, token: string) {
    this.roomId = roomId;
    this.userId = userId;
    this.token = token;
    this.baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001';

    this.initializeSocket();
  }

  private initializeSocket() {
    this.socket = io(this.baseUrl, {
      query: {
        roomId: this.roomId,
        userId: this.userId,
      },
      auth: {
        token: this.token,
      },
      reconnection: true,
      reconnectionDelay: 1000,
      reconnectionDelayMax: 5000,
      reconnectionAttempts: 5,
    });

    this.socket.on('connect', () => {
      console.log('Signaling connected:', this.socket.id);
      // Auto-join the signalling room and fetch router capabilities
      this.joinRoom().catch((e: any) => console.warn('joinRoom failed', e));
    });

    this.socket.on('disconnect', (reason: string) => {
      console.log('Signaling disconnected:', reason);
    });

    this.socket.on('error', (error: any) => {
      console.error('Signaling error:', error);
    });
  }

  // Join room on backend signalling to retrieve Mediasoup router capabilities
  private async joinRoom(): Promise<void> {
    if (this.joined) return;
    try {
      // Fetch profile for basic identity
      const profile = await api.get('/api/v1/auth/profile').then(r => r.data as any).catch(() => ({ full_name: 'Guest', email: 'guest@example.com', role: 'student' }));
      const payload = JSON.stringify({
        roomId: this.roomId,
        userId: this.userId,
        email: profile.email || 'guest@example.com',
        fullName: profile.full_name || 'Guest User',
        role: profile.role || 'student',
        roomName: this.roomId,
        isProducer: true,
      });

      await new Promise<void>((resolve, reject) => {
        // Expect a single joined-room response carrying router capabilities
        const onJoined = (resp: any) => {
          try {
            this.joined = true;
            this.routerCapabilities = resp?.Mediasoup?.RtpCapabilities || null;
            this.socket.off('joined-room', onJoined);
            resolve();
          } catch (e) {
            reject(e);
          }
        };
        this.socket.on('joined-room', onJoined);
        this.socket.emit('join-room', payload);
      });
    } catch (e) {
      throw e;
    }
  }

  async getRouterCapabilities(): Promise<RouterCapabilities> {
    if (this.routerCapabilities) return this.routerCapabilities;
    // If not yet joined, attempt join and then return
    await this.joinRoom();
    if (!this.routerCapabilities) throw new Error('Router capabilities unavailable');
    return this.routerCapabilities;
  }

  async createProducerTransport(_rtpCapabilities: any): Promise<TransportOptions> {
    // Backend expects a generic create-transport with direction=send and returns via 'transport-created'
    return new Promise((resolve, reject) => {
      const handler = (transport: any) => {
        this.socket.off('transport-created', handler);
        resolve({
          id: transport.transportId,
          iceParameters: transport.iceParameters,
          iceCandidates: transport.iceCandidates,
          dtlsParameters: transport.dtlsParameters,
        } as any);
      };
      this.socket.on('transport-created', handler);
      this.socket.emit('create-transport', JSON.stringify({ roomId: this.roomId, direction: 'send' }));
      // Basic timeout safety
      setTimeout(() => reject(new Error('createProducerTransport timeout')), 10000);
    });
  }

  async createConsumerTransport(_rtpCapabilities: any): Promise<TransportOptions> {
    return new Promise((resolve, reject) => {
      const handler = (transport: any) => {
        this.socket.off('transport-created', handler);
        resolve({
          id: transport.transportId,
          iceParameters: transport.iceParameters,
          iceCandidates: transport.iceCandidates,
          dtlsParameters: transport.dtlsParameters,
        } as any);
      };
      this.socket.on('transport-created', handler);
      this.socket.emit('create-transport', JSON.stringify({ roomId: this.roomId, direction: 'recv' }));
      setTimeout(() => reject(new Error('createConsumerTransport timeout')), 10000);
    });
  }

  async connectProducerTransport(dtlsParameters: any): Promise<void> {
    return new Promise((resolve, reject) => {
      const handler = () => {
        this.socket.off('transport-connected', handler);
        resolve();
      };
      this.socket.on('transport-connected', handler);
      this.socket.emit('connect-transport', JSON.stringify({ roomId: this.roomId, transportId: 'send', dtlsParameters }));
      setTimeout(() => reject(new Error('connectProducerTransport timeout')), 10000);
    });
  }

  async connectConsumerTransport(dtlsParameters: any): Promise<void> {
    return new Promise((resolve, reject) => {
      const handler = () => {
        this.socket.off('transport-connected', handler);
        resolve();
      };
      this.socket.on('transport-connected', handler);
      this.socket.emit('connect-transport', JSON.stringify({ roomId: this.roomId, transportId: 'recv', dtlsParameters }));
      setTimeout(() => reject(new Error('connectConsumerTransport timeout')), 10000);
    });
  }

  async produce(kind: string, rtpParameters: any): Promise<{ id: string; producerId: string }> {
    return new Promise((resolve, reject) => {
      const handler = (producer: any) => {
        this.socket.off('producer-created', handler);
        // Broadcast newProducer to notify others if backend doesn't
        this.socket.emit('newProducer', { producerId: producer.id, peerId: this.userId, kind });
        resolve({ id: producer.id, producerId: producer.id });
      };
      this.socket.on('producer-created', handler);
      this.socket.emit('produce', JSON.stringify({ roomId: this.roomId, kind, rtpParameters }));
      setTimeout(() => reject(new Error('produce timeout')), 10000);
    });
  }

  async consume(producerId: string, rtpCapabilities: any): Promise<ConsumerInfo> {
    return new Promise((resolve, reject) => {
      const handler = (consumerInfo: any) => {
        this.socket.off('consumer-created', handler);
        resolve({
          id: consumerInfo.id,
          producerId: consumerInfo.producerId,
          kind: consumerInfo.kind,
          rtpParameters: consumerInfo.rtpParameters,
        });
      };
      this.socket.on('consumer-created', handler);
      this.socket.emit('consume', JSON.stringify({ roomId: this.roomId, producerId, rtpCapabilities }));
      setTimeout(() => reject(new Error('consume timeout')), 10000);
    });
  }

  // Chat history (graceful failure if endpoint not implemented)
  async getChatHistory(roomId: string) {
    try {
      const response = await api.get(`/streaming/rooms/${roomId}/chat`);
      return response.data;
    } catch (e) {
      return [];
    }
  }

  onNewProducer(callback: (producerId: string, peerId: string, kind: string) => void) {
    this.socket.on('newProducer', ({ producerId, peerId, kind }: any) => {
      callback(producerId, peerId, kind);
    });
  }

  onPeerJoined(callback: (peerId: string, peerInfo: any) => void) {
    this.socket.on('peerJoined', ({ peerId, peerInfo }: any) => {
      callback(peerId, peerInfo);
    });
  }

  onPeerLeft(callback: (peerId: string) => void) {
    this.socket.on('peerLeft', ({ peerId }: any) => {
      callback(peerId);
    });
  }

  async disconnect(): Promise<void> {
    return new Promise((resolve) => {
      if (this.socket) {
        this.socket.disconnect();
        setTimeout(resolve, 100);
      } else {
        resolve();
      }
    });
  }

  // REST API methods for streaming sessions
  async getParticipants(roomId: string) {
    const response = await api.get(`/api/v1/streaming/rooms/${roomId}/participants`);
    return response.data;
  }

  async recordSession(roomId: string) {
    const response = await api.post(`/api/v1/streaming/rooms/${roomId}/record`, {});
    return response.data;
  }

  async stopRecording(sessionId: string) {
    const response = await api.post(`/api/v1/streaming/sessions/${sessionId}/stop-record`, {});
    return response.data;
  }

  async collectMetrics(sessionId: string, metrics: any) {
    const response = await api.post(`/api/v1/streaming/sessions/${sessionId}/metrics`, metrics);
    return response.data;
  }
}
