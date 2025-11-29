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
    });

    this.socket.on('disconnect', (reason: string) => {
      console.log('Signaling disconnected:', reason);
    });

    this.socket.on('error', (error: any) => {
      console.error('Signaling error:', error);
    });
  }

  async getRouterCapabilities(): Promise<RouterCapabilities> {
    return new Promise((resolve, reject) => {
      this.socket.emit('getRouterCapabilities', (err: any, capabilities: RouterCapabilities) => {
        if (err) {
          reject(err);
        } else {
          resolve(capabilities);
        }
      });
    });
  }

  async createProducerTransport(rtpCapabilities: any): Promise<TransportOptions> {
    return new Promise((resolve, reject) => {
      this.socket.emit(
        'createProducerTransport',
        { rtpCapabilities },
        (err: any, transportOptions: TransportOptions) => {
          if (err) {
            reject(err);
          } else {
            resolve(transportOptions);
          }
        }
      );
    });
  }

  async createConsumerTransport(rtpCapabilities: any): Promise<TransportOptions> {
    return new Promise((resolve, reject) => {
      this.socket.emit(
        'createConsumerTransport',
        { rtpCapabilities },
        (err: any, transportOptions: TransportOptions) => {
          if (err) {
            reject(err);
          } else {
            resolve(transportOptions);
          }
        }
      );
    });
  }

  async connectProducerTransport(dtlsParameters: any): Promise<void> {
    return new Promise((resolve, reject) => {
      this.socket.emit(
        'connectProducerTransport',
        { dtlsParameters },
        (err: any) => {
          if (err) {
            reject(err);
          } else {
            resolve();
          }
        }
      );
    });
  }

  async connectConsumerTransport(dtlsParameters: any): Promise<void> {
    return new Promise((resolve, reject) => {
      this.socket.emit(
        'connectConsumerTransport',
        { dtlsParameters },
        (err: any) => {
          if (err) {
            reject(err);
          } else {
            resolve();
          }
        }
      );
    });
  }

  async produce(kind: string, rtpParameters: any): Promise<{ id: string; producerId: string }> {
    return new Promise((resolve, reject) => {
      this.socket.emit(
        'produce',
        { kind, rtpParameters },
        (err: any, data: { id: string; producerId: string }) => {
          if (err) {
            reject(err);
          } else {
            resolve(data);
          }
        }
      );
    });
  }

  async consume(producerId: string, rtpCapabilities: any): Promise<ConsumerInfo> {
    return new Promise((resolve, reject) => {
      this.socket.emit(
        'consume',
        { producerId, rtpCapabilities },
        (err: any, consumerInfo: ConsumerInfo) => {
          if (err) {
            reject(err);
          } else {
            resolve(consumerInfo);
          }
        }
      );
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
    const response = await api.get(`/streaming/rooms/${roomId}/participants`);
    return response.data;
  }

  async recordSession(roomId: string) {
    const response = await api.post(`/streaming/rooms/${roomId}/record`, {});
    return response.data;
  }

  async stopRecording(sessionId: string) {
    const response = await api.post(`/streaming/sessions/${sessionId}/stop-record`, {});
    return response.data;
  }

  async collectMetrics(sessionId: string, metrics: any) {
    const response = await api.post(`/streaming/sessions/${sessionId}/metrics`, metrics);
    return response.data;
  }
}
