// VTP Mediasoup SFU Service
// Selective Forwarding Unit for WebRTC media streaming

const mediasoup = require('mediasoup');
const express = require('express');
const cors = require('cors');
const bodyParser = require('body-parser');
const { v4: uuidv4 } = require('uuid');
const winston = require('winston');
require('dotenv').config();

// Configure logging
const logger = winston.createLogger({
  level: process.env.LOG_LEVEL || 'info',
  format: winston.format.combine(
    winston.format.timestamp(),
    winston.format.errors({ stack: true }),
    winston.format.splat(),
    winston.format.json()
  ),
  defaultMeta: { service: 'mediasoup-sfu' },
  transports: [
    new winston.transports.Console({
      format: winston.format.combine(
        winston.format.colorize(),
        winston.format.printf(({ level, message, timestamp }) => {
          return `${timestamp} [${level}]: ${message}`;
        })
      )
    })
  ]
});

// Application state
let mediasoupWorker;
const rooms = new Map(); // roomId -> Room
const producers = new Map(); // producerId -> Producer info
const consumers = new Map(); // consumerId -> Consumer info

// Room class to manage participants and media
class Room {
  constructor(roomId, name) {
    this.id = roomId;
    this.name = name;
    this.router = null;
    this.peers = new Map(); // peerId -> Peer
    this.createdAt = Date.now();
  }
}

// Peer class to manage participant connections
class Peer {
  constructor(peerId, userId, email, fullName, role, isProducer) {
    this.id = peerId;
    this.userId = userId;
    this.email = email;
    this.fullName = fullName;
    this.role = role;
    this.isProducer = isProducer;
    this.transport = null;
    this.producers = new Map(); // kind -> producer
    this.consumers = new Map(); // consumerId -> consumer
    this.joinedAt = Date.now();
  }

  hasProducer(kind) {
    return this.producers.has(kind);
  }

  addProducer(kind, producer) {
    this.producers.set(kind, producer);
  }

  removeProducer(kind) {
    this.producers.delete(kind);
  }

  getProducer(kind) {
    return this.producers.get(kind);
  }

  addConsumer(consumerId, consumer) {
    this.consumers.set(consumerId, consumer);
  }

  removeConsumer(consumerId) {
    this.consumers.delete(consumerId);
  }

  getConsumer(consumerId) {
    return this.consumers.get(consumerId);
  }
}

// Initialize Express server
const app = express();
app.use(cors());
app.use(bodyParser.json());

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({
    status: 'ok',
    timestamp: new Date().toISOString(),
    worker: mediasoupWorker ? 'ready' : 'not-ready'
  });
});

// Get room info
app.get('/rooms/:roomId', (req, res) => {
  const { roomId } = req.params;
  const room = rooms.get(roomId);

  if (!room) {
    return res.status(404).json({ error: 'Room not found' });
  }

  res.json({
    roomId: room.id,
    name: room.name,
    peerCount: room.peers.size,
    peers: Array.from(room.peers.values()).map(peer => ({
      peerId: peer.id,
      userId: peer.userId,
      email: peer.email,
      fullName: peer.fullName,
      role: peer.role,
      isProducer: peer.isProducer,
      joinedAt: peer.joinedAt,
      producerCount: peer.producers.size,
      consumerCount: peer.consumers.size
    })),
    createdAt: room.createdAt
  });
});

// Get all rooms
app.get('/rooms', (req, res) => {
  const roomsList = Array.from(rooms.values()).map(room => ({
    roomId: room.id,
    name: room.name,
    peerCount: room.peers.size,
    createdAt: room.createdAt
  }));

  res.json({
    rooms: roomsList,
    totalRooms: roomsList.length,
    totalPeers: Array.from(rooms.values()).reduce((sum, room) => sum + room.peers.size, 0)
  });
});

