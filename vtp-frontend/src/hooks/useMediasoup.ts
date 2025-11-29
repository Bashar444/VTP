import { useEffect, useRef, useState, useCallback } from 'react';
import { useAuthStore } from '@/store/auth.store';
import { useStreamingStore } from '@/store/streaming.store';
import SignalingService from '@/services/signaling.service';

interface RTCPeerConnection {
  pc: any;
  transceiver?: any;
}

export const useMediasoup = (roomId: string) => {
  const authStore = useAuthStore();
  const streamingStore = useStreamingStore();
  
  const [localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [remoteStreams, setRemoteStreams] = useState<Map<string, MediaStream>>(
    new Map()
  );
  const [error, setError] = useState<string | null>(null);
  const [isConnected, setIsConnected] = useState(false);

  const mediasoupDeviceRef = useRef<any>(null);
  const producerTransportRef = useRef<any>(null);
  const consumerTransportRef = useRef<any>(null);
  const audioProducerRef = useRef<any>(null);
  const videoProducerRef = useRef<any>(null);
  const peersRef = useRef<Map<string, RTCPeerConnection>>(new Map());
  const signalingRef = useRef<SignalingService | null>(null);

  // Initialize Mediasoup connection
  const initializeMediasoup = useCallback(async () => {
    try {
      if (!authStore.user) {
        throw new Error('User not authenticated');
      }

      signalingRef.current = new SignalingService(
        roomId,
        authStore.user.id,
        authStore.token || ''
      );

      // Auto-consume new producers when announced by signaling
      signalingRef.current.onNewProducer(async (producerId: string, peerId: string, kind: string) => {
        try {
          // Only consume if connected and device loaded
          if (!mediasoupDeviceRef.current || !consumerTransportRef.current) return;
          await consumeRemoteStream(producerId, peerId);
        } catch (err) {
          console.warn('Failed to auto-consume producer', producerId, err);
        }
      });

      // Get router capabilities
      const routerCapabilities = await signalingRef.current.getRouterCapabilities();

      // Initialize device
      const { Device } = await import('mediasoup-client');
      mediasoupDeviceRef.current = new Device();
      await mediasoupDeviceRef.current.load({
        routerRtpCapabilities: routerCapabilities,
      });

      // Create transports
      const producerTransportInfo = await signalingRef.current.createProducerTransport(
        mediasoupDeviceRef.current.rtpCapabilities
      );

      producerTransportRef.current =
        mediasoupDeviceRef.current.createSendTransport(producerTransportInfo);

      producerTransportRef.current.on('connect', async ({ dtlsParameters }: any) => {
        await signalingRef.current?.connectProducerTransport(dtlsParameters);
      });

      producerTransportRef.current.on('produce', async ({ kind, rtpParameters }: any) => {
        const { id, producerId } = await signalingRef.current?.produce(
          kind,
          rtpParameters
        );
        return { id, producerId };
      });

      // Create consumer transport
      const consumerTransportInfo = await signalingRef.current.createConsumerTransport(
        mediasoupDeviceRef.current.rtpCapabilities
      );

      consumerTransportRef.current =
        mediasoupDeviceRef.current.createRecvTransport(consumerTransportInfo);

      consumerTransportRef.current.on('connect', async ({ dtlsParameters }: any) => {
        await signalingRef.current?.connectConsumerTransport(dtlsParameters);
      });

      setIsConnected(true);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to initialize Mediasoup';
      setError(message);
      console.error('Mediasoup initialization error:', err);
    }
  }, [roomId, authStore.user, authStore.token]);

  // Get local media stream
  const getLocalStream = useCallback(async (
    audioEnabled: boolean,
    videoEnabled: boolean
  ): Promise<MediaStream> => {
    try {
      const constraints = {
        audio: audioEnabled ? {
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true,
        } : false,
        video: videoEnabled ? {
          width: { ideal: 1280 },
          height: { ideal: 720 },
        } : false,
      };

      const stream = await navigator.mediaDevices.getUserMedia(constraints);
      setLocalStream(stream);

      // Produce audio track
      if (audioEnabled && producerTransportRef.current) {
        const audioTrack = stream.getAudioTracks()[0];
        if (audioTrack) {
          const params = {
            track: audioTrack,
            codecOptions: {
              opusStereo: true,
              opusDtx: true,
            },
          };
          audioProducerRef.current = await producerTransportRef.current.produce(params);
        }
      }

      // Produce video track
      if (videoEnabled && producerTransportRef.current) {
        const videoTrack = stream.getVideoTracks()[0];
        if (videoTrack) {
          const params = {
            track: videoTrack,
            encodings: [
              { maxBitrate: 100000 },
              { maxBitrate: 300000 },
              { maxBitrate: 900000 },
            ],
          };
          videoProducerRef.current = await producerTransportRef.current.produce(params);
        }
      }

      return stream;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to get local stream';
      setError(message);
      throw err;
    }
  }, []);

  // Toggle audio/video
  const toggleAudio = useCallback(async (enabled: boolean) => {
    if (audioProducerRef.current) {
      if (enabled) {
        await audioProducerRef.current.resume();
      } else {
        await audioProducerRef.current.pause();
      }
    }
  }, []);

  const toggleVideo = useCallback(async (enabled: boolean) => {
    if (videoProducerRef.current) {
      if (enabled) {
        await videoProducerRef.current.resume();
      } else {
        await videoProducerRef.current.pause();
      }
    }
  }, []);

  // Consume remote stream
  const consumeRemoteStream = useCallback(async (producerId: string, peerId: string) => {
    try {
      if (!signalingRef.current || !consumerTransportRef.current) {
        throw new Error('Transport not initialized');
      }

      const { rtpCapabilities } = mediasoupDeviceRef.current;
      const consumerInfo = await signalingRef.current.consume(
        producerId,
        rtpCapabilities
      );

      const consumer = await consumerTransportRef.current.consume(consumerInfo);

      // Get the track and create a new MediaStream
      const track = consumer.track;
      const stream = new MediaStream([track]);

      setRemoteStreams(prev => new Map(prev).set(peerId, stream));

      return { consumer, stream };
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to consume stream';
      setError(message);
      throw err;
    }
  }, []);

  // Cleanup
  const disconnect = useCallback(async () => {
    try {
      if (localStream) {
        localStream.getTracks().forEach(track => track.stop());
      }

      if (audioProducerRef.current) {
        await audioProducerRef.current.close();
      }

      if (videoProducerRef.current) {
        await videoProducerRef.current.close();
      }

      if (producerTransportRef.current) {
        producerTransportRef.current.close();
      }

      if (consumerTransportRef.current) {
        consumerTransportRef.current.close();
      }

      if (signalingRef.current) {
        await signalingRef.current.disconnect();
      }

      setLocalStream(null);
      setRemoteStreams(new Map());
      setIsConnected(false);
    } catch (err) {
      console.error('Cleanup error:', err);
    }
  }, [localStream]);

  // Initialize on mount
  useEffect(() => {
    initializeMediasoup();

    return () => {
      disconnect();
    };
  }, [initializeMediasoup, disconnect]);

  return {
    localStream,
    remoteStreams,
    error,
    isConnected,
    getLocalStream,
    toggleAudio,
    toggleVideo,
    consumeRemoteStream,
    disconnect,
  };
};
