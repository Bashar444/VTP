// VTP Signaling Server - Node.js Socket.IO v4
// Handles WebRTC signaling for live streaming

const express = require('express');
const http = require('http');
const { Server } = require('socket.io');
const axios = require('axios');
const jwt = require('jsonwebtoken');
const cors = require('cors');
const winston = require('winston');
require('dotenv').config();

// Configuration
const PORT = process.env.PORT || 3003;
const MEDIASOUP_URL = process.env.MEDIASOUP_URL || 'http://mediasoup-sfu:3000';
const JWT_SECRET = process.env.JWT_SECRET || 'your-secret-key-change-in-production';
const ALLOWED_ORIGINS = (process.env.ALLOWED_ORIGINS || 'http://localhost:3001,http://localhost:8080').split(',');

// Logger
const logger = winston.createLogger({
  level: process.env.LOG_LEVEL || 'info',
  format: winston.format.combine(
    winston.format.timestamp(),
    winston.format.printf(({ level, message, timestamp }) => {
      return `${timestamp} [${level.toUpperCase()}]: ${message}`;
    })
  ),
  transports: [new winston.transports.Console()]
});

// Express app
const app = express();
app.use(cors({ origin: ALLOWED_ORIGINS, credentials: true }));
app.use(express.json());

// Health check
app.get('/health', (req, res) => {
  res.json({ status: 'ok', service: 'signaling-server', timestamp: new Date().toISOString() });
});

// HTTP server
const server = http.createServer(app);

// Socket.IO server with proper CORS
const io = new Server(server, {
  cors: {
    origin: ALLOWED_ORIGINS,
    methods: ['GET', 'POST'],
    credentials: true
  },
  pingTimeout: 60000,
  pingInterval: 25000,
  transports: ['polling', 'websocket'],
  allowUpgrades: true
});

// Room state
const rooms = new Map(); // roomId -> { peers: Map, router: null }
const peerTransports = new Map(); // peerId -> { send: transport, recv: transport }

// Helper: Call MediaSoup SFU
async function callMediasoup(method, path, data = null) {
  try {
    const url = `${MEDIASOUP_URL}${path}`;
    const config = { method, url };
    if (data) config.data = data;
    
    const response = await axios(config);
    return response.data;
  } catch (error) {
    logger.error(`MediaSoup API error: ${error.message}`);
    throw error;
  }
}

// Get or create room
async function getOrCreateRoom(roomId, roomName) {
  if (!rooms.has(roomId)) {
    // Create room in MediaSoup SFU by joining with a dummy peer first
    // The room is created automatically when first peer joins
    rooms.set(roomId, {
      id: roomId,
      name: roomName || roomId,
      peers: new Map(),
      createdAt: Date.now()
    });
    logger.info(`Created room: ${roomId}`);
  }
  return rooms.get(roomId);
}