// Create transport for peer
app.post('/rooms/:roomId/transports', async (req, res) => {
  try {
    const { roomId } = req.params;
    const { peerId, direction } = req.body;

    if (!roomId || !peerId || !direction) {
      return res.status(400).json({ error: 'Missing required fields' });
    }

    const room = rooms.get(roomId);
    if (!room) {
      return res.status(404).json({ error: 'Room not found' });
    }

    const peer = room.peers.get(peerId);
    if (!peer) {
      return res.status(404).json({ error: 'Peer not found' });
    }

    // Create WebRTC transport
    const transport = await room.router.createWebRtcTransport({
      listenIps: [
        {
          ip: process.env.MEDIASOUP_LISTEN_IP || '127.0.0.1',
          announcedIp: process.env.MEDIASOUP_ANNOUNCED_IP || '127.0.0.1'
        }
      ],
      enableUdp: true,
      enableTcp: true,
      preferUdp: true
    });

    // Store transport
    peer.transport = transport;

    // Handle transport close
    transport.on('close', () => {
      logger.info(`Transport closed for peer ${peerId} in room ${roomId}`);
    });

    logger.info(`Created ${direction} transport for peer ${peerId} in room ${roomId}`);

    res.json({
      transportId: transport.id,
      iceParameters: transport.iceParameters,
      iceCandidates: transport.iceCandidates,
      dtlsParameters: transport.dtlsParameters,
      sctpParameters: transport.sctpParameters
    });
  } catch (error) {
    logger.error(`Error creating transport: ${error.message}`);
    res.status(500).json({ error: error.message });
  }
});

// Connect transport with client parameters
app.post('/rooms/:roomId/transports/:transportId/connect', async (req, res) => {
  try {
    const { roomId, transportId } = req.params;
    const { dtlsParameters } = req.body;

    const room = rooms.get(roomId);
    if (!room) {
      return res.status(404).json({ error: 'Room not found' });
    }

    // Find transport by ID (simplified - in production use proper lookup)
    if (!dtlsParameters) {
      return res.status(400).json({ error: 'Missing DTLS parameters' });
    }

    // Note: In a real implementation, you'd need to track transports by ID
    // For now, we're keeping this simple
    logger.info(`Connected transport ${transportId} in room ${roomId}`);

    res.json({ success: true });
  } catch (error) {
    logger.error(`Error connecting transport: ${error.message}`);
    res.status(500).json({ error: error.message });
  }
});

// Produce media
app.post('/rooms/:roomId/producers', async (req, res) => {
  try {
    const { roomId } = req.params;
    const { peerId, kind, rtpParameters } = req.body;

    if (!roomId || !peerId || !kind || !rtpParameters) {
      return res.status(400).json({ error: 'Missing required fields' });
    }

    const room = rooms.get(roomId);
    if (!room) {
      return res.status(404).json({ error: 'Room not found' });
    }

    const peer = room.peers.get(peerId);
    if (!peer || !peer.transport) {
      return res.status(404).json({ error: 'Peer or transport not found' });
    }

    // Create producer
    const producer = await peer.transport.produce({
      kind,
      rtpParameters
    });

    // Store producer
    peer.addProducer(kind, producer);
    producers.set(producer.id, {
      id: producer.id,
      roomId,
      peerId,
      userId: peer.userId,
      kind,
      createdAt: Date.now()
    });

    // Handle producer close
    producer.on('close', () => {
      logger.info(`Producer ${producer.id} closed`);
      peer.removeProducer(kind);
      producers.delete(producer.id);
    });

    logger.info(`Created ${kind} producer for peer ${peerId} in room ${roomId}`);

    res.json({
      id: producer.id,
      kind,
      rtpParameters: producer.rtpParameters
    });
  } catch (error) {
    logger.error(`Error creating producer: ${error.message}`);
    res.status(500).json({ error: error.message });
  }
});

