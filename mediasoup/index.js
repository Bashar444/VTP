const mediasoup = require('mediasoup');
const express = require('express');
const dotenv = require('dotenv');

dotenv.config();

const app = express();
const PORT = process.env.PORT || 3000;

// Global state
let worker;
let router;
const rooms = new Map(); // room_id -> { router, producers, consumers }
const transports = new Map(); // transport_id -> transport

// Initialize Mediasoup Worker
async function initMediasoup() {
  try {
    console.log('[Mediasoup] Initializing worker...');
    
    worker = await mediasoup.createWorker({
      logLevel: 'debug',
      logTags: ['info', 'ice', 'dtls', 'rtp', 'srtp', 'rtcp', 'rtp', 'media'],
      rtcMinPort: 40000,
      rtcMaxPort: 40100,
    });

    console.log(`[Mediasoup] Worker created with PID: ${worker.pid}`);

    // Create a default router with audio/video codecs
    router = await worker.createRouter({
      mediaCodecs: [
        {
          kind: 'audio',
          mimeType: 'audio/opus',
          clockRate: 48000,
          channels: 2,
        },
        {
          kind: 'video',
          mimeType: 'video/VP8',
          clockRate: 90000,
          parameters: {
            'x-google-start-bitrate': 1000,
          },
        },
        {
          kind: 'video',
          mimeType: 'video/VP9',
          clockRate: 90000,
        },
        {
          kind: 'video',
          mimeType: 'video/H264',
          clockRate: 90000,
        },
      ],
    });

    console.log('[Mediasoup] Router created');
  } catch (error) {
    console.error('[Mediasoup] Failed to initialize:', error);
    process.exit(1);
  }
}

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ status: 'ok', service: 'mediasoup-sfu' });
});

// Get router capabilities
app.get('/rtp-capabilities', (req, res) => {
  if (!router) {
    return res.status(500).json({ error: 'Router not initialized' });
  }
  res.json(router.rtpCapabilities);
});

// Create transport
app.post('/transports', express.json(), async (req, res) => {
  const { roomId, type } = req.body;

  if (!roomId || !type) {
    return res.status(400).json({ error: 'Missing roomId or type' });
  }

  try {
    if (!rooms.has(roomId)) {
      rooms.set(roomId, {
        producers: new Map(),
        consumers: new Map(),
      });
      console.log(`[Mediasoup] Created room: ${roomId}`);
    }

    const transport = await router.createWebRtcTransport({
      listenIps: [{ ip: process.env.MEDIASOUP_LISTEN_IP || '0.0.0.0' }],
      announceIp: process.env.MEDIASOUP_ANNOUNCED_IP || '127.0.0.1',
      enableUdp: true,
      enableTcp: true,
      preferUdp: true,
    });

    const transportId = transport.id;
    transports.set(transportId, { transport, roomId, type });

    console.log(`[Mediasoup] Created transport ${transportId} in room ${roomId}`);

    res.json({
      id: transport.id,
      iceParameters: transport.iceParameters,
      iceCandidates: transport.iceCandidates,
      dtlsParameters: transport.dtlsParameters,
    });
  } catch (error) {
    console.error('[Mediasoup] Failed to create transport:', error);
    res.status(500).json({ error: error.message });
  }
});

// Connect transport
app.post('/transports/:transportId/connect', express.json(), async (req, res) => {
  const { transportId } = req.params;
  const { dtlsParameters } = req.body;

  try {
    const transportData = transports.get(transportId);
    if (!transportData) {
      return res.status(404).json({ error: 'Transport not found' });
    }

    await transportData.transport.connect({ dtlsParameters });
    console.log(`[Mediasoup] Connected transport: ${transportId}`);

    res.json({ ok: true });
  } catch (error) {
    console.error('[Mediasoup] Failed to connect transport:', error);
    res.status(500).json({ error: error.message });
  }
});

// Produce (send media)
app.post('/transports/:transportId/produce', express.json(), async (req, res) => {
  const { transportId } = req.params;
  const { kind, rtpParameters } = req.body;

  try {
    const transportData = transports.get(transportId);
    if (!transportData) {
      return res.status(404).json({ error: 'Transport not found' });
    }

    const producer = await transportData.transport.produce({
      kind,
      rtpParameters,
    });

    const roomData = rooms.get(transportData.roomId);
    roomData.producers.set(producer.id, producer);

    console.log(`[Mediasoup] Created producer ${producer.id} in room ${transportData.roomId}`);

    res.json({ id: producer.id });
  } catch (error) {
    console.error('[Mediasoup] Failed to produce:', error);
    res.status(500).json({ error: error.message });
  }
});

// Consume (receive media)
app.post('/transports/:transportId/consume', express.json(), async (req, res) => {
  const { transportId } = req.params;
  const { producerId } = req.body;

  try {
    const transportData = transports.get(transportId);
    if (!transportData) {
      return res.status(404).json({ error: 'Transport not found' });
    }

    const consumer = await transportData.transport.consume({
      producerId,
      rtpCapabilities: router.rtpCapabilities,
      paused: false,
    });

    const roomData = rooms.get(transportData.roomId);
    roomData.consumers.set(consumer.id, consumer);

    console.log(`[Mediasoup] Created consumer ${consumer.id} in room ${transportData.roomId}`);

    res.json({
      id: consumer.id,
      producerId: consumer.producerId,
      kind: consumer.kind,
      rtpParameters: consumer.rtpParameters,
    });
  } catch (error) {
    console.error('[Mediasoup] Failed to consume:', error);
    res.status(500).json({ error: error.message });
  }
});

// Close transport
app.post('/transports/:transportId/close', async (req, res) => {
  const { transportId } = req.params;

  try {
    const transportData = transports.get(transportId);
    if (transportData) {
      transportData.transport.close();
      transports.delete(transportId);
      console.log(`[Mediasoup] Closed transport: ${transportId}`);
    }

    res.json({ ok: true });
  } catch (error) {
    console.error('[Mediasoup] Failed to close transport:', error);
    res.status(500).json({ error: error.message });
  }
});

// Start server
app.listen(PORT, async () => {
  await initMediasoup();
  console.log(`[Server] Mediasoup SFU listening on port ${PORT}`);
});

// Graceful shutdown
process.on('SIGINT', async () => {
  console.log('\n[Server] Shutting down gracefully...');
  if (worker) {
    await worker.close();
  }
  process.exit(0);
});