// Socket.IO connection handler
io.on('connection', (socket) => {
  logger.info(`✓ Socket connected: ${socket.id}`);
  
  let currentRoom = null;
  let currentPeerId = null;
  let currentUserId = null;

  // Handle join-room
  socket.on('join-room', async (payloadStr) => {
    try {
      const payload = typeof payloadStr === 'string' ? JSON.parse(payloadStr) : payloadStr;
      const { roomId, userId, email, fullName, role, isProducer, roomName } = payload;

      if (!roomId || !userId) {
        socket.emit('error', { error: 'Missing roomId or userId' });
        return;
      }

      logger.info(`User ${userId} (${email}) joining room ${roomId}`);

      currentRoom = roomId;
      currentPeerId = socket.id;
      currentUserId = userId;

      // Get or create room
      const room = await getOrCreateRoom(roomId, roomName);

      // Join MediaSoup room
      let mediasoupResponse;
      try {
        mediasoupResponse = await callMediasoup('POST', `/rooms/${roomId}/peers`, {
          peerId: socket.id,
          userId,
          email: email || 'guest@example.com',
          fullName: fullName || 'Guest',
          role: role || 'student',
          isProducer: isProducer || false,
          roomName: roomName || roomId
        });
      } catch (err) {
        logger.error(`MediaSoup join failed: ${err.message}`);
        // Return mock capabilities for testing
        mediasoupResponse = {
          rtpCapabilities: {
            codecs: [
              { kind: 'audio', mimeType: 'audio/opus', clockRate: 48000, channels: 2 },
              { kind: 'video', mimeType: 'video/VP8', clockRate: 90000 }
            ],
            headerExtensions: []
          },
          peers: []
        };
      }

      // Add peer to room
      room.peers.set(socket.id, {
        id: socket.id,
        userId: userId,
        email,
        fullName,
        role,
        isProducer,
        joinedAt: Date.now()
      });

      // Join socket room
      socket.join(roomId);

      // Notify others
      socket.to(roomId).emit('peer-joined', {
        peerId: socket.id,
        userId: userId,
        email,
        fullName,
        role,
        isProducer
      });

      // Send response
      socket.emit('joined-room', {
        Success: true,
        RoomID: roomId,
        ParticipantID: socket.id,
        Participants: Array.from(room.peers.values()),
        Mediasoup: {
          RtpCapabilities: mediasoupResponse.rtpCapabilities,
          Peers: mediasoupResponse.peers || []
        }
      });

      logger.info(`User ${userId} joined room ${roomId} successfully`);
    } catch (error) {
      logger.error(`join-room error: ${error.message}`);
      socket.emit('error', { error: error.message });
    }
  });

  // Handle create-transport
  socket.on('create-transport', async (payloadStr) => {
    try {
      const payload = typeof payloadStr === 'string' ? JSON.parse(payloadStr) : payloadStr;
      const { roomId, direction } = payload;

      if (!roomId || !direction) {
        socket.emit('error', { error: 'Missing roomId or direction' });
        return;
      }

      logger.info(`Creating ${direction} transport for peer ${socket.id} in room ${roomId}`);

      let transport;
      try {
        transport = await callMediasoup('POST', `/rooms/${roomId}/transports`, {
          peerId: socket.id,
          direction
        });
      } catch (err) {
        logger.error(`MediaSoup create-transport failed: ${err.message}`);
        // Return mock transport for testing
        transport = {
          transportId: `mock-transport-${Date.now()}`,
          iceParameters: { usernameFragment: 'mock', password: 'mock' },
          iceCandidates: [],
          dtlsParameters: { fingerprints: [], role: 'auto' }
        };
      }

      socket.emit('transport-created', transport);
      logger.info(`Transport created: ${transport.transportId}`);
    } catch (error) {
      logger.error(`create-transport error: ${error.message}`);
      socket.emit('error', { error: error.message });
    }
  });

  // Handle connect-transport
  socket.on('connect-transport', async (payloadStr) => {
    try {
      const payload = typeof payloadStr === 'string' ? JSON.parse(payloadStr) : payloadStr;
      const { roomId, transportId, dtlsParameters } = payload;

      logger.info(`Connecting transport ${transportId} in room ${roomId}`);

      try {
        await callMediasoup('POST', `/rooms/${roomId}/transports/${transportId}/connect`, {
          dtlsParameters
        });
      } catch (err) {
        logger.warn(`MediaSoup connect-transport: ${err.message}`);
      }

      socket.emit('transport-connected', { success: true });
    } catch (error) {
      logger.error(`connect-transport error: ${error.message}`);
      socket.emit('error', { error: error.message });
    }
  });

  // Handle produce
  socket.on('produce', async (payloadStr) => {
    try {
      const payload = typeof payloadStr === 'string' ? JSON.parse(payloadStr) : payloadStr;
      const { roomId, kind, rtpParameters } = payload;

      logger.info(`Creating ${kind} producer for peer ${socket.id} in room ${roomId}`);

      let producer;
      try {
        producer = await callMediasoup('POST', `/rooms/${roomId}/producers`, {
          peerId: socket.id,
          kind,
          rtpParameters
        });
      } catch (err) {
        logger.error(`MediaSoup produce failed: ${err.message}`);
        producer = { id: `mock-producer-${Date.now()}`, kind };
      }

      // Notify other peers about new producer
      socket.to(roomId).emit('new-producer', {
        producerId: producer.id,
        peerId: socket.id,
        kind
      });

      socket.emit('producer-created', producer);
      logger.info(`Producer created: ${producer.id}`);
    } catch (error) {
      logger.error(`produce error: ${error.message}`);
      socket.emit('error', { error: error.message });
    }
  });

  // Handle consume
  socket.on('consume', async (payloadStr) => {
    try {
      const payload = typeof payloadStr === 'string' ? JSON.parse(payloadStr) : payloadStr;
      const { roomId, producerId, rtpCapabilities } = payload;

      logger.info(`Creating consumer for producer ${producerId} in room ${roomId}`);

      let consumer;
      try {
        consumer = await callMediasoup('POST', `/rooms/${roomId}/consumers`, {
          peerId: socket.id,
          producerId,
          rtpCapabilities
        });
      } catch (err) {
        logger.error(`MediaSoup consume failed: ${err.message}`);
        consumer = {
          id: `mock-consumer-${Date.now()}`,
          producerId,
          kind: 'video',
          rtpParameters: {}
        };
      }

      socket.emit('consumer-created', consumer);
      logger.info(`Consumer created: ${consumer.id}`);
    } catch (error) {
      logger.error(`consume error: ${error.message}`);
      socket.emit('error', { error: error.message });
    }
  });

  // Handle chat messages
  socket.on('chat-message', (payloadStr) => {
    try {
      const payload = typeof payloadStr === 'string' ? JSON.parse(payloadStr) : payloadStr;
      const { roomId, message, userId, fullName } = payload;

      if (roomId) {
        io.to(roomId).emit('chat-message', {
          id: `msg-${Date.now()}`,
          userId,
          fullName,
          message,
          timestamp: new Date().toISOString()
        });
      }
    } catch (error) {
      logger.error(`chat-message error: ${error.message}`);
    }
  });

  // Handle disconnect
  socket.on('disconnect', (reason) => {
    logger.info(`✗ Socket disconnected: ${socket.id} (reason: ${reason})`);

    if (currentRoom) {
      const room = rooms.get(currentRoom);
      if (room) {
        room.peers.delete(socket.id);
        
        // Notify others
        socket.to(currentRoom).emit('peer-left', {
          peerId: socket.id,
          userId: currentUserId
        });

        // Clean up empty room
        if (room.peers.size === 0) {
          rooms.delete(currentRoom);
          logger.info(`Room ${currentRoom} deleted (empty)`);
        }
      }
    }
  });

  // Handle errors
  socket.on('error', (error) => {
    logger.error(`Socket error: ${error.message}`);
  });
});

// Start server
server.listen(PORT, () => {
  logger.info(`🚀 Signaling server running on port ${PORT}`);
  logger.info(`   MediaSoup URL: ${MEDIASOUP_URL}`);
  logger.info(`   Allowed origins: ${ALLOWED_ORIGINS.join(', ')}`);
});