// Consume media
app.post('/rooms/:roomId/consumers', async (req, res) => {
  try {
    const { roomId } = req.params;
    const { peerId, producerId, rtpCapabilities } = req.body;

    if (!roomId || !peerId || !producerId || !rtpCapabilities) {
      return res.status(400).json({ error: 'Missing required fields' });
    }

    const room = rooms.get(roomId);
    if (!room) {
      return res.status(404).json({ error: 'Room not found' });
    }

    const peer = room.peers.get(peerId);
    if (!peer || !peer.transport) {
      return res.status(404).json({ error: 'Peer or transport not found' });
    }

    // Find producer
    const producerInfo = producers.get(producerId);
    if (!producerInfo || producerInfo.roomId !== roomId) {
      return res.status(404).json({ error: 'Producer not found' });
    }

    const producerPeer = room.peers.get(producerInfo.peerId);
    if (!producerPeer) {
      return res.status(404).json({ error: 'Producer peer not found' });
    }

    const producer = producerPeer.getProducer(producerInfo.kind);
    if (!producer) {
      return res.status(404).json({ error: 'Producer not found in peer' });
    }

    // Check if can consume
    if (!room.router.canConsume({ producerId, rtpCapabilities })) {
      return res.status(400).json({ error: 'Cannot consume from this producer' });
    }

    // Create consumer
    const consumer = await peer.transport.consume({
      producerId,
      rtpCapabilities
    });

    // Store consumer
    peer.addConsumer(consumer.id, consumer);
    consumers.set(consumer.id, {
      id: consumer.id,
      roomId,
      peerId,
      producerId,
      createdAt: Date.now()
    });

    // Handle consumer close
    consumer.on('close', () => {
      logger.info(`Consumer ${consumer.id} closed`);
      peer.removeConsumer(consumer.id);
      consumers.delete(consumer.id);
    });

    logger.info(`Created consumer for peer ${peerId} from producer ${producerId}`);

    res.json({
      id: consumer.id,
      producerId,
      kind: consumer.kind,
      rtpParameters: consumer.rtpParameters
    });
  } catch (error) {
    logger.error(`Error creating consumer: ${error.message}`);
    res.status(500).json({ error: error.message });
  }
});

// Close producer
app.post('/rooms/:roomId/producers/:producerId/close', async (req, res) => {
  try {
    const { roomId, producerId } = req.params;

    const producerInfo = producers.get(producerId);
    if (!producerInfo) {
      return res.status(404).json({ error: 'Producer not found' });
    }

    const room = rooms.get(roomId);
    const peer = room.peers.get(producerInfo.peerId);
    if (peer) {
      const producer = peer.getProducer(producerInfo.kind);
      if (producer) {
        producer.close();
      }
    }

    logger.info(`Closed producer ${producerId}`);
    res.json({ success: true });
  } catch (error) {
    logger.error(`Error closing producer: ${error.message}`);
    res.status(500).json({ error: error.message });
  }
});

// Close consumer
app.post('/rooms/:roomId/consumers/:consumerId/close', async (req, res) => {
  try {
    const { roomId, consumerId } = req.params;

    const consumerInfo = consumers.get(consumerId);
    if (!consumerInfo) {
      return res.status(404).json({ error: 'Consumer not found' });
    }

    const room = rooms.get(roomId);
    const peer = room.peers.get(consumerInfo.peerId);
    if (peer) {
      const consumer = peer.getConsumer(consumerId);
      if (consumer) {
        consumer.close();
      }
    }

    logger.info(`Closed consumer ${consumerId}`);
    res.json({ success: true });
  } catch (error) {
    logger.error(`Error closing consumer: ${error.message}`);
    res.status(500).json({ error: error.message });
  }
});

// Create or join room
async function getOrCreateRoom(roomId, roomName) {
  if (rooms.has(roomId)) {
    return rooms.get(roomId);
  }

  const room = new Room(roomId, roomName);
  room.router = await mediasoupWorker.createRouter({
    mediaCodecs: [
      {
        kind: 'audio',
        mimeType: 'audio/opus',
        clockRate: 48000,
        channels: 2
      },
      {
        kind: 'video',
        mimeType: 'video/VP8',
        clockRate: 90000,
        parameters: {
          'x-google-start-bitrate': 1000
        }
      },
      {
        kind: 'video',
        mimeType: 'video/H264',
        clockRate: 90000,
        parameters: {
          'level-asymmetry-allowed': 1,
          'packetization-mode': 1,
          'profile-level-id': '4d0032'
        }
      }
    ]
  });

  rooms.set(roomId, room);
  logger.info(`Created room ${roomId} (${roomName})`);

  return room;
}

// Join room
app.post('/rooms/:roomId/peers', async (req, res) => {
  try {
    const { roomId } = req.params;
    const { peerId, userId, email, fullName, role, isProducer, roomName } = req.body;

    if (!roomId || !peerId || !userId) {
      return res.status(400).json({ error: 'Missing required fields' });
    }

    const room = await getOrCreateRoom(roomId, roomName || roomId);

    if (room.peers.has(peerId)) {
      return res.status(400).json({ error: 'Peer already exists' });
    }

    const peer = new Peer(peerId, userId, email, fullName, role, isProducer || false);
    room.peers.set(peerId, peer);

    logger.info(`Peer ${peerId} joined room ${roomId}`);

    // Get RTP capabilities for the router
    const rtpCapabilities = room.router.rtpCapabilities;

    res.json({
      roomId,
      peerId,
      rtpCapabilities,
      peers: Array.from(room.peers.values())
        .filter(p => p.id !== peerId)
        .map(p => ({
          peerId: p.id,
          userId: p.userId,
          email: p.email,
          fullName: p.fullName,
          role: p.role,
          isProducer: p.isProducer
        }))
    });
  } catch (error) {
    logger.error(`Error joining room: ${error.message}`);
    res.status(500).json({ error: error.message });
  }
});

// Leave room
app.post('/rooms/:roomId/peers/:peerId/leave', async (req, res) => {
  try {
    const { roomId, peerId } = req.params;

    const room = rooms.get(roomId);
    if (!room) {
      return res.status(404).json({ error: 'Room not found' });
    }

    const peer = room.peers.get(peerId);
    if (!peer) {
      return res.status(404).json({ error: 'Peer not found' });
    }

    // Close all producers
    for (const producer of peer.producers.values()) {
      producer.close();
    }

    // Close all consumers
    for (const consumer of peer.consumers.values()) {
      consumer.close();
    }

    // Close transport
    if (peer.transport) {
      peer.transport.close();
    }

    // Remove peer
    room.peers.delete(peerId);

    logger.info(`Peer ${peerId} left room ${roomId}`);

    // Delete room if empty
    if (room.peers.size === 0) {
      rooms.delete(roomId);
      logger.info(`Room ${roomId} deleted (empty)`);
    }

    res.json({ success: true });
  } catch (error) {
    logger.error(`Error leaving room: ${error.message}`);
    res.status(500).json({ error: error.message });
  }
});

// Initialize Mediasoup and start server
async function start() {
  try {
    // Create Mediasoup worker
    mediasoupWorker = await mediasoup.createWorker({
      logLevel: process.env.MEDIASOUP_LOG_LEVEL || 'error',
      logTags: ['info', 'ice', 'dtls', 'rtp', 'srtp', 'rtcp', 'rtx', 'bwe', 'score', 'simulcast', 'svc'],
      rtcMinPort: parseInt(process.env.MEDIASOUP_RTC_MIN_PORT || 40000),
      rtcMaxPort: parseInt(process.env.MEDIASOUP_RTC_MAX_PORT || 49999)
    });

    logger.info('✓ Mediasoup worker created');

    // Handle worker close
    mediasoupWorker.on('died', () => {
      logger.error('❌ Mediasoup worker died, exiting');
      process.exit(1);
    });

    // Start Express server
    const PORT = parseInt(process.env.MEDIASOUP_PORT || 3000);
    app.listen(PORT, () => {
      logger.info(`✓ Mediasoup SFU server listening on port ${PORT}`);
      logger.info('═══════════════════════════════════════════════════════════');
      logger.info('  VTP Mediasoup SFU Service Started');
      logger.info('═══════════════════════════════════════════════════════════');
      logger.info(`  Endpoint: http://localhost:${PORT}`);
      logger.info(`  Health: http://localhost:${PORT}/health`);
      logger.info(`  Listen IP: ${process.env.MEDIASOUP_LISTEN_IP || '127.0.0.1'}`);
      logger.info(`  Announced IP: ${process.env.MEDIASOUP_ANNOUNCED_IP || '127.0.0.1'}`);
      logger.info(`  RTC Port Range: ${process.env.MEDIASOUP_RTC_MIN_PORT || 40000} - ${process.env.MEDIASOUP_RTC_MAX_PORT || 49999}`);
      logger.info('═══════════════════════════════════════════════════════════');
    });
  } catch (error) {
    logger.error(`Failed to start server: ${error.message}`);
    process.exit(1);
  }
}

start();
